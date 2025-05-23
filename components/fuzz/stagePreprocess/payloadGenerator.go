package stagePreprocess

import (
	"FuzzGIU/components/plugin"
	"bufio"
	"fmt"
	"os"
	"strings"
)

// 从字典文件中按行读取payload，跳过空行，用于wordlist generator
func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	lines := make([]string, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) == "" {
			continue
		}
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// getPayloadsWordlist 从文本文件中直接读取payload
func getPayloadsWordlist(files []string, processorPlugins []plugin.Plugin) []string {
	payloads := make([]string, 0)
	for _, file := range files {
		lines, err := readLines(file)
		if err != nil {
			panic(err)
		}
		for _, payload := range lines {
			if len(processorPlugins) > 0 {
				payloads = append(payloads, PayloadProcessor(payload, processorPlugins))
			} else {
				payloads = append(payloads, payload)
			}
		}
	}
	return payloads
}

// generatePayloadsPlugin 使用插件生成payload
func generatePayloadsPlugin(generatorPlugins []plugin.Plugin, processorPlugins []plugin.Plugin) []string {
	payloads := make([]string, 0)
	for _, plugin1 := range generatorPlugins {
		payloadGen := plugin.Call(plugin.PTypePlGen, plugin1, nil, nil).([]string)
		for _, payload := range payloadGen {
			if len(processorPlugins) > 0 {
				payloads = append(payloads, PayloadProcessor(payload, processorPlugins))
			} else {
				payloads = append(payloads, payload)
			}
		}
	}
	return payloads
}

// GeneratePayloads 根据payloadGenerator生成payload，同时使用payloadProcessor对生成的payload进行处理
// 返回[]string类型 - 生成的payload
func GeneratePayloads(payloadGenerator string, payloadProcessor string) []string {
	generators := payloadGenerator[:strings.LastIndex(payloadGenerator, "|")]
	generatorType := payloadGenerator[strings.LastIndex(payloadGenerator, "|")+1:]
	// 根据generator生成payload
	var payloads []string
	switch generatorType {
	case "wordlist": // wordlist类型的generator
		payloads = getPayloadsWordlist(strings.Split(generators, ","), plugin.ParsePluginsStr(payloadProcessor))
	case "plugin": // plugin类型的generator
		payloads = generatePayloadsPlugin(plugin.ParsePluginsStr(generators), plugin.ParsePluginsStr(payloadProcessor))
	default:
		fmt.Printf("Unsupported generator type: %s\n", generatorType)
		payloads = []string{""}
	}
	// patchLog#3: 修改了payloadGenerator逻辑使得即使生成的payload列表为空也至少会传入一个空字符串，避免之后主循环中curInd为0
	if len(payloads) == 0 {
		payloads = append(payloads, "")
	}
	return payloads
}
