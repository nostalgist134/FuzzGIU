package tviewOutput

import (
	"sync"
)

var (
	proportions   = []int{6, 2, 9, 2}
	titles        = []string{"JOB_INFO", "PROGRESS", "OUTPUT", "LOGS", "OUTPUT(LOCKED)", "LOGS(LOCKED)"}
	appCreateOnce = sync.Once{}
	screen        *tviewScreen
)
