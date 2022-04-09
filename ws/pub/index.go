package pub

import (
	"net/url"
	"sync"
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
	sleep = internal.Delag * time.Second
)

var (
	transfer *ws.Transfer = nil
)

var (
	mu   sync.Mutex
	flag bool
)

func set(b bool) {
	mu.Lock()
	defer mu.Unlock()

	flag = b
}

func Connected() bool {
	mu.Lock()
	defer mu.Unlock()

	return flag
}

//-------------------------------------------------------------------------------------------------

type handler struct {
}

func (v handler) Read(id string, msg []byte) {
	internal.Console("pub", zap.String("read", id), zap.String("msg", string(msg)))
}

func (v handler) Close(id string) {
	internal.Console("pub", zap.String("close", id))
	set(false)
	dial(false)
}

func (v handler) Error(id string, e interface{}) {
	internal.File("pub", zap.Any("error", e))
}

//-------------------------------------------------------------------------------------------------

func dial(b bool) {

loop:

	e := transfer.Dial()
	if e != nil {
		internal.File("pub", zap.Error(e))
		if b {
			panic(e)
		}
		time.Sleep(sleep)
		goto loop
	}

	go ws.R(*transfer)
	go ws.W(*transfer)

	set(true)
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

	dial(true)

}

/*
Push
*/
func Push(msg []byte) {
	transfer.Send <- msg
}
