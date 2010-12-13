package main

import (
	"ngaro"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

var Usage = func() {
	fmt.Fprint(os.Stderr, `
Gonga usage:
	gonga [options] [image file]

Gonga is the Go version of the Ngaro virtual machine.

If no image file is specified in the command line
gongaImage will be loaded, retroImage if that fails.

Options:
`)
	flag.PrintDefaults()
}

var shrink = flag.Bool("shrink", false, "shrink image dump file")
var size = flag.Int("s", 50000, "image size")
var dump = flag.String("d", "gongaImage", "image dump file")
var port = flag.Int("p", 0, "listen on port")

type withFiles []*os.File

func (wf *withFiles) String() string {
	return fmt.Sprint(*wf)
}

func (wf *withFiles) Set(value string) bool {
	if f, err := os.Open(value, os.O_RDONLY, 0666); err == nil {
		nwf := append(*wf, f)
		wf = &nwf
		return true
	}
	return false
}

func main() {
	var wf withFiles
	var l net.Listener
	flag.Var(&wf, "w", "input files")
	flag.Usage = Usage
	flag.Parse()

	ngaro.ShrinkImage = *shrink
	ngaro.ClearScreen = func() { fmt.Printf("\033[2J\033[1;1H") }

	var img []int
	var err os.Error

	switch flag.NArg() {
	case 0:
		img, err = ngaro.Load("gongaImage", *size)
		if err != nil {
			img, err = ngaro.Load("retroImage", *size)
		}
	case 1:
		img, err = ngaro.Load(flag.Arg(0), *size)
	default:
		fmt.Fprint(os.Stderr, "too many arguments\n")
		flag.Usage()
		os.Exit(2)
	}
	if *port > 0 {
		l, err = net.Listen("tcp", fmt.Sprintf(":%d", *port))
	}
	if err != nil {
		fmt.Fprint(os.Stderr, "error starting gonga: ", err.String())
		os.Exit(1)
	}

	// If listening on a port, run a VM per connection
	for l != nil {
		c, err := l.Accept()
		if err != nil {
			os.Exit(1)
		}
		img := append([]int{}, img...)
		vm := ngaro.New(img, *dump, c, c)
		go vm.Run()
	}

	// Reverse wf and add os.Stdin
	rs := make([]io.Reader, 0, len(wf)+1)
	for i, _ := range wf {
		rs = append(rs, wf[len(wf)-1-i])
	}
	input := io.MultiReader(append(rs, os.Stdin)...)

	// Run a new VM
	vm := ngaro.New(img, *dump, input, os.Stdout)
	if vm.Run() != nil {
		os.Exit(1)
	}
}
