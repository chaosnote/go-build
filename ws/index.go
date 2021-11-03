package ws

import (
	"log"
	"sync"
	"time"

	"github.com/chaosnote/go-build/packet"
	"github.com/chaosnote/go-kernel/evt"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// 設定(常數)
const (
	mReadWait       = 10 * time.Second     // (等待)讀取時間
	mWriteWait      = 10 * time.Second     // (等待)寫入時間
	mPongWait       = 60 * time.Second     // (等待)回應時間
	mPingPeriod     = (mPongWait * 9) / 10 // 偵測時間
	mMaxMessageSize = 1024                 // 訊息最大長度
)

const (
	READ  string = "read"
	ERROR string = "error"
	CLOSE string = "close"
)

//----------------------------------------------------------------------------------------------

/*
Param ...
*/
type Param struct {
	ID          string // or uid
	Conn        *websocket.Conn
	MessageType int
	Event       evt.IEvent
	Send        chan []byte
}

func (p *Param) Destory() {
	p.Conn = nil
	close(p.Send)
}

//----------------------------------------------------------------------------------------------

type Group struct {
	mMu   sync.Mutex
	mPool map[string]Param
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
		log.Println("socket lost id >>", id)
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

	for _, conn := range v.mPool {
		conn.Send <- msg
	}
}

/*
Add
串流處理行為( 改由外部設置 )
	委派
		讀
		錯誤
	併發
		寫
*/
func (v *Group) Add(p Param) {
	v.mMu.Lock()
	defer v.mMu.Unlock()

	v.mPool[p.ID] = p
}

/*
Remove
(移除/摧毀)已註冊連線

	id string
		外部指定

*/
func (v *Group) Remove(id string) {
	v.mMu.Lock()
	defer v.mMu.Unlock()

	if p, ok := v.mPool[id]; ok {
		p.Destory()
		delete(v.mPool, id)
	}
}

//----------------------------------------------------------------------------------------------

func R(p Param) {

	defer func() {

		if e := recover(); e != nil {
			if p.Event != nil {

				p.Event.Dispatch(
					ERROR,
					packet.Write(
						packet.Bytes([]byte(p.ID)),
						[]byte(zap.Any("Read", e).String),
					),
				)

			}
		}

		p.Conn.Close()

		if p.Event != nil {

			p.Event.Dispatch(
				CLOSE,
				packet.Write(
					packet.Bytes([]byte(p.ID)),
					[]byte{},
				),
			)

		}

	}()

	p.Conn.SetReadLimit(mMaxMessageSize)
	p.Conn.SetReadDeadline(time.Now().Add(mPongWait))
	p.Conn.SetPongHandler(func(string) error { p.Conn.SetReadDeadline(time.Now().Add(mPongWait)); return nil })

	for {

		_, msg, e := p.Conn.ReadMessage()

		if e != nil {

			if websocket.IsUnexpectedCloseError(e, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) { // 非預期連線錯誤

				if p.Event != nil {

					p.Event.Dispatch(
						ERROR,
						packet.Write(
							packet.Bytes([]byte(p.ID)),
							[]byte(zap.Any("Read", e).String),
						),
					)

				}

			}

			return
		}

		if p.Event != nil {

			p.Event.Dispatch(
				READ,
				packet.Write(
					packet.Bytes([]byte(p.ID)),
					msg,
				),
			)

		}

	}

}

func W(p Param) {

	t := time.NewTicker(mPingPeriod)

	defer func() {

		if e := recover(); e != nil {

			if p.Event != nil {

				p.Event.Dispatch(
					ERROR,
					packet.Write(
						packet.Bytes([]byte(p.ID)),
						[]byte(zap.Any("Write", e).String),
					),
				)

			}

		}

		t.Stop()
		p.Conn.Close()

	}()

	for {
		select {

		case b, ok := <-p.Send:

			p.Conn.SetWriteDeadline(time.Now().Add(mWriteWait))
			if !ok {
				p.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			writer, e := p.Conn.NextWriter(p.MessageType)
			if e != nil {
				return
			}

			writer.Write(b)
			if e := writer.Close(); e != nil {
				return
			}

		case <-t.C:
			p.Conn.SetWriteDeadline(time.Now().Add(mWriteWait))

			if e := p.Conn.WriteMessage(websocket.PingMessage, nil); e != nil { // (已知)心跳封包錯誤、不在處理
				return
			}

		}

	}

}

//----------------------------------------------------------------------------------------------

func New() *Group {
	return &Group{
		mPool: map[string]Param{},
	}
}
