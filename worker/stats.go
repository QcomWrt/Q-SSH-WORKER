package worker

import (
	"net"
	"sync/atomic"
)

// TrafficStats menyimpan jumlah byte masuk dan keluar menggunakan operasi atomik
type TrafficStats struct {
	RxBytes uint64 // Receiver / Download
	TxBytes uint64 // Transmitter / Upload
}

// ObservedConn membungkus net.Conn standar untuk menghitung setiap byte yang lewat
type ObservedConn struct {
	net.Conn
	stats *TrafficStats
}

// NewObservedConn menginisialisasi pembungkus koneksi untuk tracking statistik
func NewObservedConn(conn net.Conn, stats *TrafficStats) net.Conn {
	return &ObservedConn{
		Conn:  conn,
		stats: stats,
	}
}

// Read mengintersep data masuk (Download / Rx)
func (o *ObservedConn) Read(b []byte) (n int, err error) {
	n, err = o.Conn.Read(b)
	if n > 0 && o.stats != nil {
		atomic.AddUint64(&o.stats.RxBytes, uint64(n))
	}
	return n, err
}

// Write mengintersep data keluar (Upload / Tx)
func (o *ObservedConn) Write(b []byte) (n int, err error) {
	n, err = o.Conn.Write(b)
	if n > 0 && o.stats != nil {
		atomic.AddUint64(&o.stats.TxBytes, uint64(n))
	}
	return n, err
}

// GetStats mengembalikan jumlah Rx dan Tx saat ini
func (o *ObservedConn) GetStats() (uint64, uint64) {
	if o.stats == nil {
		return 0, 0
	}
	return atomic.LoadUint64(&o.stats.RxBytes), atomic.LoadUint64(&o.stats.TxBytes)
}