package redis

import (
	"flag"
	"time"
)

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
