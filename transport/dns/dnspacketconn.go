package dns

import (
	"bytes"
	"crypto/rand"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"

	dnstt "github.com/QcomWrt/Q-SSH-WORKER/transport/dns/internal/dns"
	turbotunnel "github.com/QcomWrt/Q-SSH-WORKER/transport/dns/internal/turbotunnel"

	"github.com/QcomWrt/Q-SSH-WORKER/debug"
)

const (
	// Jumlah byte padding acak di setiap query. Harus sama dengan dnstt-client.
	numPadding = 3
	// Padding tambahan untuk polling kosong, supaya tidak kena cache resolver.
	numPaddingForPoll = 8

	initPollDelay       = 500 * time.Millisecond
	maxPollDelay        = 10 * time.Second
	pollDelayMultiplier = 2.0

	// Batas poll kosong yang boleh dikirim sebagai respons atas data masuk.
	pollLimit = 16
)

// base32Encoding adalah base32 tanpa padding.
var base32Encoding = base32.StdEncoding.WithPadding(base32.NoPadding)

// DNSPacketConn menyediakan antarmuka kirim dan terima paket di atas DNS.
// Paket dikodekan ke nama DNS pada query (upstream), dan ke TXT pada balasan
// (downstream). Pengiriman UDP sebenarnya dilakukan oleh transport.
type DNSPacketConn struct {
	clientID turbotunnel.ClientID
	domain   dnstt.Name
	pollChan chan struct{}
	*turbotunnel.QueuePacketConn
}

// NewDNSPacketConn membuat DNSPacketConn. transport adalah socket UDP tempat
// query dikirim dan balasan diterima. addr adalah alamat resolver.
func NewDNSPacketConn(transport net.PacketConn, addr net.Addr, domain dnstt.Name) *DNSPacketConn {
	clientID := turbotunnel.NewClientID()
	c := &DNSPacketConn{
		clientID:        clientID,
		domain:          domain,
		pollChan:        make(chan struct{}, pollLimit),
		QueuePacketConn: turbotunnel.NewQueuePacketConn(clientID, 0),
	}
	go func() {
		if err := c.recvLoop(transport); err != nil {
			debug.Printf("[SLOWDNS DEBUG] recvLoop berhenti: %v", err)
		}
	}()
	go func() {
		if err := c.sendLoop(transport, addr); err != nil {
			debug.Printf("[SLOWDNS DEBUG] sendLoop berhenti: %v", err)
		}
	}()
	return c
}

// dnsResponsePayload mengambil payload downstream dari balasan DNS.
// Mengembalikan nil jika format tidak sesuai.
func dnsResponsePayload(resp *dnstt.Message, domain dnstt.Name) []byte {
	if resp.Flags&0x8000 != 0x8000 {
		return nil
	}
	if resp.Flags&0x000f != dnstt.RcodeNoError {
		return nil
	}
	if len(resp.Answer) != 1 {
		return nil
	}
	answer := resp.Answer[0]

	if _, ok := answer.Name.TrimSuffix(domain); !ok {
		return nil
	}
	if answer.Type != dnstt.RRTypeTXT {
		return nil
	}
	payload, err := dnstt.DecodeRDataTXT(answer.Data)
	if err != nil {
		return nil
	}
	return payload
}

// nextPacket membaca satu paket dengan prefix panjang 2 byte.
func nextPacket(r *bytes.Reader) ([]byte, error) {
	var n uint16
	if err := binary.Read(r, binary.BigEndian, &n); err != nil {
		return nil, err
	}
	p := make([]byte, n)
	_, err := io.ReadFull(r, p)
	if err == io.EOF {
		err = io.ErrUnexpectedEOF
	}
	return p, err
}

