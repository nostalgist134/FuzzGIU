package common

import (
	"FuzzGIU/components/fuzzTypes"
	"time"
)

var GlobOutSettings *fuzzTypes.OutputSettings

// globCounter 计数器
var globCounter struct {
	taskCounter counter
	jobCounter  counter
	timeStart   time.Time
	rate        int32
}
