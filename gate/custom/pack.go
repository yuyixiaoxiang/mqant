package customPack

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
)

//定义包体格式
type Pack struct {
	MsgId   uint16
	MsgSize uint16
	MsgBody []byte
}

//包路由映射
var mapPackRouter map[uint16]string

//注册映射关系
func RegisterPackRouter(msgId uint16, router string) {
	if mapPackRouter == nil {
		mapPackRouter = make(map[uint16]string, 100)
	}
	mapPackRouter[msgId] = router
}

func ReadPack(r *bufio.Reader) (pack *Pack, err error) {
	fmt.Println("begin read byte")
	msgId, err := readInt16(r)
	if err != nil {
		return
	}
	msgSize, err := readInt16(r)
	if err != nil {
		return
	}
	msgBody := make([]byte, msgSize)
	_, err = io.ReadFull(r, msgBody)
	if err != nil {
		return
	}
	pack = &Pack{
		MsgId:   msgId,
		MsgSize: msgSize,
		MsgBody: msgBody,
	}
	fmt.Println("finish read byte")
	return
}

func readInt16(r *bufio.Reader) (uint16, error) {
	buf := make([]byte, 2)
	_, err := io.ReadFull(r, buf[:2])
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(buf[:2]), nil
}

func WritePack(pack *Pack, w *bufio.Writer) error {
	if err := DelayWritePack(pack, w); err != nil {
		return err
	}
	return w.Flush()
}

func DelayWritePack(pack *Pack, w *bufio.Writer) (err error) {
	// Write the fixed header

	return
}
