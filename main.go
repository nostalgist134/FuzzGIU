package main

import (
	"FuzzGIU/components/fuzz"
	"FuzzGIU/components/options"
	"FuzzGIU/components/output"
)

func main() {
	opts := options.ParseOptCmdline()
	fuzz2 := opt2fuzz(opts)
	if err := output.InitOutput(fuzz2); err != nil {
		panic(err)
	}
	fuzz.JQ.AddJob(fuzz2)
	fuzz.DoJobs(opts.Output.ToFile)
	output.WaitForQuit()
}
