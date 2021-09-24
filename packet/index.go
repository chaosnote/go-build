package packet

import (
	"fmt"

	"log"

	"github.com/shopspring/decimal"
)

// 封包處理機制、會依專案有不同做法、思考後，還是另開共用區比較好

//-------------------------------------------------------------------------------------------------

// MaxMessageSize ...
const MaxMessageSize = 1024

// Key ...
var (
	Empty = Bytes([]byte{})
	OK    = Bytes([]byte("OK"))
)

//-------------------------------------------------------------------------------------------------[Packet]

// Packet ...
type Packet struct {
	// Action 操作行為 長度固定為 2
	// 例 : [0,0] OK
	Action Bytes
	// Key 相同行為下，用來區分差異，值，長度不限
	Key Bytes
	// Content ...
	Content []byte
}

// // Action ...
// func (v Packet) Action() string {
// 	dist := ""
// 	for _, b := range v[0:2] {
// 		dist = dist + decimal.NewFromInt(int64(b)).String()
// 	}
// 	return dist
// }

// // Key use for identify ( []int64 )
// func (v Packet) Key() string {

// 	// 0, 1, 2, 3, ...

// 	e := int(v[2])
// 	s := 3

// 	return Key(v[s : s+e])
// }

// // Content ...
// func (v Packet) Content() []byte {
// 	e := int(v[2])
// 	s := 3 + e
// 	return v[s:]
// }

//-------------------------------------------------------------------------------------------------

// Bytes ...
type Bytes []byte

// ToString ...
func (v Bytes) ToString() string {
	d := ""
	for _, b := range v {

		d = d + decimal.NewFromInt(int64(b)).String()
	}
	return d
}

// ToByte ...
func (v Bytes) ToByte() []byte {
	return []byte(v)
}

// Write ...
// action + key + content ( json format )
//
func Write(action Bytes, key Bytes, content []byte) Packet {
	b := []byte{}
	b = append(b, action.ToByte()...)
	b = append(b, byte(len(key))) // 記錄 Key 值長度
	b = append(b, key...)

	if content != nil {
		b = append(b, content...)
	}

	if len(b) > MaxMessageSize {
		log.Fatal(fmt.Sprintf("content length %d > 1024 !\n", len(b)))
	}

	return Packet{}
	// return Packet(b)
}

// Read 解析
// action [0:2] [底層]操作行為，不會回傳至客端
// key 取得 key 長度 [2], 從 [3] 取值至指定長度、遊戲狀態
// content
func Read(p []byte) Packet {
	s := 3             // key 值(起點位置)
	e := s + int(p[2]) // Key 長度( 0-255 )

	return Packet{
		Action:  Bytes(p[0:2]),
		Key:     Bytes(p[s:e]),
		Content: p[e:],
	}
}
