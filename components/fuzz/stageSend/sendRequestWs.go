package stageSend

import (
	"FuzzGIU/components/common"
	"FuzzGIU/components/fuzzTypes"
	"bytes"
	"github.com/gorilla/websocket"
	"sync"
	"time"
)

var timeConn time.Time
var connectionPool = sync.Pool{
	New: func() interface{} {
		return new(websocket.Conn)
	},
}

func sendRequestWs(req *fuzzTypes.Req, timeout int, retry int, retryRegex string) *fuzzTypes.Resp {
	resp := &fuzzTypes.Resp{}
	startTime := time.Now()

	for ; retry >= 0; retry-- {
		conn := connectionPool.Get().(*websocket.Conn)
		defer connectionPool.Put(conn)

		// 建立 WebSocket 连接
		var err error
		conn, _, err = websocket.DefaultDialer.Dial(req.URL, nil)
		if err != nil {
			resp.ErrMsg = err.Error()
			continue
		}
		defer conn.Close()

		// 设置超时
		if timeout > 0 {
			conn.SetReadDeadline(time.Now().Add(time.Duration(timeout) * time.Second))
		}

		// 发送请求数据
		err = conn.WriteMessage(websocket.TextMessage, []byte(req.Data))
		if err != nil {
			resp.ErrMsg = err.Error()
			continue
		}

		// 读取响应
		_, message, err := conn.ReadMessage()
		if err != nil {
			resp.ErrMsg = err.Error()
			continue
		}

		resp.RawResponse = message
		resp.Size = len(message)
		resp.Words = len(bytes.Fields(message))
		resp.Lines = bytes.Count(message, []byte{'\n'})
		if message[len(message)-1] != '\n' {
			resp.Lines++
		}

		// 检查是否需要重试
		if retryRegex != "" && common.RegexMatch(message, retryRegex) {
			continue
		}

		// 计算响应时间
		resp.ResponseTime = time.Since(startTime)
		resp.ErrMsg = ""
		break
	}

	return resp
}
