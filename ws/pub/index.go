package pub

import (
	"net/url"
	"time"

	"github.com/chaosnote/go-build/internal"
	"github.com/chaosnote/go-build/ws"
	"github.com/chaosnote/go-kernel/net/conn"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

/*
publish 單一連線
*/

//-------------------------------------------------------------------------------------------------

const (
	sleep = 1 * time.Second
)

var (
	transfer *ws.Transfer = nil
)

//-------------------------------------------------------------------------------------------------

type handler struct {
}

func (v handler) Read(id string, msg []byte) {
	internal.Console("read", zap.String("id", id), zap.String("msg", string(msg)))
}

func (v handler) Close(id string) {
	internal.Console("close", zap.String("id", id))
	dial()
}

func (v handler) Error(id string, e interface{}) {
	internal.File(id, zap.Any("error", e))
}

//-------------------------------------------------------------------------------------------------

func dial() {

loop:

	e := transfer.Dial()
	if e != nil {
		time.Sleep(sleep)
		goto loop
	}

	go ws.R(*transfer)
	go ws.W(*transfer)

}

//-------------------------------------------------------------------------------------------------

/*
Build

*/
func Build(uri url.URL) {

	transfer = &ws.Transfer{
		Param: ws.Param{
			ID:          "push",
			WS:          conn.WebSocket(uri),
			MessageType: websocket.BinaryMessage,
		},
		Send:    make(chan []byte),
		Handler: handler{},
	}

	dial()

}

/*
Push
*/
func Push(msg []byte) {
	transfer.Send <- msg
}
