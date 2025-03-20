package service

import (
	"fmt"
	"sync"
	"testing"
)

func TestNewsnowflake(t *testing.T) {
	sf, _ := newSnowflake(1, 1)
	t.Log("TestNewSnowflake with goroutine")
	var wg sync.WaitGroup
	for i := range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fmt.Printf("test %d ", i)
			id, _ := sf.Generate()
			if id < 0 {
				t.Errorf("Generate error: %v", id)
			}
			fmt.Printf("Snowflake ID: %d\t", id)
			shortcode := encodeSnowflake(id)
			fmt.Printf("Short Code: %s\n", shortcode)
		}()
	}
	wg.Wait()
}
