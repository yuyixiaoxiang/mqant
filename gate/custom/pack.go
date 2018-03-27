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

type PackRouter struct {
	Id     uint16
	Module string
	Func   string
}

//包路由映射
var mapPackRouter map[uint16]*PackRouter

//注册映射关系
func RegisterPackRouter(msgId uint16, router *PackRouter) {
	if mapPackRouter == nil {
		mapPackRouter = make(map[uint16]*PackRouter, 100)
	}
	mapPackRouter[msgId] = router
}

func GetPackRouter(msgId uint16) *PackRouter {
	router := mapPackRouter[msgId]
	if router != nil {
		return router
	}
	return nil
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
func writeUInt16(w *bufio.Writer, v uint16) error {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, v)
	return writeFull(w, b)
}
func WritePack(pack *packAndType, w *bufio.Writer) error {
	msgId := pack.msgId
	msgSize := uint16(len(pack.bytes))
	writeUInt16(w, msgId)
	writeUInt16(w, msgSize)
	writeFull(w, pack.bytes)
	return w.Flush()
}

func writeFull(w *bufio.Writer, b []byte) (err error) {
	hasRead, n := 0, 0
	for n < len(b) {
		n, err = w.Write(b[hasRead:])
		if err != nil {
			break
		}
		hasRead += n
	}
	return err
}
