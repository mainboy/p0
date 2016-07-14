// Implementation of a MultiEchoServer. Students should write their code in this file.

package p0

import (
	"bufio"
	"net"
	"strconv"
)

type multiEchoServer struct {
	// TODO: implement this!
	client_count   int
	clients        map[string]multiEchoClient
	listener       net.Listener
	client_mutux   chan bool
	server_message chan []byte
}

type multiEchoClient struct {
	client_message chan []byte
	connection     net.Conn
}

// New creates and returns (but does not start) a new MultiEchoServer.
func New() MultiEchoServer {
	// TODO: implement this!
	server := new(multiEchoServer)
	server.client_count = 0
	server.clients = make(map[string]multiEchoClient)
	server.listener = nil
	server.client_mutux = make(chan bool, 1)
	server.server_message = make(chan []byte)

	return server
}

func NewClient(connection net.Conn) multiEchoClient {
	client := new(multiEchoClient)
	client.client_message = make(chan []byte, 100)
	client.connection = connection

	return *client
}

func (mes *multiEchoServer) Start(port int) error {
	// TODO: implement this!
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	mes.listener = listener
	unlock(mes)
	if err != nil {
		return err
	}
	// broadcast message to all clients
	go func() {
		for {
			message := <-mes.server_message
			lock(mes)
			for _, client := range mes.clients {
				if len(client.client_message) < 100 {
					client.client_message <- message
				}
			}
			unlock(mes)
		}
	}()
	// set up server listener and add client connection
	go func() {
		for {
			net_connection, err := mes.listener.Accept()
			if err != nil {
				continue
			}
			add_client(net_connection, mes)
		}
	}()

	return nil
}

func (mes *multiEchoServer) Close() {
	// TODO: implement this!
	lock(mes)
	for _, client := range mes.clients {
		client.connection.Close()
		mes.client_count--
	}

	mes.listener.Close()
	unlock(mes)
}

func (mes *multiEchoServer) Count() int {
	// TODO: implement this!
	lock(mes)
	count := mes.client_count
	unlock(mes)
	return count
}

// TODO: add additional methods/functions below!
func add_client(net_connection net.Conn, mes *multiEchoServer) {
	newclient := NewClient(net_connection)
	lock(mes)
	mes.clients[net_connection.RemoteAddr().String()] = newclient
	mes.client_count++
	unlock(mes)
	//write message
	go func() {
		for {
			message := <-newclient.client_message
			_, err := net_connection.Write(message)
			if err != nil {
				del_client(net_connection, mes)
				return
			}
		}
	}()
	// read message
	go func() {
		str := bufio.NewReader(net_connection)
		for {
			message, err := str.ReadBytes('\n')
			if err != nil {
				del_client(net_connection, mes)
				return
			}
			mes.server_message <- message
		}
	}()
}

func del_client(net_connection net.Conn, mes *multiEchoServer) {
	lock(mes)
	delete(mes.clients, net_connection.RemoteAddr().String())
	mes.client_count--
	unlock(mes)
}

func lock(mes *multiEchoServer) {
	<-mes.client_mutux
}

func unlock(mes *multiEchoServer) {
	mes.client_mutux <- true
}
