package rp

import (
	"context"
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/fuzz/fuzzCtx"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"sync/atomic"
	"testing"
	"time"
)

func testExec(*fuzzCtx.TaskCtx) *fuzzTypes.Reaction {
	//fmt.Println("nishigiuwoshigiumilaogiu")
	return &fuzzTypes.Reaction{
		Flag: 3,
	}
}

func test2Exec(*fuzzCtx.TaskCtx) *fuzzTypes.Reaction {
	//fmt.Println("nishigiuwoshigiumilaogiu2")
	return &fuzzTypes.Reaction{
		Flag: 4,
	}
}

func BenchmarkRp(b *testing.B) {
	b.ReportAllocs()
	for j := 0; j < b.N; j++ {
		p := NewRp(10)
		p.RegisterExecutor(testExec, ExecMajor)
		p.RegisterExecutor(test2Exec, ExecMinor)
		p.Start()
		ctx, cancel := context.WithCancel(context.Background())
		total := atomic.Int64{}
		//wg := sync.WaitGroup{}
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				default:
					for r := p.GetSingleResult(); r != nil; r = p.GetSingleResult() {
						total.Add(1)
					}
				}
			}
		}()
		for i := 0; i < 10000; i++ {
			for !p.Submit(nil, ExecMajor, 50*time.Millisecond) {
			}
			for !p.Submit(nil, ExecMinor, 50*time.Millisecond) {
			}
		}
		for r := p.GetSingleResult(); r != nil; r = p.GetSingleResult() {
			total.Add(1)
		}
		p.Wait(-1)
		cancel()
		p.ReleaseSelf()
	}
}

func TestRp(t *testing.T) {
	p := NewRp(10)
	p.RegisterExecutor(testExec, ExecMajor)
	p.RegisterExecutor(test2Exec, ExecMinor)
	p.Start()
	ctx, cancel := context.WithCancel(context.Background())
	total := atomic.Int64{}
	//wg := sync.WaitGroup{}
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(100 * time.Millisecond):
				fmt.Println(cap(p.results))
			default:
				for r := p.GetSingleResult(); r != nil; r = p.GetSingleResult() {
					fmt.Println(r)
					total.Add(1)
					time.Sleep(5 * time.Microsecond)
				}
			}
		}
	}()
	for i := 0; i < 10000; i++ {
		for !p.Submit(nil, ExecMajor, 50*time.Millisecond) {
		}
		for !p.Submit(nil, ExecMinor, 50*time.Millisecond) {
		}
	}
	for r := p.GetSingleResult(); r != nil; r = p.GetSingleResult() {
		fmt.Println(r)
		total.Add(1)
	}
	fmt.Println("done submit")
	fmt.Println(total.Load())
	p.Wait(-1)
	cancel()
	p.Stop()
}

func TestRpResize(t *testing.T) {
	p := NewRp(25)
	p.Start()
	p.Resize(15)
	time.Sleep(3 * time.Second)
	p.Stop()
}
