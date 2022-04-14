package sub

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/chaosnote/go-build/internal"
	"github.com/chaosnote/go-build/ws"
	"github.com/chaosnote/go-kernel/net/conn"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

//-------------------------------------------------------------------------------------------------

const (
	sleep = internal.Delag * time.Second
)

var (
	ErrDisconnect = fmt.Errorf("sub disconnect")
)

var (
	group = ws.New()
)

//-------------------------------------------------------------------------------------------------

type Handler struct {
}

func (v Handler) Close(id string) {
	internal.File("sub", zap.String("close", id))
	dial(id)
}

func (v Handler) Error(id string, e interface{}) {
	internal.File("sub", zap.Any("error", e))
}

//-------------------------------------------------------------------------------------------------

func dial(id string) {
	t, ok := group.Get(id)
	if !ok {
		internal.File("sub", zap.String("dial", id))
		return
	}

	var f bool
loop:

	e := t.Dial()
	if e != nil {
		if !f {
			f = true
			internal.File("sub", zap.String("dial", id), zap.Error(e))
		}
		time.Sleep(sleep)
		goto loop
	}

	go ws.R(*t)
	go ws.W(*t)

	internal.File("sub", zap.String("dial", id))

}

//-------------------------------------------------------------------------------------------------

/*
Build
*/
func Build(id string, uri url.URL, handler ws.Handler) {

	transfer := &ws.Transfer{
		Param: ws.Param{
			ID:          id,
			WS:          conn.WebSocket(uri),
			MessageType: websocket.BinaryMessage,
		},
		Send:    make(chan []byte),
		Handler: handler,
	}

	internal.File("sub", zap.String("path", uri.String()))

	if strings.Count(uri.Path, "/") < 2 {
		internal.Fatal("sub", zap.Error(fmt.Errorf("error uri.Path ( %s )", uri.Path)))
	}

	e := transfer.Dial()
	if e != nil {
		internal.Fatal("sub", zap.Error(e))
	}

	group.Add(transfer)

	go ws.R(*transfer)
	go ws.W(*transfer)

}
