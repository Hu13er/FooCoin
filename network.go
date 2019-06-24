package main

import (
	"errors"
	"io"
	"log"
	"net"
)

type Node struct {
	Name    string
	Addr    string
	Parties map[string]net.Conn

	listener *net.TCPListener
	stop     chan struct{}
}

func (n *Node) Start() error {
	addr, err := net.ResolveTCPAddr("tcp4", n.Addr)
	if err != nil {
		return err
	}
	n.listener, err = net.ListenTCP("tcp", addr)
	if err != nil {
		return err
	}
	go n.listen()
	return nil
}

func (n *Node) listen() error {
	defer n.listener.Close()
	ch := make(chan net.Conn)
	cherr := make(chan error)
	for {
		n.accept(ch, cherr)
		select {
		case conn := <-ch:
			hername, err := read(conn)
			if err != nil {
				return err
			}
			if err := write([]byte(n.Name), conn); err != nil {
				return err
			}
			n.Parties[string(hername)] = conn
		case err := <-cherr:
			return err
		case <-n.stop:
			return nil
		}
	}
}

func (n *Node) accept(ch chan net.Conn, cherr chan error) {
	c, err := n.listener.Accept()
	if err == nil {
		ch <- c
	} else {
		cherr <- err
	}
}

func (n *Node) Connect(addr string) error {
	ipaddr, err := net.ResolveTCPAddr("tcp4", addr)
	if err != nil {
		return err
	}
	conn, err := net.DialTCP("tcp", nil, ipaddr)
	if err != nil {
		return err
	}
	if err := write([]byte(n.Name), conn); err != nil {
		return err
	}
	hername, err := read(conn)
	if err != nil {
		return err
	}
	n.Parties[string(hername)] = conn
	return nil
}

func (n *Node) SendAll(data []byte) {
	for _, p := range n.Parties {
		err := write(data, p)
		if err != nil {
			log.Println("WARN:", err)
		}
	}
}

func (n *Node) Send(party string, data []byte) error {
	conn, exists := n.Parties[party]
	if !exists {
		return errors.New("no party")
	}
	return write(data, conn)
}

func (n *Node) ReadAny(f func(string, []byte)) {
	for name, conn := range n.Parties {
		go func(name string, conn net.Conn) {
			for {
				data, err := read(conn)
				if err != nil {
					log.Println("WARN:", err)
					continue
				}
				f(name, data)
			}
		}(name, conn)
	}
}

func (n *Node) Stop() {
	for _, conn := range n.Parties {
		err := conn.Close()
		if err != nil {
			log.Println("WARN:", err)
		}
	}
	n.stop <- struct{}{}
}

func read(r io.Reader) ([]byte, error) {
	size := readSize(r)
	payload := make([]byte, size)
	if _, err := r.Read(payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func write(data []byte, w io.Writer) error {
	writeSize(len(data), w)
	_, err := w.Write(data)
	return err
}

func readSize(r io.Reader) (size int) {
	buf := make([]byte, 4)
	r.Read(buf)
	for i := 0; i < 4; i++ {
		size <<= 4
		size += int(buf[i])
	}
	return size
}

func writeSize(size int, w io.Writer) {
	buf := make([]byte, 4)
	for i := 0; i < 4; i++ {
		buf[3-i] = byte(size & 0xff)
		size >>= 4
	}
	w.Write(buf)
}
