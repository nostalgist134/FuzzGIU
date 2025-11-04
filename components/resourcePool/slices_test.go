package resourcePool

import (
	"fmt"
	"testing"
)

func TestSlices(t *testing.T) {
	sp := NewSlicePool[string](30)
	nishigiu := sp.Get(20)
	fmt.Println(nishigiu)
}
