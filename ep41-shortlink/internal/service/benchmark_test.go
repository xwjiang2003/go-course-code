package service

import (
	"testing"

	"shortlink/internal/cache"
	"shortlink/internal/store"
)

func BenchmarkShorten(b *testing.B) {
	svc := New(store.NewMemory(), cache.New())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.Shorten("http://example.com/benchmark")
	}
}

func BenchmarkResolve(b *testing.B) {
	svc := New(store.NewMemory(), cache.New())
	code, _ := svc.Shorten("http://example.com/resolve")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.Resolve(code)
	}
}
