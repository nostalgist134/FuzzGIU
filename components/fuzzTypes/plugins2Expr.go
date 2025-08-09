package fuzzTypes

import (
	"fmt"
	"strings"
)

// Plugins2Expr 将plugin类型转为字符串表达式
func Plugins2Expr(plugins []Plugin) string {
	sb := strings.Builder{}
	for i, p := range plugins {
		if len(p.Args) == 0 && p.Name == "" {
			continue
		}
		sb.WriteString(p.Name)
		// 参数列表
		if len(p.Args) != 0 {
			sb.WriteByte('(')
			for j, a := range p.Args {
				switch a.(type) {
				case string:
					sb.WriteString(fmt.Sprintf("\"%s\"", a))
				default:
					sb.WriteString(fmt.Sprintf("%v", a))
				}
				if j != len(p.Args)-1 {
					sb.WriteString(", ")
				}
			}
			sb.WriteByte(')')
		}
		if i != len(plugins)-1 {
			sb.WriteString(", ")
		}
	}
	return sb.String()
}
