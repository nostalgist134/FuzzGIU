package fuzzCommon

import (
	"errors"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIU/components/output"
)

type JobQueue []*fuzzTypes.Fuzz

var curFuzz *fuzzTypes.Fuzz

var ErrJobStop = errors.New("job stop")
var ErrExit = errors.New("exit")

var globJq *JobQueue

func (jq *JobQueue) AddJob(fuzz *fuzzTypes.Fuzz) {
	*jq = append(*jq, fuzz)
}

func SetCurFuzz(fuzz1 *fuzzTypes.Fuzz) {
	curFuzz = fuzz1
}

func GetCurFuzz() *fuzzTypes.Fuzz {
	return curFuzz
}

func SetJQ(jq *JobQueue) {
	globJq = jq
}

func GetJQ() JobQueue {
	return *globJq
}

func AddJob(newJob *fuzzTypes.Fuzz) bool {
	if globJq != nil {
		globJq.AddJob(newJob)
		output.SetJobTotal(int64(len(*globJq)))
		return true
	}
	return false
}
