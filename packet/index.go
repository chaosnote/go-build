package packet

import (
	"fmt"

	"log"
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

//-------------------------------------------------------------------------------------------------[Bytes]

/*
Bytes
*/
type Bytes []byte

/*
ToString
另一種做法，將 byte 轉為 int
*/
// func (v Bytes) ToString() string {
// 	d := ""
// 	for _, b := range v {

// 		d = d + decimal.NewFromInt(int64(b)).String()
// 	}
// 	return d
// }

/*
ToString
*/
func (v Bytes) ToString() string {
	return string(v)
}

/*
ToByte
*/
func (v Bytes) ToByte() []byte {
	return []byte(v)
}

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
		log.Fatal(fmt.Sprintf("content length %d > 1024 !\n", len(b)))
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
