package shortener

const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func Base62Encode(num uint64) string {
    encoded := ""
    for num > 0 {
        remainder := num % 62
        encoded = string(base62Chars[remainder]) + encoded
        num = num / 62
    }
    return encoded
}

func GenerateShortener() string {
	return ""
}