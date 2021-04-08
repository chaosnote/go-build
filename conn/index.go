package conn

import (
	"time"

	"github.com/gorilla/websocket"
)

// socket 連線管理
// 連線(成功/失敗)
// log 記錄
// 訊息(接受/傳送)

// 設定(常數)
const (
	mReadWait       = 10 * time.Second     // (等待)讀取時間
	mWriteWait      = 10 * time.Second     // (等待)寫入時間
	mPongWait       = 60 * time.Second     // (等待)回應時間
	mPingPeriod     = (mPongWait * 9) / 10 // 偵測時間
	mMaxMessageSize = 1024
)

//----------------------------------------------------------------------------------------------

// Param ...
type Param struct {
	Addr    string
	Conn    *websocket.Conn
	ID      string // 可辨識 ID
	MsgType int
}

//----------------------------------------------------------------------------------------------

// Client 結構
type Client struct {
	mParam Param // 固定(包含)參數
	Send   chan []byte
}

//-----------------------------------------------[private]

func (v *Client) r() {
	conn := v.mParam.Conn
	// [TODO]修改回 callback // 告知啟動讀取
	// v.mObserver.Register <- v // trigger when init complete

	defer func() {

		// [TODO]修改回 callback // 告知關閉讀取

		if e := recover(); e != nil {
			// [TODO]修改回 callback // 關閉原因
		}
	}()

	conn.SetReadLimit(mMaxMessageSize)
	conn.SetReadDeadline(time.Now().Add(mPongWait))
	conn.SetPongHandler(func(string) error { conn.SetReadDeadline(time.Now().Add(mPongWait)); return nil })

	for {

		_, _, e := conn.ReadMessage() // msg

		if e != nil {

			if websocket.IsUnexpectedCloseError(e, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				// [TODO]修改回 callback // 告知錯誤原因
				// v.mLogger.Info(e.Error())
			}

			return
		}

		// [TODO]修改回 callback // 將讀取內容轉發至委派對像
		// v.mParam.SetContent(msg)
		// if v.mObserver != nil {
		// 	v.mObserver.Broadcast <- v.mParam
		// }

	}

}

func (v *Client) w() {
	t := time.NewTicker(mPingPeriod)
	conn := v.mParam.Conn

	// [TODO]修改回 callback // 告知啟動寫入

	defer func() {

		// [TODO]修改回 callback // 告知關閉寫入

		t.Stop()
		conn.Close()

		close(v.Send)

		if e := recover(); e != nil {
			// [TODO]修改回 callback // 告知錯誤原因
		}

	}()

	for {
		select {

		case b, ok := <-v.Send:

			conn.SetWriteDeadline(time.Now().Add(mWriteWait))
			if !ok {
				conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			writer, e := conn.NextWriter(v.mParam.MsgType)
			if e != nil {
				// [TODO]修改回 callback // 告知錯誤原因
				return
			}

			writer.Write(b)
			if e := writer.Close(); e != nil {
				// [TODO]修改回 callback // 告知錯誤原因
				return
			}

		case <-t.C:
			conn.SetWriteDeadline(time.Now().Add(mWriteWait))
			if e := conn.WriteMessage(websocket.PingMessage, nil); e != nil {
				// msg, _ := url.QueryUnescape()
				// "write tcp 127.0.0.1:10001->127.0.0.1:30061: use of closed network connection"
				// [TODO]修改回 callback // 告知錯誤原因
				return
			}
		}

	}

}

// Close 主動關閉連線
func (v *Client) Close() {
	v.mParam.Conn.Close()
}

// NewClient ...
func NewClient(
	p Param,
) {

	c := &Client{
		Send:   make(chan []byte),
		mParam: p,
	}

	go c.r()
	go c.w()

}
