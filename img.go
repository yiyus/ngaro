package ngaro

import (
	"bufio"
	"encoding/binary"
	"os"
)

// Ngaro Image
type Image []int

// Load returns an Image of the given size
func Load(filename string, size int) (img Image, err os.Error) {
	r, err := os.Open(filename, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	img = make([]int, size)
	br := bufio.NewReader(r)
	var ui uint32
	var i int
	for i, _ = range img {
		if err := binary.Read(br, binary.LittleEndian, &ui); err != nil {
			break
		}
		img[i] = int(ui)
	}
	return img, err
}

func (img Image) save(filename string) os.Error {
	w, err := os.Open(filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	if ShrinkImage {
		img = img[0:img[3]]
	}
	for _, i := range img {
		if err = binary.Write(w, binary.LittleEndian, uint32(i)); err != nil {
			w.Close()
			break
		}
	}
	w.Close()
	return err
}

func read0(data []int, a int) string {
	var w int
	return string(data[a:w])
}

func (img Image) string(offset int) string {
	var end int
	for end = offset; end < len(img) && img[end] != 0; end++ {
	}
	return string([]int(img[offset:end]))
}
