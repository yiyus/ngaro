package ngaro

import (
	"encoding/binary"
	"os"
)

// Ngaro Image
type Image []int32

// Load returns an Image of the given size
func Load(filename string, size int) (Image, os.Error) {
	f, err := os.Open(filename, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img := make([]int32, size)
	for i := 0; i < len(img) && err == nil; i++ {
		err = binary.Read(f, binary.LittleEndian, &img[i])
	}
	if err == os.EOF {
		return img, nil
	}
	return nil, err
}

func (img Image) save(filename string, shrink bool) os.Error {
	f, err := os.Open(filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	if shrink {
		img = img[0:img[3]]
	}
	return binary.Write(f, binary.LittleEndian, img)
}

func (img Image) string(offset int32) string {
	off := int(offset)
	last := off
	for ; last < len(img) && img[last] != 0; last++ {
	}
	str := make([]byte, last-off)
	for i, _ := range str {
		str[i] = byte(img[off+i])
	}
	return string(str)
}
