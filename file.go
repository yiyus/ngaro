package ngaro

import (
	"os"
)

type file struct {
	*os.File
}

func open(filename string, mode int) file {
	switch mode {
	case 0:
		mode = os.O_RDONLY
	case 1:
		mode = os.O_WRONLY | os.O_CREATE
	case 2:
		mode = os.O_WRONLY | os.O_APPEND | os.O_CREATE
	case 3:
		mode = os.O_RDWR | os.O_CREATE
	}
	if f, err := os.Open(filename, mode, 0666); err == nil {
		return file{f}
	}
	return file{nil}
}

func (f file) read() int {
	c := make([]byte, 1)
	if _, err := f.Read(c); err == nil {
		return int(c[0])
	}
	return 0
}

func (f file) write(c int) int {
	if _, err := f.Write([]byte{byte(c)}); err == nil {
		return 1
	}
	return 0
}

func (f file) close() int {
	f.Close()
	return 1
}

func (f file) pos() int {
	if p, err := f.Seek(0, 1); err == nil {
		return int(p)
	}
	return 0
}

func (f file) seek(p int) int {
	if r, err := f.Seek(int64(p), 0); err == nil {
		return int(r)
	}
	return 0
}

func (f file) size() int {
	if d, err := f.Stat(); err == nil {
		return int(d.Size)
	}
	return 0
}

func delete(filename string) int {
	if err := os.Remove(filename); err == nil {
		return 1
	}
	return 0
}
