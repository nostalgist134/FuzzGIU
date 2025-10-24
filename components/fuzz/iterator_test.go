package fuzz

import (
	"fmt"
	"testing"
)

func TestIterIndexClusterbomb(t *testing.T) {
	var (
		lengths = []int{2, 71, 399, 9812}
		out     = make([]int, len(lengths))
	)
	fmt.Println(iterLenClusterbomb(lengths))
	for i := 0; i < iterLenClusterbomb(lengths); i++ {
		iterIndexClusterbomb(lengths, i, out)
	}
	fmt.Println("done")
}
