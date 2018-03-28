// Copyright 2014 mqant Author. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package customPack

import (
	"bufio"
	"fmt"
	"time"

	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/network"
)

// Tcp write queue
type PackQueue struct {
	conf conf.Mqtt
	// The last error in the tcp connection
	writeError error
	// Notice read the error
	errorChan chan error
	noticeFin chan byte
	writeChan chan *packAndType
	readChan  chan<- *Pack
	// Pack connection
	r *bufio.Reader
	w *bufio.Writer

	conn network.Conn

	alive int
}

type packAndErr struct {
	pack *Pack
	err  error
}

// 1 is delay, 0 is no delay, 2 is just flush.
const (
	NO_DELAY = iota
	DELAY
	FLUSH
)

type packAndType struct {
	msgId uint16
	bytes []byte
}

// Init a pack queue
func NewPackQueue(conf conf.Mqtt, r *bufio.Reader, w *bufio.Writer, conn network.Conn, readChan chan *Pack, alive int) *PackQueue {
	if alive < 1 {
		alive = conf.ReadTimeout
	}
	alive = int(float32(alive)*1.5 + 1)
	return &PackQueue{
		conf:      conf,
		alive:     alive,
		r:         r,
		w:         w,
		conn:      conn,
		noticeFin: make(chan byte, 2),
		writeChan: make(chan *packAndType, conf.WirteLoopChanNum),
		readChan:  readChan,
		errorChan: make(chan error, 1),
	}
}

// Start a pack write queue
// It should run in a new grountine
func (queue *PackQueue) writeLoop() {
	// defer recover()
	var err error
loop:
	for {
		select {
		case pt, ok := <-queue.writeChan:
			if !ok {
				break loop
			}
			if pt == nil {
				break loop
			}
			if queue.conf.WriteTimeout > 0 {
				queue.conn.SetWriteDeadline(time.Now().Add(time.Second * time.Duration(queue.conf.WriteTimeout)))
			}
			err = WritePack(pt, queue.w)

			if err != nil {
				// Tell listener the error
				// Notice the read
				queue.writeError = err
				queue.errorChan <- err
				queue.noticeFin <- 0
				break loop
			}
		}
	}
}

// Write a pack , and get the last error
func (queue *PackQueue) WritePack(msgId uint16, bytes []byte) error {
	if queue.writeError != nil {
		return queue.writeError
	}
	queue.writeChan <- &packAndType{
		msgId: msgId,
		bytes: bytes,
	}
	return nil
}

func (queue *PackQueue) SetAlive(alive int) error {
	if alive < 1 {
		alive = queue.conf.ReadTimeout
	}
	alive = int(float32(alive)*1.5 + 1)
	queue.alive = alive
	return nil
}

// Get a read pack queue
// Only call once
func (queue *PackQueue) ReadPackInLoop() {
	go func() {
		// defer recover()
		is_continue := true
	loop:
		for {
			if queue.alive > 0 {
				queue.conn.SetReadDeadline(time.Now().Add(time.Second * time.Duration(int(float64(queue.alive)*1.5))))
			} else {
				queue.conn.SetReadDeadline(time.Now().Add(time.Second * 90))
			}
			if is_continue {
				pack, err := ReadPack(queue.r)
				if err != nil {
					is_continue = false
				}
				if pack == nil {
					continue
				}

				select {
				case queue.readChan <- pack:
					break
					// Without anything to do
				case <-queue.noticeFin:
					//queue.Close()
					log.Info("Queue FIN")
					break loop
				}
			} else {
				<-queue.noticeFin
				log.Info("Queue not continue")
				break loop
			}
		}
		queue.Close()
	}()
}

// Close the all of queue's channels
func (queue *PackQueue) Close() error {
	close(queue.writeChan)
	close(queue.readChan)
	close(queue.errorChan)
	close(queue.noticeFin)
	return nil
}

// Buffer
type buffer struct {
	index int
	data  []byte
}

func newBuffer(data []byte) *buffer {
	return &buffer{
		data:  data,
		index: 0,
	}
}
func (b *buffer) readString(length int) (s string, err error) {
	if (length + b.index) > len(b.data) {
		err = fmt.Errorf("Out of range error:%v", length)
		return
	}
	s = string(b.data[b.index:(length + b.index)])
	b.index += length
	return
}
func (b *buffer) readByte() (c byte, err error) {
	if (1 + b.index) > len(b.data) {
		err = fmt.Errorf("Out of range error")
		return
	}
	c = b.data[b.index]
	b.index++
	return
}
