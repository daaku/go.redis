package redis

import (
	"flag"
	"time"
)

// Define a Client via flag parameters. For example if name is
// "redis", it will provide:
//
//     -redis.proto=unix
//     -redis.addr=/run/redis.sock
//     -redis.pool-size=10
//     -redis.timeout=1s
func ClientFlag(name string) *Client {
	client := &Client{}
	flag.StringVar(
		&client.Proto,
		name+".proto",
		"tcp",
		name+" redis proto")
	flag.StringVar(
		&client.Addr,
		name+".addr",
		"127.0.0.1:6379",
		name+" redis addr")
	flag.UintVar(
		&client.PoolSize,
		name+".pool-size",
		50,
		name+" redis connection pool size")
	flag.DurationVar(
		&client.Timeout,
		name+".timeout",
		time.Second,
		name+" redis per call timeout")
	return client
}
