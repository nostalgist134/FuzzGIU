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
)

var outputPara *widgets.Paragraph
var outputBuffer = make([]string, 0)
var currentBufferInd = 0
var lockUpdate = true

var mu = &sync.Mutex{}

func InitOutput(info *fuzzTypes.Fuzz) error {
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
		return err
	}
	FuzzGIUFlag := "     GIUGIU#G#IUGIUGIUGIU                                        #GIUGIUGIUGIUGIUGIU#G#       #G#\n          #I#                                                   #GI            #U# #I#       #I#\n         #U#                                                   #U#            #G# #U#       #U#\n        #G#                                                   #G#            #I# #G#       #G#\n       #I#                                                   #I#            #U# #I#       #I#\n      #U#GIUGIUGIUGIU#G#     #U#GIUGIUGIUGIUGIUGIUGIUGIU    #U#            #G# #U#       #U#\n     #G#            #I#     #G#         #G#         #G#    #G#      #GIU# #I# #G#       #G#\n    #I#            #U#     #I#        #I#         #I#     #I#         G# #U# #I#       #I#\n   #U#            #G#     #U#       #U#         #U#      #U#         I# #G# #U#       #U#\n  #G#            #I#     #G#      #G#         #G#       #G#         UG #I# #G#       #G#\n #I#            #U#     #I#     #I#         #I#        #IU#        IU #U# #IU       IU#\n#U#             #GIUGIU#       GIUGIUGIUGIUGIUGIUGIUGIU#GIUGIUGIUGIU GIUGIU#GIUGIUGIU#"
	sb := strings.Builder{}
	sb.WriteString(FuzzGIUFlag)
	sb.WriteString("\n\n\n")
	sb.WriteString(strings.Repeat("_", 61))
	sb.WriteString("-INFO-")
	sb.WriteString(strings.Repeat("_", 61))
	sb.WriteString("\n\n")
	sb.WriteString(fmt.Sprintf("URL %-20s: %-57s\n", "", info.Send.Request.URL))
	sb.WriteString("Keywords\n")
	for keyword, pTemp := range info.Preprocess.PlTemp {
		sb.WriteString(fmt.Sprintf("    %-20s: %-57s - \t%30s\n", keyword, pTemp.Processors, pTemp.Generators))
	}
	sb.WriteString(fmt.Sprintf("Routine pool size%-7s: %-57d\n", "", info.Misc.PoolSize))
	sb.WriteString(strings.Repeat("_", 128))
	infoPara := widgets.NewParagraph()
	infoPara.Border = false
	infoPara.Text = sb.String()

	infoPara.SetRect(0, 0, 150, 40)
	ui.Render(infoPara)
	outputPara = widgets.NewParagraph()
	outputPara.Border = false
	outputPara.SetRect(0, strings.Count(sb.String(), "\n")+2, 150, 60)
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
			}
		}
	}()
	return nil
}

// 翻页函数，如果direction==true向上翻，反之下翻
func scroll(direction bool, unlock bool) {
	mu.Lock()
	defer mu.Unlock()
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
	mu.Lock()
	defer mu.Unlock()
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

func WaitForQuit() {
	uiEvents := ui.PollEvents()
	for {
		e, ok := <-uiEvents
		if !ok {
			return // termui 关闭，退出监听
		}
		switch e.ID {
		case "q", "<C-c>":
			fmt.Printf("User interrupted, exiting\n")
			ui.Close() // 关闭 termui
			os.Exit(0)
			return
		}
	}
}
