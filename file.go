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

func (f file) read() (int, os.Error) {
	c := make([]byte, 1)
	_, err := f.Read(c)
	return int(c[0]), err
}

func (f file) write(c int) (int, os.Error) {
	return f.Write([]byte{byte(c)})
}

func (f file) close() (int, os.Error) {
	f.Close()
	return 1, nil
}

func (f file) pos() (int, os.Error) {
	p, err := f.Seek(0, 1)
	return int(p), err
}

func (f file) seek(p int) (int, os.Error) {
	r, err := f.Seek(int64(p), 0)
	return int(r), err
}

func (f file) size() (int, os.Error) {
	d, err := f.Stat()
	return int(d.Size), err
}

func delete(filename string) (int, os.Error) {
	err := os.Remove(filename)
	return 1, err
}
