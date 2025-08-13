package input

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/common"
	"github.com/nostalgist134/FuzzGIU/components/output"
	"io"
	"net"
)

// readRawInput 从连接中读取输入，直到遇到全空行
func readRawInput(conn net.Conn) ([]byte, error) {
	reader := bufio.NewReader(conn)
	var buf bytes.Buffer
	for {
		// 读取一行数据（包含换行符）
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				// 如果已经读取了数据，正常返回；否则返回EOF错误
				if buf.Len() > 0 {
					return buf.Bytes(), nil
				}
				return nil, err
			}
			return nil, err
		}
		if len(bytes.TrimSpace(line)) == 0 {
			break
		}
		// 将读取到的行写入缓冲区
		if _, err := buf.Write(line); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

func parseInput(rawInput []byte, peer net.Conn) *Input {
	var firstLine []byte
	var data []byte
	if firstLineInd := bytes.Index(rawInput, []byte{'\n'}); firstLineInd != -1 {
		firstLine = rawInput[:firstLineInd]
		data = rawInput[firstLineInd+1:]
	} else {
		firstLine = rawInput
		data = nil
	}
	cmdAndArgs := bytes.Split(firstLine, []byte{' '})
	cmd := string(cmdAndArgs[0])
	args := make([]string, 0)
	if len(cmdAndArgs) > 1 {
		for _, a := range cmdAndArgs[1:] {
			args = append(args, string(a))
		}
	}
	return &Input{
		Cmd:  cmd,
		Args: args,
		Data: data,
		Peer: peer,
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()
	for {
		rawInput, err := readRawInput(conn)
		if err != nil && err != io.EOF {
			output.Logf(common.OutputToWhere, "error when reading input for %v", conn.RemoteAddr())
			conn.Write([]byte(fmt.Sprintf("error: %v. please try again\n", err)))
			continue
		} else if len(rawInput) == 0 || err == io.EOF { // 若接收到全空行或EOF，则退出当前连接
			break
		}
		inp := parseInput(rawInput, conn)
		inputStk.push(inp)
	}
}

func serve(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			output.Logf(common.OutputToWhere, "error when accepting conn: %v", err)
			continue
		}
		go handleClient(conn)
	}
}

// InitInput 初始化输入，在指定地址上监听
func InitInput(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		listener, err = net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			listener, err = net.Listen("tcp", ":0")
			if err != nil {
				return err
			}
		}
	}
	tcpAddr := listener.Addr().(*net.TCPAddr)
	output.PendLog(fmt.Sprintf("input listening on %s", tcpAddr.String()))
	go serve(listener)
	return nil
}

// GetSingleInput 从input栈中取出一个值，并返回，若没有需要处理的值，则返回nil, false
func GetSingleInput() (*Input, bool) {
	if !Enabled {
		return nil, false
	} else if ret := inputStk.pop(); ret != nil {
		return ret, true
	}
	return nil, false
}
