package inputHandler

import (
	"errors"
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/fuzz/fuzzCommon"
	"reflect"
	"strings"
)

var gettableObjects = []string{
	"fuzz",
	"jq",
}

var ErrMissingObjName = errors.New("missing object name to get")

// 使用反射根据路径获取结构体成员值（大小写不敏感）
func getFieldValue(data any, path string) (any, error) {
	// 将路径按"."分割成各个部分
	parts := strings.Split(path, ".")
	if len(parts) < 2 {
		return data, nil
	}
	parts = parts[1:]
	// 获取反射值
	val := reflect.ValueOf(data)

	// 如果是指针类型，获取其指向的值
	for val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// 遍历路径的每个部分
	for _, part := range parts {
		// 检查当前值是否为结构体
		if val.Kind() != reflect.Struct {
			return nil, fmt.Errorf("%s of %s is not struct", part, path)
		}
		// 转换为小写以便大小写不敏感比较
		partLower := strings.ToLower(part)
		var field reflect.Value
		found := false
		// 遍历结构体所有字段，进行大小写不敏感匹配
		for i := 0; i < val.NumField(); i++ {
			fieldName := val.Type().Field(i).Name
			if strings.ToLower(fieldName) == partLower {
				field = val.Field(i)
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("unknown field: %s", part)
		}
		// 更新当前值为字段值
		val = field
		// 如果是指针类型，获取其指向的值
		for val.Kind() == reflect.Ptr {
			val = val.Elem()
		}
	}
	return val.Interface(), nil
}

func getObjByName(name string) (any, error) {
	lowerName := strings.ToLower(name)
	// 在gettableObjects中寻找
	switch lowerName {
	case gettableObjects[0]:
		return fuzzCommon.GetCurFuzz(), nil
	case gettableObjects[1]:
		return fuzzCommon.GetJQ(), nil
	}
	switch {
	case lowerName == "jqlen": // 动态获取JQ（任务列表）的长度并返回
		return len(fuzzCommon.GetJQ()), nil
	case strings.HasPrefix(lowerName, "fuzz."): // 获取当前任务fuzz结构体某个字段的值
		return getFieldValue(fuzzCommon.GetCurFuzz(), name)
	}
	return nil, fmt.Errorf("unknown object: %s", name)
}

func get(args []string, _ []byte) (any, error) {
	if len(args) < 1 {
		return nil, ErrMissingObjName
	}
	return getObjByName(args[0])
}
