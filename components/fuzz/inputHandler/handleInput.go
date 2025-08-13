package inputHandler

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/fuzz/fuzzCommon"
	"github.com/nostalgist134/FuzzGIU/components/input"
	"github.com/nostalgist134/FuzzGIU/components/output"
	"net"
	"os"
)

type cmdFun func(args []string, data []byte) (any, error)

var availableCommands = map[string]cmdFun{
	"get":      get,
	"alter":    alter,
	"exit":     exit,
	"job":      job,
	"poolctrl": poolCtrl,
}

func exit([]string, []byte) (any, error) {
	output.ScreenClose()
	return nil, fuzzCommon.ErrExit
}

func outputToPeer(data any, err error, peer net.Conn) {
	// 出现错误
	if err != nil {
		switch {
		case errors.Is(err, fuzzCommon.ErrJobStop): // 忽略job停止信号
		case errors.Is(err, fuzzCommon.ErrExit): // err为退出信号
			peer.Write([]byte("now exiting\n"))
			fmt.Printf("exit via remote %v\n", peer.RemoteAddr())
			os.Exit(0)
		default:
			peer.Write([]byte(fmt.Sprintf("{\"error\":%q}\n", err)))
			return
		}
	}
	switch res := data.(type) {
	case int, float64, bool, string: // data为基本类型
		peer.Write([]byte(fmt.Sprintln(data)))
	case nil: // data为nil
		peer.Write([]byte{'[', 'n', 'i', 'l', ']', '\n'})
	case []byte:
		peer.Write(res)
	default: // data为其它类型
		jsonBytes, err := json.MarshalIndent(res, "", "  ")
		if err != nil {
			peer.Write([]byte(fmt.Sprintf("{\"error\":%q}\n", err.Error())))
		} else {
			peer.Write(jsonBytes)
			peer.Write([]byte{'\n'})
		}
	}
}

func HandleInput(inp *input.Input) error {
	if fun, ok := availableCommands[inp.Cmd]; ok {
		result, err := fun(inp.Args, inp.Data)
		outputToPeer(result, err, inp.Peer)
		return err
	}
	err := fmt.Errorf("unknown command %s", inp.Cmd)
	outputToPeer(nil, err, inp.Peer)
	return err
}
