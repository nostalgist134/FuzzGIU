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
	"poolCtrl": poolCtrl,
}

func exit([]string, []byte) (any, error) {
	output.ScreenClose()
	return nil, fuzzCommon.ErrExit
}

func outputToPeer(data any, err error, peer net.Conn) {
	// 出现错误
	if err != nil && !errors.Is(err, fuzzCommon.ErrJobStop) && !errors.Is(err, fuzzCommon.ErrExit) {
		peer.Write([]byte(fmt.Sprintf("{\"error\":%q}", err)))
		return
	}
	// 若调用的命令为exit则直接退出
	if errors.Is(err, fuzzCommon.ErrExit) {
		peer.Write([]byte("now exiting\n"))
		fmt.Printf("exit via remote poolCtrl %v\n", peer.RemoteAddr())
		os.Exit(0)
	}
	switch res := data.(type) {
	case int, float64, bool, string: // data为基本类型
		peer.Write([]byte(fmt.Sprintf("%v", data)))
	case nil: // data为nil
		peer.Write([]byte{'[', 'n', 'i', 'l', ']'})
	default: // data为其它类型
		jsonBytes, err := json.MarshalIndent(res, "", "  ")
		if err != nil {
			peer.Write([]byte(fmt.Sprintf("{\"error\":%q}", err.Error())))
		} else {
			peer.Write(jsonBytes)
		}
	}
}

func HandleInput(inp *input.Input) error {
	if fun, ok := availableCommands[inp.Cmd]; ok {
		result, err := fun(inp.Args, inp.Data)
		outputToPeer(result, err, inp.Peer)
		return err
	}
	return fmt.Errorf("unknown command %s", inp.Cmd)
}
