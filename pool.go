package redis

type connPool struct {
	free chan Connection
}

func newConnPool(max int) *connPool {
	p := connPool{make(chan Connection, max)}
	for i := 0; i < max; i++ {
		p.free <- nil
	}
	return &p
}

func (p *connPool) pop() Connection {
	return <-p.free
}

func (p *connPool) push(c Connection) {
	p.free <- c
}
