package stagePreprocess

import (
	"bufio"
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/common"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIU/components/output"
	"github.com/nostalgist134/FuzzGIU/components/plugin"
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
func getPayloadsWordlist(files []string) []string {
	payloads := make([]string, 0)
	for _, file := range files {
		lines, err := readLines(file)
		if err != nil {
			output.Log(fmt.Sprintf("read file %s failed - %v", file, err), common.OutputToWhere)
			return []string{}
		}
		for _, payload := range lines {
			payloads = append(payloads, payload)
		}
	}
	return payloads
}

// generatePayloadsPlugin 使用插件生成payload
func generatePayloadsPlugin(generatorPlugins []fuzzTypes.Plugin) []string {
	payloads := make([]string, 0)
	for _, p := range generatorPlugins {
		// payloadGen := plugin.Call(plugin.PTypePlGen, plugin1, nil, nil).([]string)
		payloadsGen := plugin.PayloadGenerator(p)
		for _, payload := range payloadsGen {
			payloads = append(payloads, payload)
		}
	}
	return payloads
}

// GeneratePayloads 根据payloadGenerator生成payload，同时使用payloadProcessor对生成的payload进行处理
// 返回[]string类型 - 生成的payload
func GeneratePayloads(payloadGenerator fuzzTypes.PlGen) []string {
	generators := payloadGenerator.Gen
	generatorType := payloadGenerator.Type
	// 根据generator生成payload
	var payloads []string
	switch generatorType {
	case "wordlist": // wordlist类型的generator
		files := strings.Split(generators[0].Name, ",")
		payloads = getPayloadsWordlist(files)
	case "plugin": // plugin类型的generator
		payloads = generatePayloadsPlugin(generators)
	default:
		output.Log(fmt.Sprintf("unsupported generator type [%s]", generatorType), common.OutputToWhere)
		payloads = []string{""}
	}
	// patchLog#3: 修改了payloadGenerator逻辑使得即使生成的payload列表为空也至少会传入一个空字符串，避免doFuzz主循环中curInd为0
	if len(payloads) == 0 {
		payloads = append(payloads, "")
	}
	return payloads
}
