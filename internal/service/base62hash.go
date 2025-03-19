package service

import (
	"math/rand"
)

const (
	base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

var shuffledChars = shuffleBase62(base62Chars)

// Base62Encode，将 Snowflake ID 转换为 Base62 短码, 该短码会按顺序生成。
//
// Base62Encode converts a Snowflake ID to a base62 string, which is generated sequentially.
func Base62Encode(num int64) string {
	if num == 0 {
		return string(base62Chars[0])
	}

	var encoded []byte
	for num > 0 {
		remainder := num % 62
		num = num / 62
		encoded = append([]byte{base62Chars[remainder]}, encoded...)
	}

	return string(encoded)
}

// shuffleBase62 用于随机打乱 base62 字符串。
//
// shuffleBase62 shuffles a base62 string randomly.
func shuffleBase62(s string) string {
	chars := []rune(s)
	rand.Shuffle(len(chars), func(i, j int) { chars[i], chars[j] = chars[j], chars[i] })
	return string(chars)
}

// bitConfusion 将高32位与低32位交换，并异或掩码，进行位混淆。
//
// bitConfusion swaps the high 32 bits with the low 32 bits and XORs with a mask.
func bitConfusion(id int64) uint64 {
	high := uint64(id >> 32)
	low := uint64(id & 0xFFFFFFFF)
	masked := (high ^ 0x55555555) | (low << 32) // 掩码异或 + 高低位交换
	return masked & 0x3FFFFFFFFFFFFF            // 保留50位（2^50 ≈ 1e15）
}

// EncodeSnowflake 将 50 位混淆值编码为 10 位 Base62。
//
// EncodeSnowflake encodes a 50-bit confusion value to a 10-bit Base62 string.
func EncodeSnowflake(id int64) string {
	masked := bitConfusion(id)
	const base = 62
	const length = 10
	var encoded [length]byte

	for i := length - 1; i >= 0 && masked > 0; i-- {
		remainder := masked % base
		masked /= base
		encoded[i] = shuffledChars[remainder]
	}

	// 不足10位时填充随机字符
	for i := 0; masked == 0 && i < length; i++ {
		if encoded[i] == 0 {
			encoded[i] = shuffledChars[rand.Intn(base)]
		}
	}
	return string(encoded[:])
}

// CreateShortURL 集成 Snowflake 和 Base62 编码，生成短链。
//
// CreateShortURL integrates Snowflake and Base62 encoding to generate a short URL.
func CreateShortURL() (string, error) {
	sf, err := NewSnowflake(0, 0)
	if err != nil {
		return "", err
	}
	id, err := sf.Generate()
	if err != nil {
		return "", err
	}
	return EncodeSnowflake(id), nil
}
