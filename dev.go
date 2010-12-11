// Ngaro VM
// Original Ngaro Virtual Machine and Uki framework:
//   Copyright (C) 2008, 2009, 2010 Charles Childers
// Go port
//   Copyright 2009, 2010 JGL
package ngaro

import (
	"io"
	"os"
)

var fd int
var ClearScreen func() = func() {}
var ShrinkImage = false

func (vm *VM) wait(port *[nports]int, tos, sp, rsp int, data []int) (drop int) {
	c := make([]byte, 1)
	switch {
	case port[0] == 1:
		return

	case port[1] == 1: // Input
	readInput:
		switch _, err := vm.in.Read(c); err {
		case os.EOF:
			if rc, ok := vm.in.Reader.(io.ReadCloser); ok {
				rc.Close()
			}
			vm.in = vm.in.next
			if vm.in != nil {
				goto readInput
			} else {
				vm.ret <- os.EOF
			}
		case nil:
			port[1] = int(c[0])
		}

	case port[2] == 1: // Output
		c[0] = byte(tos)
		if tos < 0 {
			ClearScreen()
		} else if _, err := vm.out.Write(c); err == nil {
			port[2] = 0
			drop = 1
		}

	case port[4] != 0: // Files
		switch port[4] {
		case 1: // Write dump
			vm.img.save(vm.dump)
			port[4] = 0
		case 2: // Include file
			name := vm.img.string(tos)
			if f, err := os.Open(name, os.O_RDONLY, 0); err == nil {
				vm.in = &input{f, vm.in}
			}
			port[4] = 0
		case -1:
			vm.file[fd] = open(vm.img.string(data[sp-1]), tos)
			port[4] = fd
			fd++
			drop = 2
		case -2:
			port[4] = vm.file[tos].read()
			drop = 1
		case -3:
			port[4] = vm.file[tos].write(data[sp-1])
			drop = 2
		case -4:
			port[4] = vm.file[tos].close()
			vm.file[tos] = file{}, false
			drop = 1
		case -5:
			port[4] = vm.file[tos].pos()
			drop = 1
		case -6:
			port[4] = vm.file[tos].seek(data[sp-1])
			drop = 2
		case -7:
			port[4] = vm.file[tos].size()
			drop = 1
		case -8:
			port[4] = delete(vm.img.string(tos))
			drop = 1
		}

	case port[5] != 0: // Capabilities
		switch port[5] {
		case -1: // Image size
			port[5] = len(vm.img)
		case -2: // canvas exists?
			port[5] = 0
		case -3: // screen width
			port[5] = 0
		case -4: // screen height
			port[5] = 0
		case -5: // Stack depth
			port[5] = sp - 2
		case -6: // Address stack depth
			port[5] = rsp
		case -7: // mouse exists?
			port[5] = 0
		case -8: // Seconds from the epoch
			if t, _, err := os.Time(); err == nil {
				port[5] = int(t)
			}
		case -9: // Bye!
			vm.ret <- nil
		case -10: // getenv
			env := os.Getenv(vm.img.string(tos))
			copy(vm.img[data[sp-1]:], []int(env))
		default:
			port[5] = 0
		}

	// TODO: port[6] (canvas)
	// TODO: port[7] (mouse)

	case port[13] == 1:
		go vm.core(tos)
		port[13] = 0
		drop = 1
	}
	port[0] = 1
	return
}
