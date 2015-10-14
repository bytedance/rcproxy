package proxy

import (
	"crypto/rand"
	"testing"
)

func benchmarkCRC16ForNBytes(b *testing.B, numBytes int) {
	buf := make([]byte, numBytes)
	_, err := rand.Read(buf)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CRC16(buf)
	}
}

func BenchmarkCRC16For16Bytes(b *testing.B) {
	benchmarkCRC16ForNBytes(b, 16)
}

func BenchmarkCRC16For32Bytes(b *testing.B) {
	benchmarkCRC16ForNBytes(b, 32)
}

func BenchmarkCRC16For64Bytes(b *testing.B) {
	benchmarkCRC16ForNBytes(b, 64)
}

func BenchmarkCRC16For128Bytes(b *testing.B) {
	benchmarkCRC16ForNBytes(b, 128)
}
