// Package redis implements a client for Redis.
//
// Client
//
// The Client implements one method; Call(). This writes your command to the
// server, then reads the subsequent reply and returns it to you. 
//
// The Client struct also has a pool of connections so it's safe to use a
// client in a concurrent context. You can create one client for your entire
// program and share it between go routines.
//
//      c := redis.NewClient("tcp:127.0.0.1:6379", 0, "")
//      reply, err := c.Call("GET", "foo")
//      if err != nil {
//          // handle error
//      }
//
//      println(reply.Elem.String())
//
// AsyncClient
// 
// The AsyncClient works exactly like the regular Client, and implements a
// single method Call(), but this method does not return any reply, only an
// error or nil. 
//
//      c := redis.NewAsyncClient("tcp:127.0.0.1:6379")
//      err := c.Call("SET", "foo", 1)
//      err = c.Call("GET", "foo")
//
// When we send our command and arguments to the Call() method nothing is sent
// to the Redis server. To get the reply for our commands from Redis we use the
// Read() method. Read sends any buffered commands to the Redis server, and
// then reads one reply. Subsequent calls to Read will return more replies or
// block if there are none.
//
//      // reply from SET 
//      reply, _ := c.Read()
//
//      // reply from GET
//      reply, _ = c.Read()
//
//      fmt.Println(reply.Elem.Int())
// 
// Due to the nature of how the AsyncClient works, it's not safe to share it
// between go routines.
package redis

import (
	"bytes"
	"strings"
)

// Client implements a Redis client which handles connections to the database
// in a pool. The size of the pool can be adjusted with by setting the
// MaxConnections variable before creating a client.
type Client struct {
	Addr     string
	Proto    string
	Db       int
	Password string
	pool     *connPool
}

// Create a new Client to connect to redis. addr must be like
// "tcp:127.0.0.1:6379".
func NewClient(addr string, db int, password string, max int) *Client {
	if addr == "" {
		addr = "tcp:127.0.0.1:6379"
	}
	na := strings.SplitN(addr, ":", 2)
	return &Client{
		Proto:    na[0],
		Addr:     na[1],
		Db:       db,
		Password: password,
		pool:     newConnPool(max),
	}
}

// Call is the canonical way of talking to Redis. It accepts any 
// Redis command and a arbitrary number of arguments.
func (c *Client) Call(args ...interface{}) (*Reply, error) {
	conn, err := c.connect()
	defer c.pool.push(conn)
	if err != nil {
		return nil, err
	}

	err = conn.Write(args...)
	if err != nil {
		return nil, err
	}
	return conn.Read()
}

// Pop a connection from pool 
func (c *Client) connect() (conn Conn, err error) {
	conn = c.pool.pop()
	if conn == nil {
		conn, err = NewConn(c.Addr, c.Proto, c.Db, c.Password)
		if err != nil {
			return nil, err
		}
	}
	return conn, nil
}

// Use the connection settings from Client to create a new AsyncClient
func (c *Client) AsyncClient() *AsyncClient {
	return &AsyncClient{c, bytes.NewBuffer(make([]byte, 0, 1024*16)), nil, 0}
}

// Async client implements an asynchronous client. It is very similar to Client
// except that it maintains a buffer of commands which first are sent to Redis
// once we explicitly request a reply.
type AsyncClient struct {
	*Client
	buf    *bytes.Buffer
	conn   Conn
	queued int
}

// Call appends a command to the write buffer or returns an error.
func (ac *AsyncClient) Call(args ...interface{}) (err error) {
	_, err = ac.buf.Write(format(args...))
	ac.queued++
	return err
}

// Read does three things. 
// 
//      1) Open connection to Redis server, if there is none.
//      2) Write any buffered commands to the server.
//      3) Try to read a reply from the server, or block on read.
//
// Read returns a Reply or error.
func (ac *AsyncClient) Read() (*Reply, error) {
	if ac.conn == nil {
		conn, e := NewConn(ac.Addr, ac.Proto, ac.Db, ac.Password)

		if e != nil {
			return nil, e
		}

		ac.conn = conn
	}

	if ac.buf.Len() > 0 {
		_, err := ac.buf.WriteTo(ac.conn.Sock())

		if err != nil {
			return nil, err
		}
	}

	reply, e := ac.conn.Read()
	ac.queued--
	return reply, e
}

func (ac *AsyncClient) Queued() int {
	return ac.queued
}

func (ac *AsyncClient) ReadAll() ([]*Reply, error) {
	replies := make([]*Reply, 0, ac.queued)

	for ac.Queued() > 0 {
		r, e := ac.Read()

		if e != nil {
			return nil, e
		}

		replies = append(replies, r)
	}

	return replies, nil
}

// The AsyncClient will only open one connection. This is not automatically
// closed, so to close it we need to call this method.
func (ac *AsyncClient) Close() {
	ac.conn.Close()
	ac.conn = nil
}
