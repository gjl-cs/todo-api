package utils

import "testing"

// BenchmarkGenerateShortCode 测试短码生成性能。
func BenchmarkGenerateShortCode(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		GenerateShortCode(6)
	}
}
