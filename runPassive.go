package main

import (
	"encoding/json"
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/fuzz"
	"github.com/nostalgist134/FuzzGIU/components/fuzz/fuzzCommon"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIU/components/options"
	"net/http"
)

var jobs = make(chan *fuzzTypes.Fuzz, 256)

func handler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	newJob := new(fuzzTypes.Fuzz)
	if err := json.NewDecoder(r.Body).Decode(newJob); err != nil {
		http.Error(w, "Invalid json data", http.StatusBadRequest)
		return
	}
	select {
	case jobs <- newJob:
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{\"status\":\"ok\"}"))
	default:
		http.Error(w, "Job queue full", http.StatusServiceUnavailable)
	}
}

func runPassive(opt *options.Opt) {
	http.HandleFunc("/job", handler)
	go func() {
		fmt.Printf("listening on %s...\n", opt.General.PassiveAddr)
		if err := http.ListenAndServe(opt.General.PassiveAddr, nil); err != nil {
			panic(err)
		}
	}()
	for j := range jobs {
		// 输出到原生stdout
		j.React.OutSettings.NativeStdout = true
		fuzz.JQ = fuzzCommon.JobQueue{j}
		fuzz.DoJobs()
	}
}
