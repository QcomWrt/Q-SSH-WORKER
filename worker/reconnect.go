package worker

import (
	"math/rand"
	"time"
)

// ReconnectPolicy menyimpan status durasi backoff untuk satu siklus kegagalan
type ReconnectPolicy struct {
	baseDelay float64       // Jeda awal (misal 2 detik)
	maxDelay  time.Duration // Jeda maksimal (misal 30 detik)
	attempts  int           // Jumlah kegagalan berturut-turut
}

// NewReconnectPolicy menginisialisasi kebijakan koneksi ulang
func NewReconnectPolicy(base time.Duration, max time.Duration) *ReconnectPolicy {
	return &ReconnectPolicy{
		baseDelay: base.Seconds(),
		maxDelay:  max,
		attempts:  0,
	}
}

// GetDelay menghitung durasi tunggu berikutnya menggunakan Exponential Backoff + Jitter
func (p *ReconnectPolicy) GetDelay() time.Duration {
	p.attempts++

	// Rumus eksponensial: base * 2^(attempts-1)
	temp := p.baseDelay * float64(uint(1)<<uint(p.attempts-1))
	
	// Berikan batas maksimal agar tidak menunggu selamanya
	if temp > p.maxDelay.Seconds() {
		temp = p.maxDelay.Seconds()
	}

	// Tambahkan Jitter (Acak antara 50% hingga 100% dari nilai temp)
	// Ini mencegah fenomena "Thundering Herd" pada multi-worker
	jitter := 0.5 + rand.Float64()*0.5
	finalDelay := temp * jitter

	return time.Duration(finalDelay * float64(time.Second))
}

// Reset mengembalikan hitungan kegagalan ke nol saat koneksi resmi sukses terhubung
func (p *ReconnectPolicy) Reset() {
	p.attempts = 0
}