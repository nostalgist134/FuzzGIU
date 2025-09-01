package main

import (
	"encoding/json"
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/common"
	"github.com/nostalgist134/FuzzGIU/components/fuzz"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIU/components/options"
	"github.com/nostalgist134/FuzzGIU/components/output"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
)

var randMarker = common.GetRandMarker()
var jobs = make(chan *fuzzTypes.Fuzz, 4096)
var muJobs = sync.Mutex{}

func handler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if tok := r.Header.Get("Access-Token"); tok != randMarker {
		http.Error(w, "access token not right", http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 限制 body 最大 1 MB
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	// 读取请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to read data: %v", err), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")

	// 尝试解码为单个 job
	newJob := new(fuzzTypes.Fuzz)
	if err = json.Unmarshal(body, newJob); err == nil {
		muJobs.Lock()
		defer muJobs.Unlock()
		select {
		case jobs <- newJob:
			w.Write([]byte("ok"))
		default:
			http.Error(w, "job queue full", http.StatusServiceUnavailable)
		}
		return
	}

	// 尝试解码为 job 数组
	var newJobs []*fuzzTypes.Fuzz
	if err = json.Unmarshal(body, &newJobs); err == nil {
		muJobs.Lock()
		defer muJobs.Unlock()
		if len(newJobs) > cap(jobs)-len(jobs) {
			http.Error(w, "no enough space in job queue", http.StatusServiceUnavailable)
			return
		}
		for _, j := range newJobs {
			jobs <- j
		}
		w.Write([]byte(fmt.Sprintf("submit %d jobs", len(newJobs))))
		return
	}

	// 两种格式都不匹配
	http.Error(w, "invalid job format", http.StatusBadRequest)
}

// RunPassive 以被动模式运行
func RunPassive(opt *options.Opt) {
	http.HandleFunc("/add_job", handler)
	go func() {
		f, err := os.Create("token.txt")
		defer os.Remove("token.txt")
		if err == nil {
			f.WriteString(randMarker)
			f.Close()
		}
		fmt.Printf("submit job on %s/add_job with access token: %s...\n", opt.General.PassiveAddr, randMarker)
		if err := http.ListenAndServe(opt.General.PassiveAddr, nil); err != nil {
			log.Fatalf("fatal error: %v. exitting...", err)
		}
	}()
	output.SetJobCounter(-1)
	for j := range jobs {
		// 输出到原生 stdout
		j.React.OutSettings.NativeStdout = true
		fuzz.DoSingleJob(j)
	}
}
