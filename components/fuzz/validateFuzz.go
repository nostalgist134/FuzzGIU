package fuzz

import (
	"errors"
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
)

var (
	errNilJob        = errors.New("job is nil")
	errNoKeywords    = errors.New("job has no fuzz keywords")
	errEmptyIterator = errors.New("job uses a empty iterator")
)

// ValidateJob 判断一个任务是否可执行
func ValidateJob(job *fuzzTypes.Fuzz) error {
	var errTot error
	if job == nil {
		return errNilJob
	}
	if len(job.Preprocess.PlTemp) == 0 {
		errTot = errors.Join(errTot, errNoKeywords)
	}
	for kw, pl := range job.Preprocess.PlTemp {
		if pl.Generators.Type != "wordlist" && pl.Generators.Type != "plugin" {
			errTot = errors.Join(errTot, fmt.Errorf("unsupported payload generator type '%s' for keyword '%s'",
				pl.Generators.Type, kw))
		}
	}
	if retry := job.Request.Retry; retry < 0 {
		errTot = errors.Join(errTot, fmt.Errorf("invalid count of retry %d", retry))
	}
	if timeout := job.Request.Timeout; timeout < 0 {
		errTot = errors.Join(errTot, fmt.Errorf("invalid request timeout %d", timeout))
	}
	if pSize := job.Control.PoolSize; pSize <= 0 {
		errTot = errors.Join(errTot, fmt.Errorf("invalid pool size %d", pSize))
	}
	if delay := job.Control.Delay; delay < 0 {
		errTot = errors.Join(errTot, fmt.Errorf("invalid delay %v", delay))
	}
	iter := &(job.Control.IterCtrl)
	if iter.Start < 0 {
		errTot = errors.Join(errTot, fmt.Errorf("iteration started at negtive %d", iter.Start))
	}
	if iter.Iterator.Name == "" {
		errTot = errors.Join(errTot, errEmptyIterator)
	}
	return errTot
}
