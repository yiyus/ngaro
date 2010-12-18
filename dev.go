package ngaro

import (
	"io"
	"os"
)

var fd int32

type input []io.Reader

func (in *input)Read(b []byte) (n int, err os.Error) {
	r := (*in)[len(*in)-1]
	if n, err = r.Read(b); err == os.EOF {
		if rc, ok := r.(io.ReadCloser); ok {
			rc.Close()
		}
		if len(*in) > 1 {
			(*in) = (*in)[:len(*in)-1]
			return in.Read(b)
		}
	}
	return
}

func open(filename string, mode int32) *os.File {
	var m int
	switch mode {
	case 0:
		m = os.O_RDONLY
	case 1:
		m = os.O_WRONLY | os.O_CREATE
	case 2:
		m = os.O_WRONLY | os.O_APPEND | os.O_CREATE
	case 3:
		m = os.O_RDWR | os.O_CREATE
	}
	if f, err := os.Open(filename, m, 0666); err == nil {
		return f
	}
	return nil
}

func (vm *VM) wait(data, addr, port []int32) (drop int) {
	sp, rsp := len(data)-1, len(addr)-1
	tos := data[sp]
	c := make([]byte, 1)
	switch {
	case port[0] == 1:
		return

	case port[1] == 1: // Input
		if _, err := vm.in.Read(c); err != nil {
			panic(err)
		}
		port[1] = int32(c[0])

	case port[1] > 1: // Receive from (or delete) channel
		if port[2] == port[1] {
			vm.ch[port[1]] = nil, false
			port[1] = 0
			port[2] = 0
		} else {
			port[1] = <-vm.Chan(port[1])
		}

	case port[2] == 1: // Output
		c[0] = byte(tos)
		if tos < 0 {
			ClearScreen()
		} else if _, err := vm.out.Write(c); err == nil {
			port[2] = 0
			drop = 1
		}

	case port[2] > 1: // Send to channel
		vm.Chan(port[2]) <- tos
		port[2] = 0
		drop = 1

	case port[4] != 0: // Files
		var err os.Error
		var p int64
		switch port[4] {
		case 1: // Write dump
			vm.img.save(vm.dump)
			port[4] = 0
		case 2: // Include file
			name := vm.img.string(tos)
			if f, err := os.Open(name, os.O_RDONLY, 0); err == nil {
				in := append(*vm.in, f)
				vm.in = &in
			}
			port[4] = 0
			drop = 1
		case -1:
			vm.file[fd] = open(vm.img.string(data[sp-1]), tos)
			port[4] = fd
			fd++
			drop = 2
		case -2:
			_, err = vm.file[tos].Read(c)
			port[4] = int32(c[0])
			drop = 1
		case -3:
			c[0] = byte(data[sp-1])
			n, _ := vm.file[tos].Write(c)
			port[4] = int32(n)
			drop = 2
		case -4:
			port[4], err = 1, vm.file[tos].Close()
			if err != nil {
				vm.file[tos] = nil, false
			}
			drop = 1
		case -5:
			p, err = vm.file[tos].Seek(0, 1)
			port[4] = int32(p)
			drop = 1
		case -6:
			p, err = vm.file[tos].Seek(int64(data[sp-1]), 0)
			port[4] = int32(p)
			drop = 2
		case -7:
			var fi *os.FileInfo
			fi, err = vm.file[tos].Stat()
			port[4] = int32(fi.Size)
			drop = 1
		case -8:
			port[4], err = 1, os.Remove(vm.img.string(tos))
			drop = 1
		}
		if err != nil {
			port[4] = 0
		}

	case port[5] != 0: // Capabilities
		switch port[5] {
		case -1: // Image size
			port[5] = int32(len(vm.img))
		case -2: // canvas exists?
			port[5] = 0
		case -3: // screen width
			port[5] = 0
		case -4: // screen height
			port[5] = 0
		case -5: // Stack depth
			port[5] = int32(sp)
		case -6: // Address stack depth
			port[5] = int32(rsp)
		case -7: // mouse exists?
			port[5] = 0
		case -8: // Seconds from the epoch
			if t, _, err := os.Time(); err == nil {
				port[5] = int32(t)
			}
		case -9: // Bye!
			panic(nil)
		case -10: // getenv
			env := os.Getenv(vm.img.string(tos))
			off := int(data[sp-1])
			for i, b := range env {
				vm.img[off+i] = int32(b)
			}
		default:
			port[5] = 0
		}

	// TODO: port[6] (canvas)
	// TODO: port[7] (mouse)

	case port[13] == 1:
		go vm.core(tos)
		port[13] = 0
	}
	port[0] = 1
	return
}
