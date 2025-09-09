package output

import (
	"encoding/json"
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/output/common"
	"time"
)

var stop = make(chan struct{}, 1)

func Output(obj *common.OutObj) {
	b, err := json.Marshal(obj)
	if err != nil {
		fmt.Printf("Cannot marshal obj at %p - %v\n", obj, err)
		return
	}
	fmt.Println(string(b))
}

func Log(log string) {
	l, _ := json.Marshal(log)
	fmt.Printf("{\"log\":%s}\n", string(l))
}

func InitOutput() {
	fmt.Println("NATIVE_OUTPUT_BEGIN")
	go func() {
		for {
			select {
			case <-stop:
				return
			default:
			}
			c := common.GetCounter()
			fmt.Printf("{\"counter\":{\"tasks\":%d,\"task_total\":%d,\"jobs\":%d,\"job_total\":%d,\"duration\""+
				":%d,\"rate\":%d}}\n", c[0], c[1], c[2], c[3], common.GetTimeLapsed(), common.GetCurrentRate())
			time.Sleep(225 * time.Millisecond)
		}
	}()
}

func FinishOutput() {
	stop <- struct{}{}
	fmt.Println("NATIVE_OUTPUT_END")
}
