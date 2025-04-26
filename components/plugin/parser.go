package plugin

import (
	"fmt"
	"strconv"
	"strings"
)

func unexpectedTokenError(i int, r rune) {
	fmt.Printf("Error parsing plugin string: unexpected token \"%v\" at index %d\n", r, i)
}

// parseArgStr 识别插件的参数
func parseArgStr(argStr string) any {
	// 以“'”或者“"”开头的参数解析为字符串
	if argStr[0] == '\'' || argStr[0] == '"' {
		// 找到“'”或者“"”第二次出现的地方作为字符串的结束
		return argStr[1 : strings.IndexRune(argStr[1:], rune(argStr[0]))+1]
	} else if arg, err := strconv.ParseBool(argStr); err == nil {
		// 尝试解析为bool类型
		return arg
	} else if arg, err := strconv.ParseInt(argStr, 10, 64); err == nil {
		// 尝试解析为10进制int类型
		return arg
	} else if arg, err := strconv.ParseInt(argStr, 16, 64); err == nil {
		// 尝试解析为16进制int类型
		return arg
		// patchLog#5: 将解析float放到int之后，因为parseFloat也能解析整数，这会导致整数型参数永远无法解析
	} else if arg, err := strconv.ParseFloat(argStr, 64); err == nil { // 尝试解析为float类型
		return arg
	} else {
		// 未知的参数类型，返回nil
		return nil
	}
}

// ParsePluginsStr 用来解析插件字符串，具体规则参考fuzzTypes.go中的注释
// 解析结果为Plugin类型
func ParsePluginsStr(pluginsStr string) []Plugin {
	if len(pluginsStr) == 0 {
		return nil
	}
	pluginsStr = strings.TrimSpace(pluginsStr)
	plugins := make([]Plugin, 1)
	tmpStrArgBuilder := strings.Builder{}
	tmpPlugNameBuilder := strings.Builder{}
	// 根据下标遍历整个pluginsStr字符串，i为下标，j为当前所处的状态
	// 整个循环中有3种状态，0代表在读取插件名，1代表在读取参数列表，2代表在读取字符串参数，3代表读取字符串结束
	for i, j, curPluginInd := 0, 0, 0; i < len(pluginsStr); i++ {
		switch pluginsStr[i] {
		case '(':
			switch j {
			case 0: // 读取到左括号，进入参数名读取状态
				j++
				plugins[curPluginInd].Args = make([]any, 0)
				plugins[curPluginInd].Name = tmpPlugNameBuilder.String()
			case 1, 3: // 在读取参数列表的时候是不允许出现括号参数的，读完字符串参数后也不能
				unexpectedTokenError(i, '(')
				return nil
			case 2:
				tmpStrArgBuilder.WriteByte(pluginsStr[i])
			}
		case '\'', '"':
			switch j {
			case 0, 3: // 读取插件名的状态不能直接跳到读取字符串参数的状态，读取单个字符串结束后也不能
				unexpectedTokenError(i, rune(pluginsStr[i]))
				return nil
			case 1:
				j++
				tmpStrArgBuilder.WriteByte(pluginsStr[i])
			case 2: // 引号匹配时，结束读取字符串参数
				if tmpArg := tmpStrArgBuilder.String(); len(tmpArg) > 0 && tmpArg[0] == pluginsStr[i] {
					j++
					tmpStrArgBuilder.WriteByte(pluginsStr[i])
					plugins[curPluginInd].Args = append(plugins[curPluginInd].Args,
						parseArgStr(tmpStrArgBuilder.String()))
					tmpStrArgBuilder.Reset()
				} else {
					tmpStrArgBuilder.WriteByte(pluginsStr[i])
				}
			}
		case ')':
			switch j { // 遇到右括号时，如果是在读取参数的情况下就停止，如果是参数名则返回错误，如果是在字符串中就继续读
			case 1:
				j--
				if len(tmpStrArgBuilder.String()) > 0 {
					plugins[curPluginInd].Args = append(plugins[curPluginInd].Args, parseArgStr(tmpStrArgBuilder.String()))
					tmpStrArgBuilder.Reset()
				}
			case 0:
				unexpectedTokenError(i, ')')
				return nil
			case 2:
				tmpStrArgBuilder.WriteByte(pluginsStr[i])
			case 3: // 读完字符串参数遇到右括号说明参数列表的读取结束了
				tmpStrArgBuilder.Reset()
				j = 0
			}
		case ',':
			switch j { // 遇到逗号，如果在读取参数或者读取插件名则代表读取的结束
			case 0:
				plugins[curPluginInd].Name = strings.TrimSpace(tmpPlugNameBuilder.String())
				tmpPlugNameBuilder.Reset()
				curPluginInd++
				plugins = append(plugins, Plugin{})
			case 1:
				plugins[curPluginInd].Args = append(plugins[curPluginInd].Args, parseArgStr(tmpStrArgBuilder.String()))
				tmpStrArgBuilder.Reset()
			case 2: // 在字符串中，则继续读取
				tmpStrArgBuilder.WriteByte(pluginsStr[i])
			case 3: // 字符串参数读取完毕，进入下一个参数的读取
				tmpStrArgBuilder.Reset()
				j = 1
			}
		default:
			switch j {
			case 0:
				tmpPlugNameBuilder.WriteByte(pluginsStr[i])
				if i == len(pluginsStr)-1 {
					plugins[curPluginInd].Name = tmpPlugNameBuilder.String()
				}
			case 1, 2:
				tmpStrArgBuilder.WriteByte(pluginsStr[i])
			case 3: // 字符串参数读取之后如果再有其它的字符视为语法错误
				unexpectedTokenError(i, rune(pluginsStr[i]))
				return nil
			}
		}
	}
	return plugins
}
