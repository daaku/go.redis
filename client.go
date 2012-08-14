// Package redis implements a client for Redis.
package redis

import (
	"github.com/daaku/go.stats"
	"time"
)

// Client implements a Redis connection which is what you should
// typically use instead of the lower level Conn interface.
//
// * Connection Pooling
// * Timeouts
type Client struct {
	Addr     string
	Proto    string
	PoolSize uint
	Timeout  time.Duration
	pool     chan Conn
}

// Call is the canonical way of talking to Redis. It accepts any
// Redis command and a arbitrary number of arguments.
func (c *Client) Call(args ...interface{}) (*Reply, error) {
	start := time.Now()
	conn, err := c.connect()
	stats.Record(
		"redis connection acquire", float64(time.Since(start).Nanoseconds()))
	defer func() {
		stats.Record(
			"redis connection release", float64(time.Since(start).Nanoseconds()))
		c.pool <- conn
	}()
	if err != nil {
		stats.Inc("redis connection accquire error")
		return nil, err
	}
	err = conn.Sock().SetDeadline(start.Add(c.Timeout))
	if err != nil {
		stats.Inc("redis connection set deadline error")
		return nil, err
	}
	err = conn.Write(args...)
	stats.Record("redis write", float64(time.Since(start).Nanoseconds()))
	if err != nil {
		stats.Inc("redis write error")
		return nil, err
	}
	reply, err := conn.Read()
	stats.Record("redis read", float64(time.Since(start).Nanoseconds()))
	if err != nil {
		stats.Inc("redis read error")
	}
	return reply, err
}

// Pop a connection from the pool or create a fresh one.
func (c *Client) connect() (conn Conn, err error) {
	if c.pool == nil {
		c.pool = make(chan Conn, c.PoolSize)
		go func() {
			var i uint
			for i = 0; i < c.PoolSize; i++ {
				c.pool <- nil
			}
		}()
	}
	conn = <-c.pool
	if conn == nil {
		stats.Inc("new redis connection")
		conn, err = Dial(c.Addr, c.Proto, c.Timeout)
		if err != nil {
			return nil, err
		}
	}
	return conn, err
}
