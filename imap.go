package main

import (
	"time"
	"io"
	"bufio"
	"net"
	"crypto/tls"
	"fmt"
	"strings"
	"log"
)

const (
	cr   = '\r'
	lf   = '\n'
	space = ' '
)
var crlf = []byte{cr, lf}

type Command struct {
	Id string `json:"id"`

	// Command name without the UID prefix.
	Name string `json:"name"`

	// Arguments for command
	Args []string `json:"args"`

	// // UID flag for FETCH, STORE, COPY, and SEARCH commands.
    // uid bool
}

func (cmd *Command) ToBytes() []byte {
	s := []byte(fmt.Sprintf("%s %s", cmd.Id, cmd.Name))
	for _, arg := range cmd.Args {
		s = append(s, space)
		s = append(s, UTF7EncodeBytes([]byte(arg))...)
	}
	return append(s, crlf...)
}

type Response struct {
	Id string `json:"id"`
	Body string `json:"body"`
}

type ConnState uint8

const (
	UnknownState = ConnState(1 << iota) // not connected
	ConnectedState						// not authenticated
	AuthState							// authenticated
	// SelectedState // mailbox selected
	// Logout
	ClosedState = ConnState(0)          // connection closed
)

// Timeout values for the Dial functions.
const (
	netTimeout    = 30 * time.Second // Time to establish a TCP connection
	clientTimeout = 60 * time.Second // Time to receive greeting and capabilities
)

type Client struct {
	conn net.Conn
	reader *bufio.Reader
	host string
	send chan *Command
	recv chan *Response
	state ConnState
	timeout time.Duration
}

func Dial(addr string) (*Client, error) {
	addr = defaultPort(addr, "143")
	conn, err := net.DialTimeout("tcp", addr, netTimeout)
	if err != nil {
		return nil, err
	}
	host, _, _ := net.SplitHostPort(addr)
	return NewClient(conn, host, clientTimeout)
}

func DialTLS(addr string, config *tls.Config) (*Client, error) {
	addr = defaultPort(addr, "993")
	conn, err := net.DialTimeout("tcp", addr, netTimeout)
	if err != nil {
		return nil, err
	}
	host, _, _ := net.SplitHostPort(addr)
	tlsConn := tls.Client(conn, setServerName(config, host))
	return NewClient(tlsConn, host, clientTimeout)
}


func NewClient(conn net.Conn, host string, timeout time.Duration) (*Client, error) {
	c := &Client{
		conn: conn,
		reader: bufio.NewReader(conn),
		host: host,
		send: make(chan *Command),
		recv: make(chan *Response),
		state: UnknownState,
		timeout: timeout,
	}

	err := c.Start()

	return c, err
}

func (c *Client) Start() error {
	data, err := c.readLine()
	if err != nil {
		return err
	}
	if !strings.HasPrefix(data, "* OK ") {
		return fmt.Errorf("Invalid greeting")
	}

	c.state = ConnectedState

	// go func () {
	// 	for {
	// 		cmd := <- c.send
	// 		fmt.Printf("%v\n", cmd)
	// 	}
	// }()

	return nil
}

func (c *Client) Receive() (*Response, error) {
	r := &Response{"", ""}

	for {
		data, err := c.readLine()
		if err != nil {
			return nil, err
		}

		r.Body += data
		if data[0] != '*' {
			r.Id = strings.SplitN(data, " ", 2)[0]
			break
		}
	}

	return r, nil
}

func (c *Client) readLine() (string, error) {
	if c.state == ClosedState {
		return "", io.EOF
	}

	data, err := c.reader.ReadString('\n')
	log.Printf("<<< %s", data)
	return data, err
}

func (c *Client) Send(cmd *Command) error {
	data := cmd.ToBytes()
	log.Printf(">>> %s", data)
	_, err := c.conn.Write(data)
	return err
}
