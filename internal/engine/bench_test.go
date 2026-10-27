package engine

import (
	"path/filepath"
	"testing"

	"banner-fingerprint/internal/domain/fingerprint"
)

func BenchmarkFingerprint(b *testing.B) {
	eng, err := Load(filepath.Join("..", "..", "rules"))
	if err != nil {
		b.Fatal(err)
	}
	banners := [][]byte{
		[]byte("SSH-2.0-OpenSSH_8.9p1 Ubuntu-3"),
		[]byte("HTTP/1.1 200 OK\r\nServer: nginx/1.24.0\r\nContent-Type: text/html"),
		[]byte("J\x00\x00\x00\n8.0.32\x00"),
		[]byte("-ERR wrong number of arguments for 'get' command"),
		[]byte("220 ProFTPD 1.3.7 Server (ProFTPD)"),
		[]byte("\x16\x03\x01\x00\xa5"),
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = eng.Fingerprint(fingerprint.Input{
			IP:     "1.2.3.4",
			Port:   80,
			Banner: banners[i%len(banners)],
		})
	}
}
