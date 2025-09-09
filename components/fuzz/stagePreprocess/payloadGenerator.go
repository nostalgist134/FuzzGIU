package stagePreprocess

import (
	"bufio"
	"github.com/nostalgist134/FuzzGIU/components/common"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIU/components/output"
	"github.com/nostalgist134/FuzzGIU/components/plugin"
	"os"
	"sort"
	"strconv"
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
			output.Logf(common.OutputToWhere, "read file %s failed - %v", file, err)
			continue
		}
		payloads = append(payloads, lines...)
	}
	return payloads
}

// 生成一个范围类的int
func genIntStrings(lower int, upper int, base int) []string {
	ret := make([]string, 0)
	for i := lower; i < upper; i++ {
		ret = append(ret, strconv.FormatInt(int64(i), base))
	}
	return ret
}

// emptyStrings 生成一个全为空字符串的切片
func emptyStrings(length int) []string {
	return make([]string, length)
}

// permute 返回所有不重复的排列，maxlen控制最大返回数量，-1表示无限制
func permute(s string, maxLen int) []string {
	reverse := func(chars []rune) {
		for i, j := 0, len(chars)-1; i < j; i, j = i+1, j-1 {
			chars[i], chars[j] = chars[j], chars[i]
		}
	}
	chars := []rune(s)
	// 先排序得到最小字典序的初始排列
	sort.Slice(chars, func(i, j int) bool {
		return chars[i] < chars[j]
	})

	var result []string
	// 将初始排列加入结果集
	result = append(result, string(chars))

	// 检查是否达到最大长度限制
	if maxLen != -1 && len(result) >= maxLen {
		return result
	}

	for {
		i := len(chars) - 2
		for i >= 0 && chars[i] >= chars[i+1] {
			i--
		}
		if i < 0 {
			break
		}
		j := len(chars) - 1
		for chars[j] <= chars[i] {
			j--
		}
		chars[i], chars[j] = chars[j], chars[i]
		reverse(chars[i+1:])

		result = append(result, string(chars))

		if maxLen >= 0 && len(result) >= maxLen {
			break
		}
	}

	return result
}

// generatePayloadsPlugin 使用插件生成payload
func generatePayloadsPlugin(generatorPlugins []fuzzTypes.Plugin) []string {
	payloads := make([]string, 0)
	for _, p := range generatorPlugins {
		switch p.Name {
		case "int":
			if len(p.Args) >= 2 {
				var ok bool
				var lower int
				var upper int
				base := 10
				if lower, ok = p.Args[0].(int); !ok {
					continue
				}
				if upper, ok = p.Args[1].(int); !ok {
					continue
				}
				if len(p.Args) > 2 {
					if base, ok = p.Args[2].(int); !ok {
						base = 10
					}
				}
				payloads = append(payloads, genIntStrings(lower, upper, base)...)
			}
		case "permute":
			if len(p.Args) != 0 {
				src, ok := p.Args[0].(string)
				if !ok {
					break
				}
				maxLen := -1
				if len(p.Args) >= 2 {
					maxLen, ok = p.Args[1].(int)
				}
				payloads = append(payloads, permute(src, maxLen)...)
			}
		case "nil":
			if len(p.Args) != 0 {
				if length, ok := p.Args[0].(int); ok {
					payloads = append(payloads, emptyStrings(length)...)
				}
			}
		case "":
		default:
			payloadsGen := plugin.PayloadGenerator(p)
			payloads = append(payloads, payloadsGen...)
		}
	}
	return payloads
}

// PayloadGenerator 根据payloadGenerator生成payload，同时使用payloadProcessor对生成的payload进行处理
// 返回[]string类型 - 生成的payload
func PayloadGenerator(gen fuzzTypes.PlGen) []string {
	generators := gen.Gen
	generatorType := gen.Type
	// 根据generator生成payload
	var payloads []string
	switch generatorType {
	case "wordlist": // wordlist类型的generator
		files := strings.Split(generators[0].Name, ",")
		payloads = getPayloadsWordlist(files)
	case "plugin": // plugin类型的generator
		payloads = generatePayloadsPlugin(generators)
	default:
		output.Logf(common.OutputToWhere, "unsupported generator type [%s]", generatorType)
		payloads = []string{""}
	}
	// patchLog#3: 修改了payloadGenerator逻辑使得即使生成的payload列表为空也至少会传入一个空字符串，避免doFuzz主循环中curInd为0
	if len(payloads) == 0 {
		payloads = append(payloads, "")
	}
	return payloads
}
