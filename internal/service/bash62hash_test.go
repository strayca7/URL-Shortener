package service

import (
	"fmt"
	"testing"
)

func TestBase62Hash(t *testing.T) {
	tests := []struct {
		name string
		num  int64
	}{
		{
			name: "zero",
			num:  0,
		},
		{
			name: "positive",
			num:  123456,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Base62Encode(tt.num)
			fmt.Println("got:", got)
		})
	}
}

func TestSnowflake(t *testing.T) {
	sf, err := NewSnowflake(1, 1)
	if err != nil {
		panic(err)
	}

	// 生成唯一 ID
	id, _ := sf.Generate()

	// 转换为 Base62 短码
	shortCode := Base62Encode(id)

	fmt.Printf("Snowflake ID: %d\n", id)
	fmt.Printf("Short Code: %s\n", shortCode) // 例如: "abc123"
	fmt.Println(sf)
}
