package yaz0

/*
NAME: YAZ0
EXTENSION: .szs
DESCRIPTION: Nintendo's run-length encoding, It is used a lot in Nintendo games across their various consoles
*/

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type Header struct {
	Magic     string
	DataSize  uint32
	reserved1 uint32
	reserved2 uint32
}
type File struct {
	Header Header
	Data   []byte
}
type Work []byte

func Read(data []byte) (*File, error) {
	file := &File{}

	if !bytes.Equal(data[:4], []byte("YAZ0")) {
		return nil, fmt.Errorf("This is not a yaz0 encoded file, the  file magic is wrong!")
	}

	file.Header.Magic = string(data[:4])
	file.Header.DataSize = binary.BigEndian.Uint32(data[4:8])
	file.Header.reserved1 = binary.BigEndian.Uint32(data[8:12])
	file.Header.reserved2 = binary.BigEndian.Uint32(data[12:16])
	file.Data = data[16:]

	return file, nil
}
func Write(data *File) ([]byte, error) {
	buffer := &bytes.Buffer{}
	_, err := buffer.WriteString("YAZ0")
	if err != nil {
		return nil, err
	}

	err = binary.Write(buffer, binary.BigEndian, uint32(data.Header.DataSize))
	if err != nil {
		return nil, err
	}

	err = binary.Write(buffer, binary.BigEndian, uint32(0))
	if err != nil {
		return nil, err
	}

	err = binary.Write(buffer, binary.BigEndian, uint32(0))
	if err != nil {
		return nil, err
	}

	err = binary.Write(buffer, binary.BigEndian, data.Data)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}
func Decode(data *File) Work {
	return decompress(data.Data)
}
func Encode(data Work) *File {
	file := &File{}

	file.Header.Magic = "YAZ0"
	file.Header.DataSize = uint32(len(data))
	file.Header.reserved1 = uint32(0)
	file.Header.reserved2 = uint32(0)
	file.Data = compress(data)

	return file
}

func decompress(data []byte) []byte {
	//temporarily just return data
	return data
}
func compress(data []byte) []byte {
	//temporarily just return data
	return data
}
