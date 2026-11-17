package base62

const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// Encode 把非负整数转成 Base62 字符串。
func Encode(n int64) string {
	if n == 0 {
		return string(alphabet[0])
	}
	var buf []byte
	for n > 0 {
		buf = append(buf, alphabet[n%62])
		n /= 62
	}
	for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}
	return string(buf)
}

// Decode 把 Base62 字符串还原成整数。
func Decode(s string) int64 {
	var n int64
	for i := 0; i < len(s); i++ {
		n = n*62 + indexOf(s[i])
	}
	return n
}

func indexOf(c byte) int64 {
	switch {
	case c >= '0' && c <= '9':
		return int64(c - '0')
	case c >= 'a' && c <= 'z':
		return int64(c-'a') + 10
	case c >= 'A' && c <= 'Z':
		return int64(c-'A') + 36
	}
	return -1
}
