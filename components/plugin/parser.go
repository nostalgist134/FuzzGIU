package plugin

import (
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"strconv"
	"strings"
)

func unexpectedTokenError(i int, r rune) error {
	return fmt.Errorf("failed to parse plugin string: unexpected token \"%v\" at index %d", r, i)
}

// parseArgStr 识别插件的参数
func parseArgStr(argStr string) any {
	if len(argStr) == 0 {
		return nil
	}
	switch {
	case argStr[0] == '\'' || argStr[0] == '"':
		return argStr[1 : strings.IndexRune(argStr[1:], rune(argStr[0]))+1]
	case strings.Index(argStr, "0x") == 0:
		argRet, err := strconv.ParseInt(argStr[2:], 16, 64)
		if err != nil {
			return nil
		}
		return argRet
	case argStr == "false":
		return false
	case argStr == "true":
		return true
	default:
		var argRet any
		var err error
		// 尝试解析为10进制数
		argRet, err = strconv.ParseInt(argStr, 10, 64)
		if err == nil {
			return int(argRet.(int64))
		}
		// 尝试解析为16进制数
		argRet, err = strconv.ParseInt(argStr, 16, 64)
		if err == nil {
			return int(argRet.(int64))
		}
		// 尝试解析为浮点
		argRet, err = strconv.ParseFloat(argStr, 64)
		if err == nil {
			return argRet
		}
		return nil
	}
}

// ParsePluginsStr 用来解析插件字符串，具体规则参考fuzzTypes.go中的注释
// 解析结果为Plugin类型
func ParsePluginsStr(pluginsStr string) ([]fuzzTypes.Plugin, error) {
	if len(pluginsStr) == 0 {
		return nil, nil
	}
	pluginsStr = strings.TrimSpace(pluginsStr)
	plugins := make([]fuzzTypes.Plugin, 1)
	tmpStrArgBuilder := strings.Builder{}
	tmpPlugNameBuilder := strings.Builder{}
	// 根据下标遍历整个pluginsStr字符串，i为下标，j为当前所处的状态
	// 整个循环中有3种状态，0-在读取插件名，1-在读取参数列表，2-在读取字符串参数，3-读取字符串结束
	for i, j, curPluginInd := 0, 0, 0; i < len(pluginsStr); i++ {
		switch pluginsStr[i] {
		case '(':
			switch j {
			case 0: // 读取到左括号，进入参数名读取状态
				j++
				plugins[curPluginInd].Args = make([]any, 0)
				plugins[curPluginInd].Name = tmpPlugNameBuilder.String()
			case 1, 3: // 在读取参数列表的时候是不允许出现括号参数的，读完字符串参数后也不能
				return nil, unexpectedTokenError(i, '(')
			case 2:
				tmpStrArgBuilder.WriteByte(pluginsStr[i])
			}
		case '\'', '"':
			switch j {
			case 0, 3: // 读取插件名的状态不能直接跳到读取字符串参数的状态，读取单个字符串结束后也不能
				return nil, unexpectedTokenError(i, rune(pluginsStr[i]))
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
					plugins[curPluginInd].Args = append(plugins[curPluginInd].Args, parseArgStr(
						tmpStrArgBuilder.String()))
					tmpStrArgBuilder.Reset()
				}
			case 0:
				return nil, unexpectedTokenError(i, ')')
			case 2:
				tmpStrArgBuilder.WriteByte(pluginsStr[i])
			case 3: // 读完字符串参数遇到右括号说明参数列表的读取结束了
				tmpStrArgBuilder.Reset()
				j = 0
			}
		case ',':
			switch j {
			case 0: // 遇到逗号，如果在读取参数或者读取插件名则代表读取的结束
				plugins[curPluginInd].Name = strings.TrimSpace(tmpPlugNameBuilder.String())
				tmpPlugNameBuilder.Reset()
				curPluginInd++
				plugins = append(plugins, fuzzTypes.Plugin{})
			case 1:
				plugins[curPluginInd].Args = append(plugins[curPluginInd].Args, parseArgStr(tmpStrArgBuilder.String()))
				tmpStrArgBuilder.Reset()
			case 2: // 在字符串中，则继续读取
				tmpStrArgBuilder.WriteByte(pluginsStr[i])
			case 3: // 字符串参数读取完毕，进入下一个参数的读取
				tmpStrArgBuilder.Reset()
				j = 1
			}
		case ' ': // 忽略插件表达式中的括号
			switch j {
			case 2:
				tmpStrArgBuilder.WriteByte(pluginsStr[i])
			default:
				continue
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
				return nil, unexpectedTokenError(i, rune(pluginsStr[i]))
			}
		}
	}
	return plugins, nil
}
