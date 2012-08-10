package redis_test

import (
	"bytes"
	"github.com/daaku/go.redis"
	"strconv"
	"testing"
)

func error_(t *testing.T, name string, expected, got interface{}, err error) {
	if err != nil {
		t.Errorf("`%s` expected `%v` got `%v`, err(%v)", name, expected, got, err.Error())
	} else {
		t.Errorf("`%s` expected `%v` got `%v`, err(%v)", name, expected, got, err)
	}
}

func TestClient(t *testing.T) {
	c := redis.NewClient("", 0, "", 50)

	if _, err := c.Call("SET", "foo", "foo"); err != nil {
		t.Fatal(err.Error())
	}
}

func BenchmarkItoa(b *testing.B) {
	for i := 0; i < b.N; i++ {
		strconv.Itoa(i)
	}
}

func BenchmarkSet(b *testing.B) {
	c := redis.NewClient("", 0, "", 50)

	for i := 0; i < b.N; i++ {
		c.Call("SET", "foo", "foo")
	}
}

func BenchmarkAppendUint(b *testing.B) {
	var buf []byte
	buf = make([]byte, 0, 1024*16)

	for i := 0; i < b.N; i++ {
		strconv.AppendUint(buf, uint64(i), 10)
	}
}

func BenchmarkAppendBytes(b *testing.B) {
	var buf []byte
	buf = make([]byte, 0, 1024*16)

	for i := 0; i < b.N; i++ {
		buf = append(buf, '\r')
	}
}

func BenchmarkAppendBuffer(b *testing.B) {
	buf := bytes.NewBuffer(make([]byte, 1024*16))

	for i := 0; i < b.N; i++ {
		buf.WriteByte('\r')
	}
}