// recvLoop menerima balasan DNS, memecahnya menjadi paket, dan memasukkannya
// ke antrean masuk. Jika ada data, polling segera diizinkan.
func (c *DNSPacketConn) recvLoop(transport net.PacketConn) error {
	for {
		var buf [4096]byte
		n, addr, err := transport.ReadFrom(buf[:])
		if err != nil {
			return err
		}

		resp, err := dnstt.MessageFromWireFormat(buf[:n])
		if err != nil {
			debug.Printf("[SLOWDNS DEBUG] MessageFromWireFormat: %v", err)
			continue
		}

		payload := dnsResponsePayload(&resp, c.domain)

		r := bytes.NewReader(payload)
		found := false
		for {
			p, err := nextPacket(r)
			if err != nil {
				break
			}
			found = true
			c.QueuePacketConn.QueueIncoming(p, addr)
		}

		if found {
			select {
			case c.pollChan <- struct{}{}:
			default:
			}
		}
	}
}

// chunks memecah p menjadi potongan maksimal n byte.
func chunks(p []byte, n int) [][]byte {
	var result [][]byte
	for len(p) > 0 {
		sz := len(p)
		if sz > n {
			sz = n
		}
		result = append(result, p[:sz])
		p = p[sz:]
	}
	return result
}

// send mengirim p sebagai satu query DNS. Panjang p harus kurang dari 224 byte.
// Format: ClientID | prefix padding | [panjang | data], lalu base32 lowercase,
// dipecah per 63 karakter, dan ditambah domain.
func (c *DNSPacketConn) send(transport net.PacketConn, p []byte, addr net.Addr) error {
	if len(p) >= 224 {
		return fmt.Errorf("paket terlalu panjang: %d", len(p))
	}

	var decoded []byte
	{
		var buf bytes.Buffer
		buf.Write(c.clientID[:])
		n := numPadding
		if len(p) == 0 {
			n = numPaddingForPoll
		}
		buf.WriteByte(byte(224 + n))
		io.CopyN(&buf, rand.Reader, int64(n))
		if len(p) > 0 {
			buf.WriteByte(byte(len(p)))
			buf.Write(p)
		}
		decoded = buf.Bytes()
	}

	encoded := make([]byte, base32Encoding.EncodedLen(len(decoded)))
	base32Encoding.Encode(encoded, decoded)
	encoded = bytes.ToLower(encoded)

	labels := chunks(encoded, 63)
	labels = append(labels, c.domain...)
	name, err := dnstt.NewName(labels)
	if err != nil {
		return err
	}

	var id uint16
	binary.Read(rand.Reader, binary.BigEndian, &id)
	query := &dnstt.Message{
		ID:    id,
		Flags: 0x0100, // QR = 0, RD = 1
		Question: []dnstt.Question{
			{
				Name:  name,
				Type:  dnstt.RRTypeTXT,
				Class: dnstt.ClassIN,
			},
		},
		// EDNS(0)
		Additional: []dnstt.RR{
			{
				Name:  dnstt.Name{},
				Type:  dnstt.RRTypeOPT,
				Class: 4096,
				TTL:   0,
				Data:  []byte{},
			},
		},
	}
	buf, err := query.WireFormat()
	if err != nil {
		return err
	}

	_, err = transport.WriteTo(buf, addr)
	return err
}

// sendLoop mengambil paket keluar dan mengirimnya. Jika tidak ada data,
// mengirim poll kosong sesuai jadwal atau saat diminta pollChan.
func (c *DNSPacketConn) sendLoop(transport net.PacketConn, addr net.Addr) error {
	pollDelay := initPollDelay
	pollTimer := time.NewTimer(pollDelay)
	for {
		var p []byte
		outgoing := c.QueuePacketConn.OutgoingQueue(addr)
		pollTimerExpired := false

		select {
		case p = <-outgoing:
		default:
			select {
			case p = <-outgoing:
			case <-c.pollChan:
			case <-pollTimer.C:
				pollTimerExpired = true
			}
		}

		if len(p) > 0 {
			select {
			case <-c.pollChan:
			default:
			}
		}

		if pollTimerExpired {
			pollDelay = time.Duration(float64(pollDelay) * pollDelayMultiplier)
			if pollDelay > maxPollDelay {
				pollDelay = maxPollDelay
			}
		} else {
			if !pollTimer.Stop() {
				<-pollTimer.C
			}
			pollDelay = initPollDelay
		}
		pollTimer.Reset(pollDelay)

		if err := c.send(transport, p, addr); err != nil {
			debug.Printf("[SLOWDNS DEBUG] send: %v", err)
		}
	}
}
