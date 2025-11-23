package fuzz

import (
	"errors"
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIU/components/resourcePool"
	"strings"
)

var (
	errNilJob               = errors.New("job is nil")
	errNoKeywords           = errors.New("job has no fuzz keywords")
	errEmptyIterator        = errors.New("job uses a empty iterator")
	errEmptyURL             = errors.New("job uses a empty url")
	errMultiKeywords        = errors.New("job specified sniper or recursion but provided multiple keywords")
	errPreprocPathTraverse  = errors.New("preprocessors have path traverse")
	errReactPathTraverse    = errors.New("reactor plugin has path traverse")
	errIteratorPathTraverse = errors.New("iterator plugin has path traverse")
)

func pluginPathTraverse(p fuzzTypes.Plugin) bool {
	pName := strings.Replace(p.Name, "\\", "/", -1)
	if strings.Contains(pName, "../") || strings.Contains(pName, "/..") {
		return true
	}
	return false
}

func pluginsPathTraverse(plugins []fuzzTypes.Plugin) bool {
	for _, p := range plugins {
		if pluginPathTraverse(p) {
			return true
		}
	}
	return false
}

// ValidateJob 判断一个任务是否能够正常执行
func ValidateJob(job *fuzzTypes.Fuzz) error {
	if job == nil {
		return errNilJob
	}

	var errTot error
	kwCount := len(job.Preprocess.PlMeta)
	if kwCount == 0 {
		errTot = errors.Join(errTot, errNoKeywords)
	}
	keywords := resourcePool.StringSlices.Get(kwCount)
	defer resourcePool.StringSlices.Put(keywords)
	i := 0
	for kw, pl := range job.Preprocess.PlMeta {
		if pl == nil {
			errTot = errors.Join(errTot, fmt.Errorf("keyword '%s' meta data is nil", kw))
			continue
		}
		keywords[i] = kw
		for j := 0; j < i; j++ { // 判断是否有fuzz关键字互相包含的情况（这种情况会导致模板解析失败）
			if strings.Contains(keywords[j], keywords[i]) || strings.Contains(keywords[i], keywords[j]) {
				errTot = errors.Join(errTot, fmt.Errorf("keyword %s overlapped with %s",
					keywords[j], keywords[i]))
			}
		}
		if pl.Generators.Type != "wordlist" && pl.Generators.Type != "plugin" {
			errTot = errors.Join(errTot,
				fmt.Errorf("unsupported payload generator type '%s' for keyword '%s'", pl.Generators.Type, kw))
		} else if pl.Generators.Type == "plugin" && pluginsPathTraverse(pl.Generators.Gen) {
			errTot = errors.Join(errTot, fmt.Errorf("keyword '%s' generator has path traverse", kw))
		}
		if pluginsPathTraverse(pl.Processors) {
			errTot = errors.Join(errTot, fmt.Errorf("keyword '%s' processor has path traverse", kw))
		}
		i++
	}
	if pluginsPathTraverse(job.Preprocess.Preprocessors) {
		errTot = errors.Join(errTot, errPreprocPathTraverse)
	}
	if pluginPathTraverse(job.Control.IterCtrl.Iterator) {
		errTot = errors.Join(errTot, errIteratorPathTraverse)
	}
	if pluginPathTraverse(job.React.Reactor) {
		errTot = errors.Join(errTot, errReactPathTraverse)
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
	if len(keywords) > 1 && (iter.Iterator.Name == "sniper" || job.React.RecursionControl.MaxRecursionDepth > 0) {
		errTot = errors.Join(errTot, errMultiKeywords)
	}
	if job.Preprocess.ReqTemplate.URL == "" {
		errTot = errors.Join(errTot, errEmptyURL)
	}
	return errTot
}
