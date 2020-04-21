package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
)

func handlerMux(dst io.Writer, src io.Reader) (written int64, err error) {
	dataFull := make([]byte, (16 * 1024))

	nr, er := src.Read(dataFull)
	fmt.Println(string(dataFull[0:6]))
	if nr > 0 {
		nw, ew := dst.Write(dataFull[0:nr])
		if nw > 0 {
			written += int64(nw)
		}
		if ew != nil {
			err = ew
			return
		}
		if nr != nw {
			err = io.ErrShortWrite
			return
		}
	}
	if er != nil {
		if er != io.EOF {
			err = er
		}
		return
	}

	nc, ec := io.Copy(dst, src)
	written += int64(nc)
	if ec != nil {
		if ec != io.EOF {
			err = ec
		}
		return
	}

	return
}

func handlerConn(c net.Conn) {
	cer, err := tls.LoadX509KeyPair("cert.pem", "chave.pem")
	if err != nil {
		log.Println(err)
		return
	}
	config := &tls.Config{Certificates: []tls.Certificate{cer}}
	c = tls.Server(c, config)
	dst, _ := net.Dial("tcp", ":3333")
	fmt.Println(c.RemoteAddr())
	go func() {
		defer c.Close()
		defer dst.Close()
		handlerMux(dst, c)
		fmt.Println("End src -> dst")
	}()
	go func() {
		defer dst.Close()
		defer c.Close()
		_, err := io.Copy(c, dst)
		fmt.Println(err)
		fmt.Println("End src <- dst")
	}()
}

func forwardSocket() {
	sv, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}

	for {
		c, _ := sv.Accept()
		go handlerConn(c)
	}
}

func main() {
	forwardSocket()
}
