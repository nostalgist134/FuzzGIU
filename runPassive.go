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
	"strconv"
	"syscall"
	"time"
)

var accessToken = common.GetRandMarker()
var jobs = make(chan *fuzzTypes.Fuzz, 4096)

const (
	RouteAddJob        = "/add_job"
	RouteGetResult     = "/get_result"
	RouteGetCurrentJob = "/get_cur_job"
)

func addJobHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if tok := r.Header.Get("Access-Token"); tok != accessToken {
		http.Error(w, `{"error":"incorrect access token"}`, http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// 限制 body 最大 1 MB
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	// 读取请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to read data: %v"}`, err), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// 尝试解码为单个 job
	newJob := new(fuzzTypes.Fuzz)
	if err = json.Unmarshal(body, newJob); err == nil {
		select {
		case jobs <- newJob:
			w.Write([]byte(`{"status":"ok"}`))
		default:
			http.Error(w, `{"error":"job queue full"}`, http.StatusServiceUnavailable)
		}
		return
	}

	// 尝试解码为 job 数组
	var newJobs []*fuzzTypes.Fuzz
	if err = json.Unmarshal(body, &newJobs); err == nil {
		if len(newJobs) == 0 {
			http.Error(w, `{"error":"empty job list"}`, http.StatusBadRequest)
			return
		}
		if len(newJobs) > cap(jobs)-len(jobs) {
			http.Error(w, `{"error":"no enough space in job queue"}`, http.StatusServiceUnavailable)
			return
		}
		for _, j := range newJobs {
			jobs <- j
		}
		w.Write([]byte(fmt.Sprintf(`{"status":"submitted","count":%d}`, len(newJobs))))
		return
	}

	// 两种格式都不匹配
	http.Error(w, `{"error":"invalid job format"}`, http.StatusBadRequest)
}

func getResultHandler(w http.ResponseWriter, r *http.Request) {
	pageArg := r.URL.Query().Get("page")
	sizeArg := r.URL.Query().Get("page_size")

	var objs []json.RawMessage

	// 如果没有分页参数，返回所有对象
	if pageArg == "" && sizeArg == "" {
		objs = output.GetAllMemOutObjects()
	} else {
		// 解析分页参数
		page, err := strconv.Atoi(pageArg)
		if err != nil || page <= 0 {
			http.Error(w, fmt.Sprintf(`{"error":"invalid page %s"}`, pageArg), http.StatusBadRequest)
			return
		}

		size, err := strconv.Atoi(sizeArg)
		if err != nil || size <= 0 {
			http.Error(w, fmt.Sprintf(`{"error":"invalid page size %s"}`, sizeArg), http.StatusBadRequest)
			return
		}

		start := (page - 1) * size
		end := start + size
		objs = output.GetMemOutObjects(start, end)
	}

	bytes, err := json.Marshal(objs)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to marshal results: %v"}`, err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}

func getCurrentJobHandler(w http.ResponseWriter, r *http.Request) {
	if tok := r.Header.Get("Access-Token"); tok != accessToken {
		http.Error(w, `{"error":"incorrect access token"}`, http.StatusUnauthorized)
		return
	}

	job := fuzz.GetCurrentJob()
	if job == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{}`))
		return
	}
	jsonBytes, err := json.Marshal(job)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"marshal error: %v"}`, err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonBytes)
}

// RunPassive 以被动模式运行
func RunPassive(opt *options.Opt) {
	// 捕获中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	mux := http.NewServeMux()
	mux.HandleFunc(RouteAddJob, addJobHandler)
	mux.HandleFunc(RouteGetResult, getResultHandler)
	mux.HandleFunc(RouteGetCurrentJob, getCurrentJobHandler)

	srv := &http.Server{
		Addr:    opt.General.PassiveAddr,
		Handler: mux,
	}

	go func() {
		f, err := os.Create("token.txt")
		if err == nil {
			f.WriteString(accessToken)
			f.Close()
		}
		fmt.Printf("%s\n\t%s\n\t%s\n\t%s\nAccess-Token: %s\n",
			opt.General.PassiveAddr,
			RouteAddJob,
			RouteGetResult,
			RouteGetCurrentJob,
			accessToken,
		)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("fatal error: %v. exiting...", err)
		}
	}()

	output.SetJobTotal(-1)

	for {
		select {
		case j := <-jobs:
			j.React.OutSettings.NativeStdout = true
			common.OutputToWhere = output.OutToMem
			fuzz.DoSingleJob(j)
		case <-quit:
			fmt.Println("\nreceived Ctrl+C, now exiting...")

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			if err := srv.Shutdown(ctx); err != nil {
				log.Printf("http shutdown error: %v", err)
			}
			cancel()

			os.Remove("token.txt")
			return
		}
	}
}
