package middleware

import (
	"bytes"
	"crypto/tls"
	"io"
	"log"
	"net"
)

// Handler middleware from client connection
type Handler struct {
	client    net.Conn
	socketIn  net.Conn
	socketOut net.Conn
	firstBuf  []byte
	encrypted bool
}

// Config is attr of handler middleware
type Config struct {
}

// Handle return the socket handled and error or nil
func Handle(conn net.Conn) (socket net.Conn, err error) {
	s := &Handler{
		client: conn,
	}
	s.configure()
	return
}

func (h *Handler) configure() {
	if ne, ee := h.extractBuf(); ee != nil || ne < 3 {
		if ee != nil {
			log.Fatalln(ee)
		}
		panic("No minimal bufer size")
	}
	h.socketIn, h.socketOut = net.Pipe()

	if h.encrypted = h.checkTLS(); h.encrypted {
		cer, err := tls.LoadX509KeyPair("cert.pem", "key.pem")
		if err != nil {
			log.Fatalln(err)
			return
		}
		config := &tls.Config{Certificates: []tls.Certificate{cer}}
		h.client = tls.Server(h.client, config)
	}
}

func (h *Handler) extractBuf() (readed int64, err error) {
	nr, er := h.client.Read(h.firstBuf)
	if nr > 0 {
		readed += int64(nr)
	}
	if er != nil {
		if er != io.EOF {
			err = er
		}
		return
	}
	return
}

func (h *Handler) checkTLS() bool {
	return bytes.Compare(h.firstBuf[0:3], []byte("\x16\x03\x01")) == 0
}
