package fuzz

import (
	"fmt"
	"testing"
)

func TestIterIndexClusterbomb(t *testing.T) {
	var (
		lengths = []int{3, 4, 5, 2}
		out     = make([]int, 4)
	)
	for i := 0; i < 3*4*5*2; i++ {
		iterIndexClusterbomb(lengths, i, out)
		fmt.Println(out)
	}
}
