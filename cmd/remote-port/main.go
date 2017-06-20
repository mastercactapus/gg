package main

import (
	"flag"
	"io"
	"log"
	"net"

	serial "go.bug.st/serial.v1"
)

var (
	port = flag.String("port", "", "Serial port to use.")
	rate = flag.Int("b", 115200, "Baudrate of the serial port.")
	addr = flag.String("listen", ":7456", "Listen address.")
)

func forward(w io.WriteCloser, r io.ReadCloser) {
	io.Copy(w, r)
	w.Close()
	r.Close()
}

func main() {
	flag.Parse()
	l, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalln("failed to start server:", err)
	}

	var prev, c, p io.ReadWriteCloser

	var opened bool
	for {
		c, err = l.Accept()
		if opened {
			prev.Close()
			p.Close()
		}
		if err != nil {
			break
		}
		prev = c

		p, err = serial.Open(*port, &serial.Mode{BaudRate: *rate})
		if err != nil {
			log.Printf("failed to open serial port: %v", err)
			c.Close()
			continue
		}

		go forward(p, c)
		go forward(c, p)
	}
}
