package nets

import (
	"net"
	"bufio"
)

type Client struct {
	id         int
	connection net.Conn
}

func (client *Client) Id() int {
	return client.id
}

func (client *Client) SetId(id int) {
	client.id = id
}

func (client *Client) Write(message string) (int, error) {
	return client.connection.Write([]byte(message))
}

func (client *Client) Writeln(message string) (int, error) {
	return client.connection.Write([]byte(message + "\n"))
}

func (client *Client) WriteBytes(bytes []byte) (int, error) {
	return client.connection.Write(bytes)
}

func (client *Client) Close() {
	client.connection.Close()
}

func (client *Client) Connect(network string, address string) error {
	conn, err := net.Dial(network, address)
	if err != nil {
		return err
	}
	client.connection = conn
	return nil
}

func (client *Client) Receive(receiver func(message string)) {
	scanner := bufio.NewScanner(client.connection)
	for scanner.Scan() {
		receiver(scanner.Text())
	}
}
