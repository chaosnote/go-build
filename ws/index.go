package ws

import (
	"fmt"
	"sync"
	"time"

	"github.com/chaosnote/go-build/internal"
	"github.com/chaosnote/go-kernel/net/conn"
	"go.uber.org/zap"

	"github.com/gorilla/websocket"
)

// 設定(常數)
const (
	mReadWait       = 10 * time.Second     // (等待)讀取時間
	mWriteWait      = 10 * time.Second     // (等待)寫入時間
	mPongWait       = 60 * time.Second     // (等待)回應時間
	mPingPeriod     = (mPongWait * 9) / 10 // 偵測時間
	mMaxMessageSize = 1024 * 5             // 訊息最大長度
)

//----------------------------------------------------------------------------------------------

/*
Handler (讀取/錯誤)處理
*/
type Handler interface {
	Read(string, []byte)
	Close(string)
	Error(string, interface{})
}

/*
Param 參數記錄
*/
type Param struct {
	ID          string // or uid
	WS          conn.WebSocket
	MessageType int
}

//----------------------------------------------------------------------------------------------

/*
Transfer

	例 :
	var _transfer = ws.Transfer{
		Param: ws.Param{
			ID:          id,
			WS:          conn.WebSocket(uri),
			MessageType: websocket.BinaryMessage,
		},
		Send:    make(chan []byte),
		Handler: subscribe{},
	}

	_transfer.Conn
	或
	_transfer.Dial()

*/
type Transfer struct {
	Param
	Handler

	Conn *websocket.Conn

	Send chan []byte // 當傳送值為 nil、會觸發關閉連線
}

func (v *Transfer) Dial() error {
	var e error
	v.Conn, e = v.Param.WS.Dial()
	if e != nil {
		return e
	}
	return nil
}

func (v *Transfer) Destory() {
	v.Handler = nil
}

//----------------------------------------------------------------------------------------------

type Group struct {
	mMu   sync.Mutex
	mPool map[string]*Transfer
}

/*
Send
指定傳送

	id string
	msg []byte

*/
func (v *Group) Send(id string, msg []byte) {
	v.mMu.Lock()
	defer v.mMu.Unlock()

	if p, ok := v.mPool[id]; ok {
		p.Send <- msg
	} else {
		internal.File("send", zap.Error(fmt.Errorf("socket lost id >> %s", id)))
	}
}

/*
Broadcast
擴播

	msg []byte
*/
func (v *Group) Broadcast(msg []byte) {
	v.mMu.Lock()
	defer v.mMu.Unlock()

	for _, c := range v.mPool {
		c.Send <- msg
	}
}

/*
Add
*/
func (v *Group) Add(c *Transfer) {
	v.mMu.Lock()
	defer v.mMu.Unlock()

	v.mPool[c.ID] = c
}

/*
Get
*/
func (v *Group) Get(id string) (*Transfer, bool) {
	v.mMu.Lock()
	defer v.mMu.Unlock()

	_transfer, ok := v.mPool[id]
	return _transfer, ok
}

/*
Remove
(移除)已註冊連線 (正常狀態下關閉連線)

	id string Param ID

*/
func (v *Group) Remove(id string) {
	v.mMu.Lock()
	defer v.mMu.Unlock()

	if p, ok := v.mPool[id]; ok {
		p.Destory()
		delete(v.mPool, id)
	}
}

/*
Close
(關閉)已註冊連線、非正常情形下、強迫踢線

	id string Param ID

*/
func (v *Group) Close(id string) {
	v.mMu.Lock()
	defer v.mMu.Unlock()

	if p, ok := v.mPool[id]; ok {
		p.Conn.WriteControl(websocket.CloseMessage, []byte("bye"), time.Now().Add(mWriteWait))
		p.Conn.Close()
		p.Destory()
	}
}

/*
CloseAll
*/
func (v *Group) CloseAll() {
	v.mMu.Lock()
	defer v.mMu.Unlock()

	for _, p := range v.mPool {
		p.Conn.WriteControl(websocket.CloseMessage, []byte("bye"), time.Now().Add(mWriteWait))
		p.Conn.Close()
		p.Destory()
	}
}

//----------------------------------------------------------------------------------------------

func R(c Transfer) {

	defer func() {

		if e := recover(); e != nil {

			if c.Handler != nil {
				c.Error(c.ID, e)
			}

		}

		c.Conn.Close()

		if c.Handler != nil {
			c.Close(c.ID)
		}

	}()

	c.Conn.SetReadLimit(mMaxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(mPongWait))
	c.Conn.SetPongHandler(func(string) error { c.Conn.SetReadDeadline(time.Now().Add(mPongWait)); return nil })

	for {

		_, msg, e := c.Conn.ReadMessage()

		if e != nil {

			if websocket.IsUnexpectedCloseError(e, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) { // 非預期連線錯誤

				if c.Handler != nil {
					c.Error(c.ID, e)
				}

			}

			return
		}

		if c.Handler != nil {
			c.Read(c.ID, msg)
		}

	}

}

func W(c Transfer) {

	t := time.NewTicker(mPingPeriod)

	defer func() {

		if e := recover(); e != nil {

			if c.Handler != nil {
				c.Error(c.ID, e)
			}
		}

		t.Stop()
		c.Conn.Close()

		// close(c.Send) // 關閉狀態下、自動重連( panic: send on closed channel )，每次進入重建 ? 思考

	}()

	for {
		select {

		case b, ok := <-c.Send:
			if len(b) == 0 {
				c.Conn.Close()
				return
			}
			c.Conn.SetWriteDeadline(time.Now().Add(mWriteWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			writer, e := c.Conn.NextWriter(c.MessageType)
			if e != nil {
				return
			}
			writer.Write(b)
			if e := writer.Close(); e != nil {
				return
			}
		case <-t.C:

			c.Conn.SetWriteDeadline(time.Now().Add(mWriteWait))

			if e := c.Conn.WriteMessage(websocket.PingMessage, nil); e != nil { // (已知)心跳封包錯誤、不在處理
				return
			}

		}

	}

}

//----------------------------------------------------------------------------------------------

func New() *Group {
	return &Group{
		mPool: map[string]*Transfer{},
	}
}
