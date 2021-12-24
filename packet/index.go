package packet

import (
	"fmt"

	"github.com/chaosnote/go-build/internal"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

/*
封包處理機制、會依專案有不同做法
*/

//-------------------------------------------------------------------------------------------------

/*
MaxMessageSize
*/
const (
	MaxMessageSize = 1024
)

//-------------------------------------------------------------------------------------------------[Packet]

/*
Packet
*/
type Packet struct {
	Action  Bytes
	Content []byte
}

//-------------------------------------------------------------------------------------------------

/*
Trans
key 轉換為可讀性較高的字元
*/
func Trans(p []byte) []byte {
	s := ""
	for _, b := range p {
		s = s + decimal.NewFromInt(int64(b)).String()
	}
	return []byte(s)
}

//-------------------------------------------------------------------------------------------------[Bytes]

/*
Bytes
*/
type Bytes []byte

/*
Write
action + content

	action Bytes 例: uid
	content []byte

	return []byte
*/
func Write(action Bytes, content []byte) []byte {
	b := []byte{}
	b = append(b, byte(len(action))) // 記錄 identify 值長度
	b = append(b, action...)

	if content != nil {
		b = append(b, content...)
	}

	if len(b) > MaxMessageSize {
		internal.Fatal("packet", zap.Error(fmt.Errorf("content length %d > 1024 !\n", len(b))))
	}

	return b
}

/*
Read
key 取得 key 長度 [0], 從 [1] 取值至指定長度、遊戲狀態

	p []byte

	return Packet

*/
func Read(p []byte) Packet {
	s := 1             // key 值(起點位置)
	e := s + int(p[0]) // Key 長度( 0-255 )

	return Packet{
		Action:  Bytes(p[1:e]),
		Content: p[e:],
	}
}
