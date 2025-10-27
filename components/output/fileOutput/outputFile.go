package fileOutput

import (
	"errors"
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIU/components/output/outputable"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// filePrologue 写入文件头
func filePrologue(c *Ctx, format string) error {
	// 仅有json以及xml格式需要特定的文件头与文件结尾，其它格式不需要
	if format != "json" && format != "xml" {
		return nil
	}
	if c == nil {
		return errFileOutCtxNil
	}
	if c.f == nil {
		return errFilePointerNil
	}

	var prologue string

	if format == "json" {
		prologue = "["
	} else {
		prologue = "<?xml version=\"1.0\" encoding=\"UTF-8\"?><outputs>"
	}

	_, err := c.f.WriteString(prologue)
	return err
}

// fileEpilogue 写入文件结尾
func fileEpilogue(c *Ctx, format string) error {
	if format != "json" && format != "xml" {
		return nil
	}

	if c == nil {
		return errFileOutCtxNil
	}
	if c.f == nil {
		return errFilePointerNil
	}

	var (
		err      error
		epilogue string
	)

	if format == "json" {
		epilogue = "]"
	} else {
		epilogue = "</outputs>"
	}
	_, err = c.f.WriteString(epilogue)
	return err
}

// NewFileOutputCtx 初始化文件输出，返回文件输出上下文
func NewFileOutputCtx(outSetting *fuzzTypes.OutputSetting, _ int) (*Ctx, error) {
	fname := outSetting.OutputFile
	if fname == "" {
		return nil, errEmptyFName
	}

	// 不允许使用已经存在的文件，一律只能创建新文件
	f, err := os.OpenFile(fname, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	fc := &Ctx{
		f:               f,
		fLog:            nil,
		mu:              new(sync.Mutex),
		muLog:           new(sync.Mutex),
		outputVerbosity: outSetting.Verbosity,
		outputFmt:       outSetting.OutputFormat,
		outputEmpty:     true,
	}

	fc.fileDir, err = filepath.Abs(fname)
	if err != nil {
		fc.fileDir = "."
	}
	fc.fileDir = filepath.Dir(fc.fileDir)

	err = filePrologue(fc, outSetting.OutputFormat)
	if err != nil { // 删除文件，并返回错误
		err2 := fc.f.Close()
		err3 := os.Remove(fname)
		err = errors.Join(err, err2, err3)
		fc = nil
	}
	return fc, err
}

// Output 向特定文件输出结果，允许多个Output并发调用，但是不允许Output与Close之间并发
func (c *Ctx) Output(obj *outputable.OutObj) error {
	if c.f == nil {
		return errFilePointerNil
	}
	if c.closed {
		return errCtxClosed
	}

	formatted := obj.ToFormatBytes(c.outputFmt, false, c.outputVerbosity)

	c.mu.Lock()
	defer c.mu.Unlock()

	var err error

	if !c.outputEmpty { // 使用json或json-line格式输出时，若输出不为空，则每次输出前需要写入分隔符
		switch c.outputFmt {
		case "json":
			_, err = c.f.Write([]byte{','})
		case "json-line":
			_, err = c.f.Write([]byte{'\n'})
		}
		if err != nil {
			return err
		}
	}

	_, err = c.f.Write(formatted)
	if err != nil {
		return err
	}

	c.outputEmpty = false

	return err
}

// Close 关闭文件输出上下文，注意：此方法不保证协程安全，调用时应该自行确定不会再调用Output方法以及此方法不会并发调用
func (c *Ctx) Close() error {
	if c.closed {
		return errCtxClosed
	}
	c.closed = true
	err := fileEpilogue(c, c.outputFmt)
	err0 := c.f.Close()

	var err1 error
	if c.fLog != nil {
		err1 = c.fLog.Close()
	}

	return errors.Join(err, err0, err1)
}

func cutSuffix(fname string) (name string, suffix string) {
	sufStart := strings.LastIndexByte(fname, '.')
	if sufStart == -1 {
		name = fname
	} else {
		name = fname[:sufStart]
		suffix = fname[sufStart:]
	}
	return
}

// Log 写一条日志，会自动创建日志文件
func (c *Ctx) Log(log *outputable.Log) error {
	c.muLog.Lock()
	defer c.muLog.Unlock()

	var err error

	if c.closed {
		return errCtxClosed
	} else if c.fLog == nil { // 若日志文件还未创建就创建一个
		name, suf := cutSuffix(filepath.Base(c.f.Name()))
		logFileName := filepath.Join(c.fileDir,
			fmt.Sprintf("%s_log_%x%s", name, time.Now().UnixNano(), suf))
		c.fLog, err = os.OpenFile(
			logFileName,
			os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
	}

	formatted := log.ToFormatBytes(c.outputFmt)
	_, err = c.fLog.Write(formatted)
	if err != nil {
		return err
	}
	// 理论上来讲log的格式应该和正常输入输出一样用prologue、epilogue来规范，但是我懒得写了
	_, err = c.fLog.Write([]byte("\n"))
	return err
}
