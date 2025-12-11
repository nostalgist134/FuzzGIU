package requestWs

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/nostalgist134/FuzzGIU/components/common"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"math/rand/v2"
	"time"
)

// DoRequestWs 发送WebSocket请求，支持超时、重试、响应解析
// 参数说明：
//   - req: 包含WebSocket目标URL、发送数据等信息的请求结构体
//   - timeout: 单次操作超时时间（秒），覆盖拨号、读写阶段
//   - retry: 最大重试次数（失败后再试N次，总请求数为N+1）
//   - retryRegex: 响应内容匹配该正则时触发重试
func DoRequestWs(req *fuzzTypes.Req, timeout int, retry int, retryRegex string) *fuzzTypes.Resp {
	resp := &fuzzTypes.Resp{
		ErrMsg: "no requests sent", // 初始状态
	}
	maxAttempts := retry + 1 // 总尝试次数 = 重试次数 + 1次初始请求

	// 配置WebSocket拨号器（复用默认配置，补充超时）
	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = time.Duration(timeout) * time.Second // 拨号+握手超时

	for attempt := 0; attempt < maxAttempts; attempt++ {
		// 记录本次尝试的开始时间（用于计算单次响应时间）
		attemptStart := time.Now()

		// 1. 建立WebSocket连接
		conn, _, err := dialer.Dial(req.URL, nil)
		if err != nil {
			resp.ErrMsg = fmt.Sprintf("conn failed(%d try): %v", attempt+1, err)
			// 非临时错误不重试（如URL无效、协议不支持）
			if !isTemporaryError(err) {
				break
			}
			// 重试间隔（50-300ms随机，避免高频请求）
			time.Sleep(time.Duration(rand.IntN(250)+50) * time.Millisecond)
			continue
		}

		// 2. 设置读写超时（本次连接的所有操作受限于用户指定的timeout）
		remainingTimeout := time.Duration(timeout)*time.Second - time.Since(attemptStart)
		if remainingTimeout <= 0 {
			conn.Close()
			resp.ErrMsg = fmt.Sprintf("timeout(%d try)", attempt+1)
			continue
		}
		deadline := time.Now().Add(remainingTimeout)
		conn.SetReadDeadline(deadline)
		conn.SetWriteDeadline(deadline)

		// 3. 发送数据
		err = conn.WriteMessage(websocket.TextMessage, req.Data)
		if err != nil {
			conn.Close() // 确保关闭连接再重试
			resp.ErrMsg = fmt.Sprintf("failed to send(%d try): %v", attempt+1, err)
			if !isTemporaryError(err) {
				break
			}
			time.Sleep(time.Duration(rand.IntN(500)+500) * time.Millisecond)
			continue
		}

		// 4. 读取响应
		_, message, err := conn.ReadMessage()
		// 无论读取成功与否，都关闭连接（WebSocket为短连接使用，不复用）
		safeCloseConn(conn)

		if err != nil {
			resp.ErrMsg = fmt.Sprintf("failed to read response(%d try): %v", attempt+1, err)
			if !isTemporaryError(err) {
				break
			}
			time.Sleep(time.Duration(rand.IntN(250)+50) * time.Millisecond)
			continue
		}

		// 5. 解析响应并检查是否需要重试
		resp.RawResponse = message
		resp.Size = len(message)
		resp.Words = len(bytes.Fields(message))
		resp.Lines = bytes.Count(message, []byte{'\n'})
		if len(message) > 0 && message[len(message)-1] != '\n' {
			resp.Lines++
		}

		// 6. 判断是否满足重试条件（正则匹配）
		if common.RegexMatch(message, retryRegex) {
			time.Sleep(time.Duration(rand.IntN(250)+50) * time.Millisecond)
			continue
		}

		// 7. 成功：记录本次尝试的响应时间，清空错误信息
		resp.ResponseTime = time.Since(attemptStart)
		resp.ErrMsg = ""
		break // 成功则退出循环
	}

	return resp
}

// isTemporaryError 判断错误是否为临时错误（适合重试）
func isTemporaryError(err error) bool {
	// 检查是否为gorilla/websocket包的错误
	var wsErr *websocket.CloseError
	if errors.As(err, &wsErr) {
		// 临时关闭错误（如服务器暂时过载）
		return wsErr.Code == websocket.CloseTryAgainLater
	}
	// 检查是否为临时网络错误（如超时、连接重置）
	if tempErr, ok := err.(interface{ Temporary() bool }); ok {
		return tempErr.Temporary()
	}
	if netErr, ok := err.(interface{ Timeout() bool }); ok {
		return netErr.Timeout()
	}
	// 其他错误默认不重试
	return false
}

// safeCloseConn 安全关闭WebSocket连接（核心修复：发送规范关闭帧）
// 遵循RFC 6455标准：先发送CloseMessage，再等待对方响应，最后关闭连接
func safeCloseConn(conn *websocket.Conn) error {
	if conn == nil {
		return nil
	}

	// 1. 发送关闭帧（1000=正常关闭，描述信息可选）
	closeMsg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "request completed")
	if err := conn.WriteMessage(websocket.CloseMessage, closeMsg); err != nil {
		// 发送关闭帧失败仍继续关闭连接，避免资源泄漏
		return fmt.Errorf("send close frame failed: %w", err)
	}

	// 2. 读取对方的关闭响应（设置100ms短超时，避免阻塞）
	conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	if _, _, err := conn.ReadMessage(); err != nil {
		// 忽略读取关闭响应的错误（对方可能已主动关闭连接）
	}

	// 3. 最终关闭底层连接
	return conn.Close()
}
