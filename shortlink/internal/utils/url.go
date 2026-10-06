package utils

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

// ErrInvalidLongURL 表示原始 URL 不符合要求。
var ErrInvalidLongURL = errors.New("invalid long URL")

// ValidateLongURL 校验短链接的原始 URL。
func ValidateLongURL(rawURL string) error {
	rawURL = strings.TrimSpace(rawURL)

	if rawURL == "" {
		return fmt.Errorf("%w: URL cannot be empty", ErrInvalidLongURL)
	}

	// 拒绝空格、换行和控制字符。
	if strings.IndexFunc(rawURL, func(r rune) bool {
		return r <= ' ' || r == 127
	}) >= 0 {
		return fmt.Errorf(
			"%w: URL contains whitespace or control characters",
			ErrInvalidLongURL,
		)
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidLongURL, err)
	}

	// 只允许 HTTP 和 HTTPS。
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf(
			"%w: only http and https URLs are allowed",
			ErrInvalidLongURL,
		)
	}

	// 必须包含主机名。
	if parsedURL.Host == "" || parsedURL.Hostname() == "" {
		return fmt.Errorf(
			"%w: URL must contain a hostname",
			ErrInvalidLongURL,
		)
	}

	// 检查端口是否合法。
	port := parsedURL.Port()
	if port != "" {
		number, err := strconv.Atoi(port)
		if err != nil || number < 1 || number > 65535 {
			return fmt.Errorf("%w: invalid URL port", ErrInvalidLongURL)
		}
	}

	hostname := parsedURL.Hostname()

	// IP 地址可以直接通过；其他主机名不能包含明显非法字符。
	if net.ParseIP(hostname) == nil {
		if strings.ContainsAny(hostname, "/?#@") {
			return fmt.Errorf("%w: invalid hostname", ErrInvalidLongURL)
		}
	}

	return nil
}
