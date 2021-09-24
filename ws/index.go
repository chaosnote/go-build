package ws

import (
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// 設定(常數)
const (
	mReadWait       = 10 * time.Second     // (等待)讀取時間
	mWriteWait      = 10 * time.Second     // (等待)寫入時間
	mPongWait       = 60 * time.Second     // (等待)回應時間
	mPingPeriod     = (mPongWait * 9) / 10 // 偵測時間
	mMaxMessageSize = 1024
)

/*
HandlerRead 處理串流讀取

	id string
	content []byte

*/
type HandlerRead func(string, []byte)

/*
HandlerError 處理串流錯誤

	id string
	err interface{} recover 的資訊無法定義

*/
type HandlerError func(string, interface{})

//----------------------------------------------------------------------------------------------

/*
param ...
*/
type param struct {
	mID    string // or uid
	mConn  *websocket.Conn
	mSend  chan []byte
	mRead  HandlerRead
	mError HandlerError
}

func (p *param) Read(msg []byte) {
	if p.mRead != nil {
		p.mRead(p.mID, msg)
	}
}

func (p *param) Error(e interface{}) {
	if p.mError != nil {
		p.mError(p.mID, e)
	}
}

//-----------------------------------------------[private]

func r(p param) {

	defer func() {
		if e := recover(); e != nil {
			p.Error(e)
		}
	}()

	p.mConn.SetReadLimit(mMaxMessageSize)
	p.mConn.SetReadDeadline(time.Now().Add(mPongWait))
	p.mConn.SetPongHandler(func(string) error { p.mConn.SetReadDeadline(time.Now().Add(mPongWait)); return nil })

	for {

		_, msg, e := p.mConn.ReadMessage()

		if e != nil {

			if websocket.IsUnexpectedCloseError(e, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				p.Error(e)
			} else {
				p.Error(nil)
			}

			return
		}

		p.Read(msg)

	}

}

func w(p param, msgType int) {
	t := time.NewTicker(mPingPeriod)

	defer func() {

		t.Stop()

		if e := recover(); e != nil {
			p.Error(e) // 由委派層決定流程
		}

	}()

	for {
		select {

		case b, ok := <-p.mSend:

			p.mConn.SetWriteDeadline(time.Now().Add(mWriteWait))
			if !ok {
				p.mConn.WriteMessage(websocket.CloseMessage, []byte{})
				p.Error(nil)
				return
			}

			writer, e := p.mConn.NextWriter(msgType)
			if e != nil {
				p.Error(e)
				return
			}

			writer.Write(b)
			if e := writer.Close(); e != nil {
				p.Error(e)
				return
			}

		case <-t.C:
			p.mConn.SetWriteDeadline(time.Now().Add(mWriteWait))

			if e := p.mConn.WriteMessage(websocket.PingMessage, nil); e != nil { // 心跳封包錯誤
				// (需留意)連線並非即時關閉
				// msg, _ := url.QueryUnescape()
				// "write tcp 127.0.0.1:10001->127.0.0.1:30061: use of closed network connection"
				p.Error(nil)
				return
			}

		}

	}

}

//----------------------------------------------------------------------------------------------

var mu sync.Mutex
var pool = map[string]param{}

func Send(id string, msg []byte) {
	mu.Lock()
	if p, ok := pool[id]; ok {
		p.mSend <- msg
	} else {
		log.Println("socket lost id >>", id)
	}
	mu.Unlock()
}

/*
Destory
摧毀指定(連線)

	id
		外部指定

*/
func Destory(id string) {
	mu.Lock()

	if p, ok := pool[id]; ok {
		p.mRead = nil
		p.mError = nil

		close(p.mSend)

		p.mConn.Close() // 放置於 close 後，因為順序為關閉讀後，才觸發寫入 close message

		delete(pool, id)
	}

	mu.Unlock()
}

//----------------------------------------------------------------------------------------------[build client]

/*
AddObserver
串流處理行為( 改由外部設置 )
	委派
		讀
		錯誤
	併發
		寫
*/
func AddObserver(
	id string,
	msgType int,
	conn *websocket.Conn,
	read HandlerRead,
	err HandlerError,
) {
	c := param{
		mID:    id,
		mConn:  conn,
		mRead:  read,
		mError: err,
		mSend:  make(chan []byte),
	}

	mu.Lock()
	pool[id] = c
	mu.Unlock()

	go r(c)
	go w(c, msgType)
}
