package plugin

import (
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"testing"
)

func TestIterIndex(t *testing.T) {
	p := fuzzTypes.Plugin{
		Name: "milaogiu_iterator",
		Args: []any{SelectIterIndex, 8, 9},
	}
	lengths := []int{22, 6, 3, 5}
	out := make([]int, len(lengths))
	IterIndex(p, lengths, out)
	fmt.Println(out)
}

func TestDoRequest(t *testing.T) {

}
