package output

import (
	"FuzzGIU/components/output/common"
	"encoding/json"
	"fmt"
	"time"
)

var stop = make(chan struct{}, 1)

func NativeStdOutput(obj *common.OutObj) {
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

func InitOutputStdout() {
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

func FinishOutputStdout() {
	stop <- struct{}{}
	fmt.Println("NATIVE_OUTPUT_END")
}
