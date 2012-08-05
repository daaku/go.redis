package redis

import (
	"github.com/daaku/go.redis/bufin"
	"net"
)

// Represents a single Connection to the server and abstracts the
// read/write via the connection. Unless you're implementing your own
// client, you should use the Client interface.
type Conn interface {
	// Write accepts any redis command and arbitrary list of arguments.
	//
	//     Write("SET", "counter", 1)
	//     Write("INCR", "counter")
	//
	// Write might return a net.Conn.Write error
	Write(args ...interface{}) error

	// Read reads one reply of the socket connection. If there is no reply waiting
	// this method will block.
	Read() (*Reply, error)

	// Close the Connection.
	Close() error

	// Returns the underlying net.Conn. This is useful for example to set
	// set a r/w deadline on the connection.
	//
	//      Sock().SetDeadline(t)
	Sock() net.Conn
}

type connection struct {
	rbuf *bufin.Reader
	c    net.Conn
}

// NewConn expects a network address and protocol.
//
//     NewConn("127.0.0.1:6379", "tcp")
//
// or for a unix domain socket
//
//     NewConn("/path/to/redis.sock", "unix")
func NewConn(addr, proto string, db int, password string) (Conn, error) {
	conn, err := net.Dial(proto, addr)
	if err != nil {
		return nil, err
	}
	c := &connection{bufin.NewReader(conn), conn}
	if password != "" {
		err := c.Write("AUTH", password)
		if err != nil {
			return nil, err
		}
		_, err = c.Read()
		if err != nil {
			return nil, err
		}
	}
	if db != 0 {
		err := c.Write("SELECT", db)
		if err != nil {
			return nil, err
		}
		_, err = c.Read()
		if err != nil {
			return nil, err
		}
	}
	return c, nil
}

func (c *connection) Read() (*Reply, error) {
	reply := Parse(c.rbuf)
	if reply.Err != nil {
		return nil, reply.Err
	}
	return reply, nil
}

func (c *connection) Write(args ...interface{}) error {
	_, err := c.c.Write(format(args...))
	if err != nil {
		return err
	}
	return nil
}

func (c *connection) Close() error {
	return c.c.Close()
}

func (c *connection) Sock() net.Conn {
	return c.c
}
