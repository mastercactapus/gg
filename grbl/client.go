package grbl

import (
	"bufio"
	"bytes"
	"errors"
	"io"
)

const grblBufSize = 50

type ClientMode int

const (
	ModeSendResponse ClientMode = iota
	ModeCharacterCount
)

type Client struct {
	rwc  io.ReadWriteCloser
	mode ClientMode

	closeCh          chan struct{}
	pushCh           chan []byte
	responseCh       chan []byte
	ioErrCh          chan error
	getMode, setMode chan ClientMode

	sendCh chan *clientRequest

	err error
}

type clientRequest struct {
	data  []byte
	resCh chan *Response
}
type Response struct {
	Data []byte
	Err  error
}

func NewClient(rwc io.ReadWriteCloser, mode ClientMode) *Client {
	c := &Client{
		rwc:        rwc,
		mode:       mode,
		getMode:    make(chan ClientMode),
		setMode:    make(chan ClientMode),
		closeCh:    make(chan struct{}),
		pushCh:     make(chan []byte, 100),
		responseCh: make(chan []byte, 10000),
		sendCh:     make(chan *clientRequest, 10000),
		ioErrCh:    make(chan error, 10),
	}

	go c.writeLoop()

	return c
}
func (c *Client) Close() error {
	close(c.closeCh)
	return c.rwc.Close()
}

// readLoop will constantly read lines from rwc
// ignoring blank lines.
func (c *Client) readLoop() {
	r := bufio.NewReader(c.rwc)
	var err error
	var b byte
	buf := make([]byte, 0, 1024)
	pushBuf := make([]byte, 0, 1024)
	var push byte
	for {
		b, err = r.ReadByte()
		if err != nil {
			c.ioErrCh <- err
			return
		}

		switch b {
		case '\r', ' ':
			continue
		case '<':
			push = '>'
		}
		if b != '\n' && b != push {
			if push != 0 {
				pushBuf = append(pushBuf, b)
			} else {
				buf = append(buf, b)
			}
		} else {
			if push != 0 {
				push = 0
				if len(pushBuf) == 0 {
					continue
				}
				data := make([]byte, len(pushBuf))
				copy(data, pushBuf)
				pushBuf = pushBuf[:0]
				c.responseCh <- data
			} else {
				if len(buf) == 0 {
					continue
				}
				data := make([]byte, len(buf))
				copy(data, buf)
				buf = buf[:0]
				c.responseCh <- data
			}
		}
	}
}

func (c *Client) writeLoop() {
	go c.readLoop()
	var req *clientRequest
	var data []byte

	var grblBuf []int
	var sendBuf [][]byte
	var resBuf []chan *Response

	sendOne := func() (n int) {
		n, c.err = c.rwc.Write(sendBuf[0])
		if c.err != nil {
			return n
		}
		sendBuf = sendBuf[1:]
		grblBuf = append(grblBuf, n)
		return n
	}

	fillGrbl := func() {
		if len(sendBuf) == 0 {
			return
		}
		if c.mode == ModeSendResponse {
			if len(grblBuf) != 0 {
				return
			}
			sendOne()
			return
		}

		var s int
		for _, n := range grblBuf {
			s += n
		}
		for c.err == nil && len(sendBuf) > 0 && s+len(sendBuf[0]) <= grblBufSize {
			s += sendOne()
		}
	}

	for {
		select {
		case c.getMode <- c.mode:
		case c.mode = <-c.setMode:
		case <-c.closeCh:
			return
		case req = <-c.sendCh:
			if isRealtimeCommand(req.data) {
				// write immediately, and respond after
				_, c.err = c.rwc.Write(req.data)
				if c.err != nil {
					req.resCh <- &Response{Err: c.err}
					c.errMode()
					return
				}
				req.resCh <- &Response{}
				continue
			}

			sendBuf = append(sendBuf, req.data)
			resBuf = append(resBuf, req.resCh)
			fillGrbl()
		case data = <-c.responseCh:
			if len(resBuf) > 0 && (data[0] == 'o' || data[0] == 'e') {
				resBuf[0] <- &Response{Data: data}
				grblBuf = grblBuf[1:]
				resBuf = resBuf[1:]
				fillGrbl()
			} else { //push messages
				if bytes.HasPrefix(data, []byte("Grbl")) {
					for i := 0; i < len(grblBuf); i++ {
						resBuf[i] <- &Response{Err: errors.New("soft reset")}
					}
					grblBuf = grblBuf[:0]
					resBuf = resBuf[:0]

					// don't send any commands that were pending pre-reset
					sendBuf = sendBuf[:0]
				}
				c.pushCh <- data
			}
		case c.err = <-c.ioErrCh:
			for _, r := range resBuf {
				r <- &Response{Err: c.err}
			}
			resBuf = resBuf[:0]
		}
		if c.err != nil {
			c.errMode()
			return
		}
	}
}

func (c *Client) errMode() {
	var req *clientRequest
	for {
		select {
		case <-c.closeCh:
			return
		case req = <-c.sendCh:
			req.resCh <- &Response{Err: c.err}
		}
	}
}

func (c *Client) Mode() ClientMode {
	return <-c.getMode
}
func (c *Client) SetMode(m ClientMode) {
	c.setMode <- m
}

func (c *Client) Execute(command []byte) *Response {
	ch := make(chan *Response, 1)
	c.sendCh <- &clientRequest{
		data:  command,
		resCh: ch,
	}

	return <-ch
}
func (c *Client) ExecuteMany(commands [][]byte) chan *Response {
	resCh := make(chan *Response, len(commands))
	ch := make(chan *Response, len(commands))
	for _, data := range commands {
		c.sendCh <- &clientRequest{data: data, resCh: ch}
	}

	go func() {
		max := len(commands)
		for i := 0; i < max; i++ {
			resCh <- <-ch
		}
		close(resCh)
	}()

	return resCh
}

func (c *Client) PushMessages() chan []byte {
	return c.pushCh
}

func isRealtimeCommand(data []byte) bool {
	if len(data) != 1 {
		return false
	}

	switch data[0] {
	case 0x18, '?', '~', '!':
		return true
	}

	// Bytes above this, while not all are actually realtime commands,
	// will be thrown away by Grbl. That means they all behave the same
	// from the perspective of the streaming protocol, and this Client.
	return data[0] > 0x7f
}
