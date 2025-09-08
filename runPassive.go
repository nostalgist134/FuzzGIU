package main

import (
	"context"
	"encoding/json"
	"errors"
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
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var accessToken = common.GetRandMarker()
var jobs = make(chan *fuzzTypes.Fuzz, 4096)
var muJobs = sync.Mutex{}

const (
	RouteAddJob    = "add_job"
	RouteGetResult = "get_result"
)

func addJobHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if tok := r.Header.Get("Access-Token"); tok != accessToken {
		http.Error(w, "incorrect access token", http.StatusUnauthorized)
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
	// 捕获中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	mux := http.NewServeMux()
	mux.HandleFunc("/"+RouteAddJob, addJobHandler)

	srv := &http.Server{
		Addr:    opt.General.PassiveAddr,
		Handler: mux,
	}

	go func() {
		f, err := os.Create("token.txt")
		if err != nil {
		} else {
			f.WriteString(accessToken)
			f.Close()
		}
		fmt.Printf("submit job on %s/%s\nget result on %s/%s\naccess token: %s\n", opt.General.PassiveAddr,
			RouteAddJob, opt.General.PassiveAddr, RouteGetResult, accessToken)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("fatal error: %v. exitting...", err)
		}
	}()

	output.SetJobTotal(-1)

	for {
		select {
		case j := <-jobs:
			j.React.OutSettings.NativeStdout = true
			fuzz.DoSingleJob(j)
		case <-quit:
			fmt.Println("\nReceived Ctrl+C, now exiting...")

			// 优雅关闭 HTTP server
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := srv.Shutdown(ctx); err != nil {
				log.Printf("http shutdown error: %v", err)
			}

			os.Remove("token.txt")
			return
		}
	}
}
