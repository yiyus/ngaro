// Ngaro VM
// Original Ngaro Virtual Machine and Uki framework:
//   Copyright (C) 2008, 2009, 2010 Charles Childers
// Go port
//   Copyright 2009, 2010 JGL
package ngaro

import (
	"fmt"
	"io"
	"os"
)

const (
	// Instruction set
	Nop = iota
	Lit
	Dup
	Drop
	Swap
	Push
	Pop
	Loop
	Jump
	Return
	GtJump
	LtJump
	NeJump
	EqJump
	Fetch
	Store
	Add
	Sub
	Mul
	Dinod
	And
	Or
	Xor
	ShL
	ShR
	ZeroExit
	Inc
	Dec
	In
	Out
	Wait

	stackDepth = 1024
	chanBuffer = 128
	nports     = 64
)

type input struct {
	io.Reader
	next *input
}

// Virtual machine
type VM struct {
	img  Image
	ch   map[int]chan int
	dump string
	file map[int]file
	in   *input
	out  io.Writer
	ret  chan os.Error
}

// New returns a Ngaro virtual machine with the image given
// in img. Dump files will be saved with the name 'dump'.
// r and w are the input and output of the virtual machine.
func New(img Image, dump string, r io.Reader, w io.Writer) *VM {
	vm := VM{
		img:  img,
		dump: dump,
		ch:   make(map[int]chan int),
		file: make(map[int]file),
		in:   &input{r, nil},
		out:  w,
		ret:  make(chan os.Error),
	}
	return &vm
}

// Run starts the execution of the Ngaro virtual machine vm and
// returns a finalization chanel.
func (vm *VM) Run() chan os.Error {
	go vm.core(0)
	return vm.ret
}

// Channel returns the channel with the given id. This chanel
// can be used to communicate with any running core.
func (vm *VM) Channel(id int) chan int {
	if c, ok := vm.ch[id]; ok {
		return c
	}
	vm.ch[id] = make(chan int)
	return vm.ch[id]
}

func (vm *VM) core(ip int) {
	var port [nports]int
	var sp, rsp int
	var tos int
	var data, addr [stackDepth]int
	defer func() {
		if v := recover(); v != nil {
			if err, ok := v.(os.Error); ok {
				vm.ret <- err
			} else {
				vm.ret <- os.NewError("unknown error:"+fmt.Sprint(v))
			}
		}
	}()
	for ; ip < len(vm.img); ip++ {
		switch vm.img[ip] {
		case Nop:
		case Lit:
			sp++
			ip++
			data[sp] = vm.img[ip]
		case Dup:
			sp++
			data[sp] = tos
		case Drop:
			sp--
		case Swap:
			data[sp], data[sp-1] = data[sp-1], tos
		case Push:
			rsp++
			addr[rsp] = tos
			sp--
		case Pop:
			sp++
			data[sp] = addr[rsp]
			rsp--
		case Loop:
			data[sp]--
			if data[sp] != 0 && data[sp] > -1 {
				ip++
				ip = vm.img[ip] - 1
			} else {
				ip++
				sp--
			}
		case Jump:
			ip++
			ip = vm.img[ip] - 1
		case Return:
			ip = addr[rsp]
			rsp--
		case GtJump:
			ip++
			if data[sp-1] > tos {
				ip = vm.img[ip] - 1
			}
			sp = sp - 2
		case LtJump:
			ip++
			if data[sp-1] < tos {
				ip = vm.img[ip] - 1
			}
			sp = sp - 2
		case NeJump:
			ip++
			if data[sp-1] != tos {
				ip = vm.img[ip] - 1
			}
			sp = sp - 2
		case EqJump:
			ip++
			if data[sp-1] == tos {
				ip = vm.img[ip] - 1
			}
			sp = sp - 2
		case Fetch:
			data[sp] = vm.img[tos]
		case Store:
			vm.img[tos] = data[sp-1]
			sp = sp - 2
		case Add:
			data[sp-1] += tos
			sp--
		case Sub:
			data[sp-1] -= tos
			sp--
		case Mul:
			data[sp-1] *= tos
			sp--
		case Dinod:
			data[sp] = data[sp-1] / tos
			data[sp-1] = data[sp-1] % tos
		case And:
			data[sp-1] &= tos
			sp--
		case Or:
			data[sp-1] |= tos
			sp--
		case Xor:
			data[sp-1] ^= tos
			sp--
		case ShL:
			data[sp-1] <<= uint(tos)
			sp--
		case ShR:
			data[sp-1] >>= uint(tos)
			sp--
		case ZeroExit:
			if tos == 0 {
				sp--
				ip = addr[rsp]
				rsp--
			}
		case Inc:
			data[sp]++
		case Dec:
			data[sp]--
		case In:
			data[sp] = port[tos]
			port[tos] = 0
		case Out:
			port[0] = 0
			port[tos] = data[sp-1]
			sp = sp - 2
		case Wait:
			sp -= vm.wait(&port, tos, sp, rsp, data[0:])
		default:
			rsp++
			addr[rsp] = ip
			ip = vm.img[ip] - 1
		}
		tos = data[sp]
	}
}
