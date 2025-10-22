package tviewOutput

const (
	logoMaxLines     = 8
	globInfoMaxLines = 4
	counterMaxLines  = 1
	outputMaxLines   = 8
	logMaxLines      = 2
	leastHeight      = logMaxLines + globInfoMaxLines + counterMaxLines + outputMaxLines + logoMaxLines + 3*5 - 4

	titleJobInfo       = "JOB_INFO"
	titleOutput        = "OUTPUT"
	titleCounter       = "PROGRESS"
	titlePausedCounter = "PROGRESS(PAUSED)"
	titleLogger        = "LOGS"
	titleLockedOutput  = "OUTPUT(LOCKED)"

	directionUp    = int8(0)
	directionDown  = int8(1)
	directionLeft  = int8(2)
	directionRight = int8(3)

	selectJobInfo = 0
	selectOutput  = 1
	selectLogs    = 2
)
