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
	"errors"
	"fmt"
	"math"
	"sync"

	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/network"
)

var notAlive = errors.New("Connection was dead")

type PackRecover interface {
	OnRecover(*Pack)
}

type Client struct {
	queue *PackQueue

	recover  PackRecover  //消息接收者,从上层接口传过来的 只接收正式消息(心跳包,回复包等都不要)
	readChan <-chan *Pack //读取底层接收到的所有数据包包

	closeChan   chan byte // Other gorountine Call notice exit
	isSendClose bool      // Wheather has a new login user.
	isLetClose  bool      // Wheather has relogin.

	isStop bool
	lock   *sync.Mutex

	// Online msg id
	curr_id int
}

func NewClient(conf conf.Mqtt, recover PackRecover, r *bufio.Reader, w *bufio.Writer, conn network.Conn, alive int) *Client {
	readChan := make(chan *Pack, conf.ReadPackLoop)
	return &Client{
		readChan:  readChan,
		queue:     NewPackQueue(conf, r, w, conn, readChan, alive),
		recover:   recover,
		closeChan: make(chan byte),
		lock:      new(sync.Mutex),
		curr_id:   0,
	}
}

// Push the msg and response the heart beat
func (c *Client) Listen_loop() (e error) {
	defer func() {
		if r := recover(); r != nil {
			if c.isSendClose {
				c.closeChan <- 0
			}
		}
	}()
	var (
		err error
		// wg        = new(sync.WaitGroup)
	)

	// Start the write queue
	go c.queue.writeLoop()

	c.queue.ReadPackInLoop()

	// Start push 读取数据包
	//pingtime := time.NewTimer(time.Second * time.Duration(int(float64(c.queue.alive)*1.5)))
loop:

	for {
		select {
		case pack, ok := <-c.readChan:
			if !ok {
				log.Info("Get a connection error")
				break loop
			}
			fmt.Println("pack client 中读到数据")
			//pingtime.Reset(time.Second * time.Duration(int(float64(c.queue.alive)*1.5))) //重置
			if err = c.waitPack(pack); err != nil {
				log.Info("Get a connection error , will break(%v)", err)
				break loop
			}
		//case <-pingtime.C:
		//	pingtime.Reset(time.Second * time.Duration(int(float64(c.queue.alive)*1.5)))
		//	c.timeout()
		case <-c.closeChan:
			c.waitQuit()
			break loop
		}
	}

	c.lock.Lock()
	//pingtime.Stop()
	c.isStop = true
	c.lock.Unlock()
	// Wrte the onlines msg to the db
	// Free resources
	// Close channels

	close(c.closeChan)
	log.Info("listen_loop Groutine will esc.")
	return
}

// Setting a mqtt pack's id.
func (c *Client) getOnlineMsgId() int {
	if c.curr_id == math.MaxUint16 {
		c.curr_id = 1
		return c.curr_id
	} else {
		c.curr_id = c.curr_id + 1
		return c.curr_id
	}
}
func (c *Client) timeout() (err error) {
	log.Info("timeout 主动关闭连接")
	return c.queue.conn.Close()
}
func (c *Client) waitPack(pack *Pack) (err error) {

	c.recover.OnRecover(pack)
	return
}

func (c *Client) waitQuit() {
	// Start close
	log.Info("Will break new relogin")
	c.isSendClose = true
}

func (c *Client) WriteMsg(msgId uint16, body []byte) error {
	if c.isStop {
		return fmt.Errorf("connection is closed")
	}
	err := c.queue.WritePack(msgId, body)
	return err
}
