package libfgiu

import (
	"context"
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/fuzz/fuzzCtx"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"sync/atomic"
	"testing"
	"time"
)

var times = atomic.Int64{}

func testRegister(*fuzzCtx.JobCtx) (int, time.Duration, []*fuzzTypes.Fuzz, error) {
	c := times.Add(1)
	fmt.Println("[test] test called", c)
	return int(c), 0, nil, nil
}

func TestJPool(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	jp, _ := newJobExecPool(15, 15*20, ctx, cancel)
	jp.registerExecutor(testRegister)
	jp.start()
	go func() {
		for {
			res, ok := jp.getResult()
			if ok {
				fmt.Println(res.jid)
			}
		}
	}()
	for i := 0; i < 100; {
		if jp.submit(&fuzzCtx.JobCtx{}) {
			i++
		}
	}
	jp.wait()
	fmt.Println("done")
}

func TestJpSubmit(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	jp, _ := newJobExecPool(15, 15*20, ctx, cancel)
	jp.registerExecutor(testRegister)
	jp.submit(&fuzzCtx.JobCtx{})
}
