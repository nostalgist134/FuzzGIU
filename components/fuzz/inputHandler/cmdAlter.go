package inputHandler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/fuzz/fuzzCommon"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIU/components/output"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var curFuzz *fuzzTypes.Fuzz

var settableObjects = []string{
	"fuzz.Misc.Delay",
	"fuzz.Misc.DelayGranularity",
	"fuzz.Send.RetryCode",
	"fuzz.Send.Retry",
	"fuzz.Send.RetryCode",
	"fuzz.Send.RetryRegex",
	"fuzz.Send.Timeout",
	"fuzz.Send.Proxies",
	"fuzz.React.Filter.Code",
	"fuzz.React.Filter.Lines",
	"fuzz.React.Filter.Words",
	"fuzz.React.Filter.Size",
	"fuzz.React.Matcher.Code",
	"fuzz.React.Matcher.Lines",
	"fuzz.React.Matcher.Words",
	"fuzz.React.Matcher.Size",
}

var addableObjects = []string{
	"fuzz.Send.Proxies",
	"fuzz.React.Filter.Code",
	"fuzz.React.Filter.Lines",
	"fuzz.React.Filter.Words",
	"fuzz.React.Filter.Size",
	"fuzz.React.Matcher.Code",
	"fuzz.React.Matcher.Lines",
	"fuzz.React.Matcher.Words",
	"fuzz.React.Matcher.Size",
}

var ErrMissingAlterOper = errors.New("missing alter operation(add/set)")
var ErrMissingAlterName = errors.New("missing alter field name")
var ErrPtrNotAssignable = errors.New("ptr not assignable")
var ErrSlicePtrUnsupport = errors.New("slice type unsupported")

func getFieldAddress(v any, path string) (any, error) {
	// 检查v是否为指针类型
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr {
		return nil, errors.New("v must be a pointer to struct")
	}
	// 获取指针指向的元素（结构体）
	val = val.Elem()
	if val.Kind() != reflect.Struct {
		return nil, errors.New("v must point to a struct")
	}
	// 分割路径
	fields := strings.Split(path, ".")
	if len(fields) < 2 {
		return nil, errors.New("invalid path")
	}
	fields = fields[1:]
	// 逐层解析字段
	for i, fieldName := range fields {
		// 如果当前是指针类型，获取其指向的值
		for val.Kind() == reflect.Ptr {
			val = val.Elem()
		}
		// 确保当前是结构体类型
		if val.Kind() != reflect.Struct {
			return nil, fmt.Errorf("field %s is not a struct", strings.Join(fields[:i+1], "."))
		}
		// 获取字段
		field := val.FieldByName(fieldName)
		if !field.IsValid() {
			return nil, fmt.Errorf("field %s not found", strings.Join(fields[:i+1], "."))
		}
		// 最后一个字段需要可寻址才能获取地址
		if i == len(fields)-1 && !field.CanAddr() {
			return nil, fmt.Errorf("field %s is not addressable", path)
		}
		// 移动到下一层字段
		val = field
	}
	// 获取最后一个字段的地址
	return val.Addr().Interface(), nil
}

func bytes2Ranges(b []byte) []fuzzTypes.Range {
	bytesRanges := bytes.Split(b, []byte{','})
	ranges := make([]fuzzTypes.Range, 0)
	for _, eachRng := range bytesRanges {
		var upper int64
		var lower int64
		var err error
		if joint := bytes.Index(eachRng, []byte{'-'}); joint != -1 && joint != len(eachRng)-1 {
			lower, err = strconv.ParseInt(string(eachRng[:joint]), 10, 64)
			if err != nil {
				continue
			}
			upper, err = strconv.ParseInt(string(eachRng[joint+1:]), 10, 64)
			if err != nil {
				continue
			}
		}
		ranges = append(ranges, fuzzTypes.Range{Upper: int(upper), Lower: int(lower)})
	}
	return ranges
}

func assignToPtr(ptr any, data []byte) error {
	dataString := string(data)
	switch actual := ptr.(type) {
	case *string:
		*actual = dataString
	case *[]string:
		var slic []string
		err := json.Unmarshal(data, &slic)
		if err != nil {
			return err
		}
		*actual = slic
	case *int:
		i, err := strconv.ParseInt(dataString, 10, 64)
		if err != nil {
			return err
		}
		*actual = int(i)
	case *time.Duration:
		d, err := strconv.ParseInt(dataString, 10, 64)
		if err != nil {
			return err
		}
		*actual = time.Duration(d)
	case *[]fuzzTypes.Range:
		*actual = bytes2Ranges(data)
	default:
		return ErrPtrNotAssignable
	}
	return nil
}

func addToSlicePtr(ptr any, data []byte) error {
	switch actual := ptr.(type) {
	case *[]string:
		var slic []string
		err := json.Unmarshal(data, &slic)
		if err != nil {
			return err
		}
		*actual = append(*actual, slic...)
	case *[]fuzzTypes.Range:
		var ranges []fuzzTypes.Range
		err := json.Unmarshal(data, &ranges)
		if err != nil {
			ranges = bytes2Ranges(data)
		}
		*actual = append(*actual, ranges...)
	default:
		return ErrSlicePtrUnsupport
	}
	return nil
}

func alterAdd(field string, data []byte) error {
	lowerFld := strings.ToLower(field)
	for _, o := range addableObjects {
		if strings.ToLower(o) == lowerFld {
			addr, err := getFieldAddress(curFuzz, field)
			if err != nil {
				return err
			}
			return addToSlicePtr(addr, data)
		}
	}
	return fmt.Errorf("unaddable field %s", field)
}

func alterSet(field string, data []byte) error {
	lowerFld := strings.ToLower(field)
	for _, o := range settableObjects {
		if strings.ToLower(o) == lowerFld {
			addr, err := getFieldAddress(curFuzz, field)
			if err != nil {
				return err
			}
			return assignToPtr(addr, data)
		}
	}
	return fmt.Errorf("unsettable field %s", field)
}

func alter(args []string, data []byte) (any, error) {
	curFuzz = fuzzCommon.GetCurFuzz()
	defer output.UpdateScreenInfoPage(curFuzz)
	if len(args) == 0 {
		return nil, ErrMissingAlterOper
	}
	if len(args) == 1 {
		return nil, ErrMissingAlterName
	}
	var err error
	switch strings.ToLower(args[0]) {
	case "add":
		err = alterAdd(args[1], data)
		if err != nil {
			return nil, err
		}
		return bytesOk, nil
	case "set":
		err = alterSet(args[1], data)
		if err != nil {
			return nil, err
		}
		return bytesOk, nil
	default:
		return nil, fmt.Errorf("unknown alter operation '%s'", args[0])
	}
}
