package watcher

import (
	"io"
	"sync"

	"github.com/jlaffaye/ftp"
)

type syncConnection struct {
	*ftp.ServerConn
	m *sync.Mutex
}

func (c *syncConnection) ListSave(path string) (list []*ftp.Entry, err error) {
	c.m.Lock()
	list, err = c.List(path)
	c.m.Unlock()

	return
}

func (c *syncConnection) StorSave(path string, r io.Reader) (err error) {
	c.m.Lock()
	err = c.Stor(path, r)
	c.m.Unlock()

	return
}
