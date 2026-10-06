package utils

import "testing"

func TestGenerateShortCode(t *testing.T) {
	code, err := GenerateShortCode(6)
	if err != nil {
		t.Fatalf("生成短码失败: %v", err)
	}

	if len(code) != 6 {
		t.Fatalf("短码长度错误: want 6, got %d", len(code))
	}

	for _, ch := range code {
		if !containsRune(charset, ch) {
			t.Fatalf("短码包含非法字符: %q", ch)
		}
	}
}

func TestGenerateShortCodeInvalidLength(t *testing.T) {
	_, err := GenerateShortCode(0)
	if err == nil {
		t.Fatal("长度为 0 时应该返回错误")
	}

	_, err = GenerateShortCode(-1)
	if err == nil {
		t.Fatal("长度为负数时应该返回错误")
	}
}

func TestGenerateShortCodeUniqueness(t *testing.T) {
	const count = 1000

	seen := make(map[string]struct{}, count)

	for i := 0; i < count; i++ {
		code, err := GenerateShortCode(6)
		if err != nil {
			t.Fatalf("第 %d 次生成短码失败: %v", i+1, err)
		}

		if _, exists := seen[code]; exists {
			t.Fatalf("发现重复短码: %s", code)
		}

		seen[code] = struct{}{}
	}
}

func containsRune(s string, target rune) bool {
	for _, ch := range s {
		if ch == target {
			return true
		}
	}
	return false
}
