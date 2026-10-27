package engine

import (
	"math"
	"path/filepath"
	"testing"

	"banner-fingerprint/internal/domain/fingerprint"
)

func TestFingerprintFixtures(t *testing.T) {
	eng, err := Load(filepath.Join("..", "..", "rules"))
	if err != nil {
		t.Fatalf("load rules: %v", err)
	}
	if eng.RuleCount() == 0 {
		t.Fatal("expected at least one rule")
	}
	if len(eng.Digest()) != 64 {
		t.Fatalf("unexpected digest: %q", eng.Digest())
	}

	cases := []struct {
		name string
		port int
		in   string
		want fingerprint.Result
	}{
		{
			name: "openssh ubuntu",
			port: 22,
			in:   "SSH-2.0-OpenSSH_8.9p1 Ubuntu-3",
			want: fingerprint.Result{Protocol: "SSH", Product: "OpenSSH", Version: "8.9p1", OSHint: "Ubuntu", Confidence: 0.95},
		},
		{
			name: "nginx",
			port: 80,
			in:   "HTTP/1.1 200 OK\r\nServer: nginx/1.24.0\r\nContent-Type: text/html",
			want: fingerprint.Result{Protocol: "HTTP", Product: "nginx", Version: "1.24.0", Confidence: 0.9},
		},
		{
			name: "apache",
			port: 443,
			in:   "HTTP/1.1 200 OK\r\nServer: Apache/2.4.57",
			want: fingerprint.Result{Protocol: "HTTP", Product: "Apache", Version: "2.4.57", Confidence: 0.9},
		},
		{
			name: "mysql 8",
			port: 3306,
			in:   "J\x00\x00\x00\n8.0.32\x00",
			want: fingerprint.Result{Protocol: "MySQL", Product: "MySQL", Version: "8.0.32", Confidence: 0.9},
		},
		{
			name: "redis err",
			port: 6379,
			in:   "-ERR wrong number of arguments for 'get' command",
			want: fingerprint.Result{Protocol: "Redis", Product: "Redis", Confidence: 0.7},
		},
		{
			name: "proftpd",
			port: 21,
			in:   "220 ProFTPD 1.3.7 Server (ProFTPD)",
			want: fingerprint.Result{Protocol: "FTP", Product: "ProFTPD", Version: "1.3.7", Confidence: 0.9},
		},
		{
			name: "jetty slash",
			port: 8080,
			in:   "HTTP/1.1 404 Not Found\r\nServer: Jetty/9.4.51",
			want: fingerprint.Result{Protocol: "HTTP", Product: "Jetty", Version: "9.4.51", Confidence: 0.85},
		},
		{
			name: "openssh debian",
			port: 22,
			in:   "SSH-2.0-OpenSSH_9.3 Debian-1",
			want: fingerprint.Result{Protocol: "SSH", Product: "OpenSSH", Version: "9.3", OSHint: "Debian", Confidence: 0.95},
		},
		{
			name: "nginx ubuntu",
			port: 80,
			in:   "HTTP/1.1 200 OK\r\nServer: nginx/1.18.0 (Ubuntu)",
			want: fingerprint.Result{Protocol: "HTTP", Product: "nginx", Version: "1.18.0", OSHint: "Ubuntu", Confidence: 0.9},
		},
		{
			name: "apache ubuntu",
			port: 443,
			in:   "HTTP/1.1 200 OK\r\nServer: Apache/2.4.41 (Ubuntu)",
			want: fingerprint.Result{Protocol: "HTTP", Product: "Apache", Version: "2.4.41", OSHint: "Ubuntu", Confidence: 0.9},
		},
		{
			name: "mysql 5",
			port: 3306,
			in:   "J\x00\x00\x00\n5.7.42\x00",
			want: fingerprint.Result{Protocol: "MySQL", Product: "MySQL", Version: "5.7.42", Confidence: 0.9},
		},
		{
			name: "redis pong",
			port: 6379,
			in:   "+PONG",
			want: fingerprint.Result{Protocol: "Redis", Product: "Redis", Confidence: 0.7},
		},
		{
			name: "vsftpd",
			port: 21,
			in:   "220 (vsFTPd 3.0.5)",
			want: fingerprint.Result{Protocol: "FTP", Product: "vsFTPd", Version: "3.0.5", Confidence: 0.85},
		},
		{
			name: "nginx 8443",
			port: 8443,
			in:   "HTTP/1.1 200 OK\r\nServer: nginx/1.25.3",
			want: fingerprint.Result{Protocol: "HTTP", Product: "nginx", Version: "1.25.3", Confidence: 0.9},
		},
		{
			name: "openssh old",
			port: 22,
			in:   "SSH-1.99-OpenSSH_4.3",
			want: fingerprint.Result{Protocol: "SSH", Product: "OpenSSH", Version: "4.3", Confidence: 0.95},
		},
		{
			name: "iis",
			port: 8888,
			in:   "HTTP/1.1 200 OK\r\nServer: Microsoft-IIS/10.0",
			want: fingerprint.Result{Protocol: "HTTP", Product: "Microsoft-IIS", Version: "10.0", OSHint: "Windows", Confidence: 0.85},
		},
		{
			name: "redis noauth",
			port: 6379,
			in:   "-NOAUTH Authentication required.",
			want: fingerprint.Result{Protocol: "Redis", Product: "Redis", Confidence: 0.7},
		},
		{
			name: "pure-ftpd",
			port: 21,
			in:   "220 Welcome to Pure-FTPd",
			want: fingerprint.Result{Protocol: "FTP", Product: "Pure-FTPd", Confidence: 0.85},
		},
		{
			name: "tls",
			port: 9999,
			in:   "\x16\x03\x01\x00\xa5\x01\x00\x00\xa1",
			want: fingerprint.Result{Protocol: "unknown"},
		},
		{
			name: "quit",
			port: 12345,
			in:   "QUIT\r\n",
			want: fingerprint.Result{Protocol: "unknown"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := eng.Fingerprint(fingerprint.Input{IP: "1.2.3.4", Port: tc.port, Banner: []byte(tc.in)})
			if err != nil {
				t.Fatalf("fingerprint: %v", err)
			}
			if got.Protocol != tc.want.Protocol || got.Product != tc.want.Product || got.Version != tc.want.Version || got.OSHint != tc.want.OSHint {
				t.Fatalf("got %+v, want %+v", got, tc.want)
			}
			if math.Abs(got.Confidence-tc.want.Confidence) > 1e-9 {
				t.Fatalf("confidence got %v, want %v", got.Confidence, tc.want.Confidence)
			}
			if got.Protocol == "unknown" && got.Confidence != 0 {
				t.Fatalf("unknown must have zero confidence, got %v", got.Confidence)
			}
		})
	}
}
