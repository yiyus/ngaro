// Ngaro VM
// Original Ngaro Virtual Machine and Uki framework:
//   Copyright (C) 2008, 2009, 2010 Charles Childers
// Go port
//   Copyright 2009, 2010 JGL
package main

import (
	"ngaro"
	"flag"
	"fmt"
	"io"
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
var dump = flag.String("d", "retro.img", "image dump file")

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

	if err != nil {
		fmt.Fprint(os.Stderr, "error starting gonga: ", err.String())
		os.Exit(1)
	}

	// Reverse with files order and add os.Stdin
	rs := make([]io.Reader, 0, len(wf)+1)
	for i, _ := range wf {
		rs = append(rs, wf[len(wf)-1-i])
	}
	input := io.MultiReader(append(rs, os.Stdin)...)

	vm := ngaro.New(img, *dump, input, os.Stdout)

	// vm.Run() returns a chanel that blocks until the vm finishes
	<-vm.Run()
}
