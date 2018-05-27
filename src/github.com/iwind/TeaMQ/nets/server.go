package nets

import (
	"net"
	"log"
	"bufio"
)

type Server struct {
	network string
	address string

	listener net.Listener

	onAcceptClient  func(client *Client)
	onCloseClient   func(client *Client)
	onReceiveClient func(client *Client, data []byte)
}

func NewServer(network, address string) *Server {
	return &Server{
		network: network,
		address: address,
	}
}

func (server *Server) AcceptClient(callback func(client *Client)) {
	server.onAcceptClient = callback
}

func (server *Server) CloseClient(callback func(client *Client)) {
	server.onCloseClient = callback
}

func (server *Server) ReceiveClient(callback func(client *Client, data []byte)) {
	server.onReceiveClient = callback
}

func (server *Server) Listen() error {
	listener, err := net.Listen(server.network, server.address)
	if err != nil {
		return err
	}
	server.listener = listener

	var id int
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		id ++

		var client = &Client{
			id:         id,
			connection: conn,
		}
		if server.onAcceptClient != nil {
			server.onAcceptClient(client)
		}
		go func(client *Client) {
			input := bufio.NewScanner(client.connection)
			for input.Scan() {
				if server.onReceiveClient != nil {
					server.onReceiveClient(client, input.Bytes())
				}
			}

			defer func() {
				client.connection.Close()
				if server.onCloseClient != nil {
					server.onCloseClient(client)
				}
			}()
		}(client)
	}
}

func (server *Server) Close() {
	server.listener.Close()
}
