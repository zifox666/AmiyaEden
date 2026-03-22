package service

import "testing"

func TestShouldRetryWithoutProxy(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil", err: nil, want: false},
		{name: "proxyconnect", err: testErr("proxyconnect tcp: dial tcp 127.0.0.1:7890: connect: connection refused"), want: true},
		{name: "socks", err: testErr("socks connect failed"), want: true},
		{name: "other", err: testErr("tls handshake timeout"), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldRetryWithoutProxy(tt.err); got != tt.want {
				t.Fatalf("shouldRetryWithoutProxy() = %v, want %v", got, tt.want)
			}
		})
	}
}

type testErr string

func (e testErr) Error() string { return string(e) }
