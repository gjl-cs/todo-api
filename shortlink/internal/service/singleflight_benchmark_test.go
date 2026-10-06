package service

import (
	"testing"

	"golang.org/x/sync/singleflight"
)

// BenchmarkSingleflight 测量 singleflight 的基本调用开销。
func BenchmarkSingleflight(b *testing.B) {
	var group singleflight.Group

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err, _ := group.Do("abc123", func() (interface{}, error) {
			return "https://example.com/article", nil
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkSingleflightParallel 测量多个 goroutine 并发调用 singleflight 的开销。
func BenchmarkSingleflightParallel(b *testing.B) {
	var group singleflight.Group

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err, _ := group.Do("abc123", func() (interface{}, error) {
				return "https://example.com/article", nil
			})
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
