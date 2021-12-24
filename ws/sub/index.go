package sub

import (
	"net/url"
	"time"

	"github.com/chaosnote/go-build/internal"
	"github.com/chaosnote/go-build/ws"
	"github.com/chaosnote/go-kernel/net/conn"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

//-------------------------------------------------------------------------------------------------

const (
	sleep = 3 * time.Second
)

var (
	group = ws.New()
)

//-------------------------------------------------------------------------------------------------

type Handler struct {
}

func (v Handler) Close(id string) {
	internal.File("close", zap.String("id", id))
	dial(id)
}

func (v Handler) Error(id string, e interface{}) {
	internal.File(id, zap.Any("error", e))
}

//-------------------------------------------------------------------------------------------------

func dial(id string) {

	_transfer, ok := group.Get(id)
	if !ok {
		return
	}

loop:

	e := _transfer.Dial()
	if e != nil {
		time.Sleep(sleep)
		goto loop
	}

	go ws.R(*_transfer)
	go ws.W(*_transfer)

}

//-------------------------------------------------------------------------------------------------

/*
Build
*/
func Build(key string, uri url.URL, handler ws.Handler) {

	_transfer := &ws.Transfer{
		Param: ws.Param{
			ID:          key,
			WS:          conn.WebSocket(uri),
			MessageType: websocket.BinaryMessage,
		},
		Send:    make(chan []byte),
		Handler: handler,
	}

	e := _transfer.Dial()
	if e != nil {
		panic(e)
	}

	group.Add(_transfer)

	go ws.R(*_transfer)
	go ws.W(*_transfer)

}
