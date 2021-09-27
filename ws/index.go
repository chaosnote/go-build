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
	mMaxMessageSize = 1024                 // 訊息最大長度
)

/*
HandlerRead 處理串流讀取

	id string
	content []byte

*/
type HandlerRead func(string, []byte)

/*
HandlerClose 處理串流關閉

	id string

*/
type HandlerClose func(string)

/*
HandlerError 處理串流錯誤

	id string
	err interface{} 無法定義 recover 資訊

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
	mClose HandlerClose
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

func (p *param) Close() {
	if p.mClose != nil {
		p.mClose(p.mID)
	}
}

func (p *param) Destory() {
	p.mRead = nil
	p.mError = nil
	p.mClose = nil

	p.mConn = nil

	close(p.mSend)
}

//-----------------------------------------------[private]

func r(p param) {

	defer func() {

		if e := recover(); e != nil {
			p.Error(e)
		}

		p.mConn.Close()
		p.Close()

		destory(p.mID)

	}()

	p.mConn.SetReadLimit(mMaxMessageSize)
	p.mConn.SetReadDeadline(time.Now().Add(mPongWait))
	p.mConn.SetPongHandler(func(string) error { p.mConn.SetReadDeadline(time.Now().Add(mPongWait)); return nil })

	for {

		_, msg, e := p.mConn.ReadMessage()

		if e != nil {

			if websocket.IsUnexpectedCloseError(e, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) { // 非預期連線錯誤
				p.Error(e)
			}

			return
		}

		p.Read(msg)

	}

}

func w(p param, msgType int) {

	t := time.NewTicker(mPingPeriod)

	defer func() {

		if e := recover(); e != nil {
			p.Error(e) // 由委派層決定流程
		}

		t.Stop()
		p.mConn.Close()

	}()

	for {
		select {

		case b, ok := <-p.mSend:

			p.mConn.SetWriteDeadline(time.Now().Add(mWriteWait))
			if !ok {
				p.mConn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			writer, e := p.mConn.NextWriter(msgType)
			if e != nil {
				return
			}

			writer.Write(b)
			if e := writer.Close(); e != nil {
				return
			}

		case <-t.C:
			p.mConn.SetWriteDeadline(time.Now().Add(mWriteWait))

			if e := p.mConn.WriteMessage(websocket.PingMessage, nil); e != nil { // (已知)心跳封包錯誤、不在處理
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
destory
摧毀已註冊連線

	id
		外部指定

*/
func destory(id string) {
	mu.Lock()

	if p, ok := pool[id]; ok {
		p.Destory()
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
	close HandlerClose,
	err HandlerError,
) {
	c := param{
		mID:    id,
		mConn:  conn,
		mRead:  read,
		mClose: close,
		mError: err,
		mSend:  make(chan []byte),
	}

	// 註冊連線
	mu.Lock()
	pool[id] = c
	mu.Unlock()

	go r(c)
	go w(c, msgType)
}
