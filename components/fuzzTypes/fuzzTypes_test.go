package fuzzTypes

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestPluginMarshal(t *testing.T) {
	p := Plugin{
		Name: "NISHIGIU",
		Args: []any{"NSGIU", 1, 9.0, false, "WOSHIGIU", -3},
	}
	b, _ := json.Marshal(p)
	fmt.Println(string(b))
	q := Plugin{}
	json.Unmarshal(b, &q)
	fmt.Println(q)
	type wrapped struct {
		A int
		B float64
		C Plugin
	}
	r := wrapped{
		A: 3,
		B: 4,
		C: Plugin{"NAME", []any{"woshigiu", 3, 9.7, false}},
	}
	b, _ = json.Marshal(r)
	fmt.Println(string(b))
	r2 := wrapped{}
	json.Unmarshal(b, &r2)
	fmt.Println(r2)
	for _, a := range r2.C.Args {
		fmt.Printf("%T\n", a)
	}
}
