package redis

type connPool struct {
	free chan Conn
}

func newConnPool(max int) *connPool {
	p := connPool{make(chan Conn, max)}
	for i := 0; i < max; i++ {
		p.free <- nil
	}
	return &p
}

func (p *connPool) pop() Conn {
	return <-p.free
}

func (p *connPool) push(c Conn) {
	p.free <- c
}
