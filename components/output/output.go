package output

import (
	"FuzzGIU/components/fuzzTypes"
	"fmt"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

var infoPara *widgets.Paragraph
var outputPara *widgets.Paragraph
var counterPara *widgets.Paragraph
var outputBuffer = make([]string, 0)
var currentBufferInd = 0
var lockUpdate = true

var muOutputPara = &sync.Mutex{}
var muCounterPara = &sync.Mutex{}

var timeStart time.Time

func InitOutput(info *fuzzTypes.Fuzz) error {
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
		return err
	}
	FuzzGIUFlag := "     GIUGIU#G#IUGIUGIUGIU                                        #GIUGIUGIUGIUGIUGIU#G#       #G#\n          #I#                                                   #GI            #U# #I#       #I#\n         #U#                                                   #U#            #G# #U#       #U#\n        #G#                                                   #G#            #I# #G#       #G#\n       #I#                                                   #I#            #U# #I#       #I#\n      #U#GIUGIUGIUGIU#G#     #U#GIUGIUGIUGIUGIUGIUGIUGIU    #U#            #G# #U#       #U#\n     #G#            #I#     #G#         #G#         #G#    #G#      #GIU# #I# #G#       #G#\n    #I#            #U#     #I#        #I#         #I#     #I#         G# #U# #I#       #I#\n   #U#            #G#     #U#       #U#         #U#      #U#         I# #G# #U#       #U#\n  #G#            #I#     #G#      #G#         #G#       #G#         UG #I# #G#       #G#\n #I#            #U#     #I#     #I#         #I#        #IU#        IU #U# #IU       IU#\n#U#             #GIUGIU#       GIUGIUGIUGIUGIUGIUGIUGIU#GIUGIUGIUGIU GIUGIU#GIUGIUGIU#"
	sbInfoPara := strings.Builder{}
	sbInfoPara.WriteString(FuzzGIUFlag)
	sbInfoPara.WriteString("\n\n\n")
	sbInfoPara.WriteString(strings.Repeat("_", 61))
	sbInfoPara.WriteString("-INFO-")
	sbInfoPara.WriteString(strings.Repeat("_", 61))
	sbInfoPara.WriteString("\n\n")
	sbInfoPara.WriteString(fmt.Sprintf("URL %-20s: %-57s\n", "", info.Send.Request.URL))
	sbInfoPara.WriteString("Keywords\n")
	for keyword, pTemp := range info.Preprocess.PlTemp {
		sbInfoPara.WriteString(fmt.Sprintf("    %-20s: %-57s - \t%30s\n", keyword, pTemp.Processors, pTemp.Generators))
	}
	sbInfoPara.WriteString(fmt.Sprintf("Routine pool size%-7s: %-57d\n", "", info.Misc.PoolSize))
	sbInfoPara.WriteString(strings.Repeat("_", 128))
	infoPara = widgets.NewParagraph()
	infoPara.Border = false
	infoPara.Text = sbInfoPara.String()
	infoPara.SetRect(0, 0, 150, 2+strings.Count(infoPara.Text, "\n"))
	ui.Render(infoPara)
	outputPara = widgets.NewParagraph()
	outputPara.Border = false
	outputPara.SetRect(0,
		strings.Count(sbInfoPara.String(), "\n")+6, // 将output设置在info下方5格的位置，为counter留空间
		150, 60)
	outputPara.Text = ""
	ui.Render(outputPara)
	uiEvents := ui.PollEvents()
	go func() {
		for {
			e, ok := <-uiEvents
			if !ok {
				return
			}
			switch e.ID {
			case "q", "<C-c>":
				fmt.Printf("User interruputed, exiting")
				os.Exit(-1)
				return
			case "w", "k":
				scroll(true, true)
			case "s", "j":
				scroll(false, true)
			case "l":
				muOutputPara.Lock()
				lockUpdate = true
				muOutputPara.Unlock()
			}
		}
	}()
	timeStart = time.Now()
	return nil
}

// 翻页函数，如果direction==true向上翻，反之下翻
func scroll(direction bool, unlock bool) {
	muOutputPara.Lock()
	defer muOutputPara.Unlock()
	var mergedOutput string
	lockUpdate = !unlock
	if direction {
		currentBufferInd--
	} else {
		currentBufferInd++
	}
	if currentBufferInd >= len(outputBuffer) {
		currentBufferInd = len(outputBuffer) - 1
		return
	} else if currentBufferInd <= 0 {
		currentBufferInd = 0
	}
	if currentBufferInd < 4 {
		mergedOutput = mergeOutput(outputBuffer[:])
	} else {
		mergedOutput = mergeOutput(outputBuffer[currentBufferInd-4 : currentBufferInd])
	}
	outputPara.Text = mergedOutput
	ui.Render(outputPara)
}

func mergeOutput(outputBuffer []string) string {
	sb := strings.Builder{}
	for i := 0; i < len(outputBuffer); i++ {
		sb.WriteString(outputBuffer[i])
		if outputBuffer[len(outputBuffer)-1] != "\n" {
			sb.WriteByte('\n')
		}
	}
	return sb.String()
}

func UpdateOutput(update string) {
	muOutputPara.Lock()
	defer muOutputPara.Unlock()
	outputBuffer = append(outputBuffer, update)
	var mergedOutput string
	if len(outputBuffer) < 4 {
		mergedOutput = mergeOutput(outputBuffer[:])
	} else {
		mergedOutput = mergeOutput(outputBuffer[len(outputBuffer)-4:])
	}
	outputPara.Text = mergedOutput
	if lockUpdate {
		currentBufferInd++
		ui.Render(outputPara)
	}
}

type counter struct {
	TaskCount      int
	TaskCountTotal int
	JobCount       int
	JobCountTotal  int
}

var counterGlobal = counter{0, 0, 0, 0}

func rendCounter() {
	muCounterPara.Lock()
	defer muCounterPara.Unlock()
	counterPara.Text = fmt.Sprintf(" fuzz %d/%d job %d/%d", counterGlobal.TaskCount,
		counterGlobal.TaskCountTotal, counterGlobal.JobCount, counterGlobal.JobCountTotal)
	counterPara.Border = true
	counterPara.SetRect(0, strings.Count(infoPara.Text, "\n")+2, 150,
		strings.Count(infoPara.Text, "\n")+5)
	ui.Render(counterPara)
}

func SetCounter(taskCountTotal int, jobCountTotal int) {
	if taskCountTotal != -1 {
		counterGlobal.TaskCountTotal = taskCountTotal
	}
	if jobCountTotal != -1 {
		counterGlobal.JobCountTotal = jobCountTotal
	}
	if counterPara == nil {
		counterPara = widgets.NewParagraph()
	}
	rendCounter()
}

func UpdateCounterTask() {
	counterGlobal.TaskCount++
	rendCounter()
}

func UpdateCounterJob() {
	counterGlobal.JobCount++
	rendCounter()
}

func UpdateTotalJob() {
	counterGlobal.JobCountTotal++
	rendCounter()
}

func Finish() {
	muCounterPara.Lock()
	defer muCounterPara.Unlock()
	counterPara.Text = fmt.Sprintf("%s - all jobs finished, time: %v. Q to quit", counterPara.Text, time.Now().Sub(timeStart))
	ui.Render(counterPara)
}

func WaitForQuit() {
	uiEvents := ui.PollEvents()
	for {
		e, ok := <-uiEvents
		if !ok {
			return // termui 关闭，退出监听
		}
		switch e.ID {
		case "q", "<C-c>":
			fmt.Printf("Exit\n")
			ui.Close() // 关闭 termui
			os.Exit(0)
			return
		}
	}
}
