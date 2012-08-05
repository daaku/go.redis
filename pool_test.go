package redis

import (
	"testing"
	"time"
)

func TestPool(t *testing.T) {
	max := 50
	p := newConnPool(max)

	for i := int(max) + 1; i >= 0; i-- {
		c := p.pop()

		go func() {
			time.Sleep(1e+4)
			p.push(c)
		}()
	}
}
