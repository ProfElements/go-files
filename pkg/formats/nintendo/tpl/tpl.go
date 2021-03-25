package tpl

/*
Name: Texture Palette Library
Extension: .tpl
Description: A general nintendo wii/gamecube file format that used by alot of games
*/

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type ImgFormat uint32

const (
	I4     ImgFormat = 0x00
	I8     ImgFormat = 0x01
	IA4    ImgFormat = 0x02
	IA8    ImgFormat = 0x03
	RGB565 ImgFormat = 0x04
	RGB5A3 ImgFormat = 0x05
	RGBA32 ImgFormat = 0x06
	C4     ImgFormat = 0x08
	C8     ImgFormat = 0x09
	C14X2  ImgFormat = 0x0A
	CMPR   ImgFormat = 0x0E
)

type ImgHeader struct {
	Height        uint16
	Width         uint16
	Format        uint32
	ImgDataADR    uint32
	WrapS         uint32
	WrapT         uint32
	MinFilter     uint32
	MagFilter     uint32
	LODBias       uint32
	EdgeLODEnable uint8
	MinLOD        uint8
	MaxLOD        uint8
	Unpacked      uint8
}
type PaletteHeader struct {
	entryCount uint16
	unpacked   uint8
	padding    uint8
	PalFormat  uint32
	palDataADR uint32
}
type Img struct {
	palHeader PaletteHeader
	palData   []byte
	ImgHeader ImgHeader
	ImgData   []byte
}
type ImgOffset struct {
	imgHeaderOffset    uint32
	imgPalHeaderOffset uint32
}
type Header struct {
	magic                uint32
	ImgNum               uint32
	ImgOffsetTableOffset uint32
}
type File struct {
	Header         Header
	ImgOffsetTable []ImgOffset
	ImgTable       []Img
	Data           []byte
}

type Work struct{}

func Read(data []byte) (*File, error) {

	tpl := &File{}
	if len(data) < 12 {
		return nil, fmt.Errorf("Data size is too small to be a .tpl file")
	}

	if !bytes.Equal(data[:4], []byte{0x00, 0x20, 0xAF, 0x30}) {
		return nil, fmt.Errorf("Wrong file magic, It isn't `0x00,0x20,0xAF,0x30`")
	}

	tpl.Header.magic = binary.BigEndian.Uint32(data[:4])
	tpl.Header.ImgNum = binary.BigEndian.Uint32(data[4:8])
	tpl.Header.ImgOffsetTableOffset = binary.BigEndian.Uint32(data[8:12])

	idx := uint32(12)
	var imgOffsets []ImgOffset
	imgOffsets = make([]ImgOffset, tpl.Header.ImgNum)

	for i := 0; i < int(tpl.Header.ImgNum); i++ {
		imgOffset := ImgOffset{
			imgHeaderOffset:    binary.BigEndian.Uint32(data[idx : idx+4]),
			imgPalHeaderOffset: binary.BigEndian.Uint32(data[idx+4 : idx+8]),
		}
		imgOffsets[i] = imgOffset
		idx += 8
	}
	tpl.ImgOffsetTable = imgOffsets

	var imgs []Img
	imgs = make([]Img, tpl.Header.ImgNum)

	for i := 0; i < int(tpl.Header.ImgNum); i++ {
		img := Img{}
		if tpl.ImgOffsetTable[i].imgPalHeaderOffset == 0 {
			tempPalHeader := PaletteHeader{
				entryCount: 0,
				unpacked:   0,
				padding:    0,
				PalFormat:  0,
				palDataADR: 0,
			}
			img.palHeader = tempPalHeader

		} else {

			idx = tpl.ImgOffsetTable[i].imgPalHeaderOffset
			//tempPalFormat := binary.BigEndian.Uint32(data[idx+4:idx+8])
			//tempPalDataADR := binary.BigEndian.Uint32(data[idx+8:idx+12])

			tempPalHeader := PaletteHeader{
				entryCount: binary.BigEndian.Uint16(data[idx : idx+2]),
				unpacked:   uint8(data[idx+2]),
				padding:    uint8(data[idx+3]),
				PalFormat:  binary.BigEndian.Uint32(data[idx+4 : idx+8]),
				palDataADR: binary.BigEndian.Uint32(data[idx+8 : idx+12]),
			}
			img.palHeader = tempPalHeader
			img.palData = []byte{0x00} //DO PALETTE DATA RIGHT LATER
		}

		idx = tpl.ImgOffsetTable[i].imgHeaderOffset

		if int(idx) > len(data) {
			return nil, nil
		}
		tempImgHeight := binary.BigEndian.Uint16(data[idx : idx+2])
		tempImgWidth := binary.BigEndian.Uint16(data[idx+2 : idx+4])

		if tempImgHeight > 1024 || tempImgWidth > 1024 {

			return nil, nil
		}

		tempImgHeader := ImgHeader{
			Height:        binary.BigEndian.Uint16(data[idx : idx+2]),
			Width:         binary.BigEndian.Uint16(data[idx+2 : idx+4]),
			Format:        binary.BigEndian.Uint32(data[idx+4 : idx+8]),
			ImgDataADR:    binary.BigEndian.Uint32(data[idx+8 : idx+12]),
			WrapS:         binary.BigEndian.Uint32(data[idx+12 : idx+16]),
			WrapT:         binary.BigEndian.Uint32(data[idx+16 : idx+20]),
			MinFilter:     binary.BigEndian.Uint32(data[idx+20 : idx+24]),
			MagFilter:     binary.BigEndian.Uint32(data[idx+24 : idx+28]),
			LODBias:       binary.BigEndian.Uint32(data[idx+28 : idx+32]),
			EdgeLODEnable: uint8(data[idx+32]),
			MinLOD:        uint8(data[idx+33]),
			MaxLOD:        uint8(data[idx+34]),
			Unpacked:      uint8(data[idx+35]),
		}
		img.ImgHeader = tempImgHeader

		idx = img.ImgHeader.ImgDataADR

		tempPxSize := (tempImgWidth * tempImgHeight) / 2

		img.ImgData = data[int(idx) : int(idx)+int(tempPxSize)] //supports only CMPR for now

		imgs[i] = img
	}
	tpl.ImgTable = imgs

	totalImgData := 0
	for i := 0; i < len(tpl.ImgTable); i++ {

		totalImgData += len(tpl.ImgTable[i].ImgData)
	}
	idx = tpl.ImgTable[0].ImgHeader.ImgDataADR
	tpl.Data = data[:int(idx)+totalImgData]

	return tpl, nil

}

//func Write(data File) ([]byte, error) {}
//func Decode(data File) (*Work, error) {}
//func Encode(data Work) (*File, error) {}
