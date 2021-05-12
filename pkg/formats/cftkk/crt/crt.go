package crt

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"os"
)

/*
	NAME: Krusty Texture
	EXT: .krt | none
	DESCRIPTION: The texture format found in Spongebob Squarepants: Creature from the Krusty Krab
	BINARY STRUCTURE:

	0x20 of zeroes;       Was probably as header that got stripped.
	width of image;       uint32
	height of image;      uint32
	imageFormat of image; uint322
	blockSize of image;   uint16, can be 16, 32, or 64, which is block size 4x4 8x4 8x8 respectively
  unknown1;             uint8, definitely a flag of some sort. Almost always 1
	unknown2;             uint8, definitely a flag of some sort. Almost always 1
	4 bytes of zeroes;    just padding
	unknown3;             uint8 almost definitely a flag. Either 0xFF or 0x00 > uses uint32 space.
	imageSize of image;   width*height.
	0x30 of zeroes;       padding of some sort?
	paletteOffset;        offset to palette for paletted images
	imageOffset;          offset to image data
	fileSize;             size of file
	padding;              pad until paletteOffset if not 0, otherwise pad until imageOffset
*/

type KRTImage struct {
	width         uint32
	height        uint32
	imageFormat   uint32
	blockSize     uint16
	unknown3      uint8
	paletteOffset uint32
	imageOffset   uint32
	fileSize      uint32
	paletteData   []byte
	imageData     []byte
}

func ReadKRT(filepath string) (*KRTImage, error) {
	raw, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("error while reading KRTImage %v", err)
	}

	if len(raw) < 0xA0 {
		return nil, fmt.Errorf("data is not large enough to be KRTImage")
	}

	//index starts at 0x20 because the the first 0x20 bytes are just zeroes and hold no value
	image := &KRTImage{}
	index := 0x20

	image.width = binary.BigEndian.Uint32(raw[index : index+4])
	index += 4

	image.height = binary.BigEndian.Uint32(raw[index : index+4])
	index += 4

	image.imageFormat = binary.BigEndian.Uint32(raw[index : index+4])
	index += 4

	image.blockSize = binary.BigEndian.Uint16(raw[index : index+2])
	index += 2

	//skip unknown1 and unknown2 and just assume they are one until otherwise found
	index += 2

	//Skip 4 padding bytes.
	index += 4

	image.unknown3 = uint8(raw[index])
	index += 4

	//Skip imageSize
	index += 4

	//Skip 0x30 padding bytes
	index += 0x30

	image.paletteOffset = binary.BigEndian.Uint32(raw[index : index+4])
	index += 4

	image.imageOffset = binary.BigEndian.Uint32(raw[index : index+4])
	index += 4

	image.fileSize = binary.BigEndian.Uint32(raw[index : index+4])

	if image.paletteOffset != 0 {
		image.paletteData = raw[image.paletteOffset:image.imageOffset]
	}

	image.imageData = raw[image.imageOffset:image.fileSize]

	return image, nil
}

func (image *KRTImage) WriteKRT(filepath string) error {
	buf := bytes.NewBuffer([]byte{})

	//Write padding
	binary.Write(buf, binary.BigEndian, make([]byte, 0x20))

	binary.Write(buf, binary.BigEndian, image.width)
	binary.Write(buf, binary.BigEndian, image.height)
	binary.Write(buf, binary.BigEndian, image.imageFormat)
	binary.Write(buf, binary.BigEndian, image.blockSize)

	//Write unknown1 and unknown as 1 since that is what almost all textures are
	binary.Write(buf, binary.BigEndian, []byte{0x01, 0x01})

	//Write padding
	binary.Write(buf, binary.BigEndian, make([]byte, 4))

	binary.Write(buf, binary.BigEndian, uint32(image.unknown3))
	binary.Write(buf, binary.BigEndian, uint32(image.width*image.height))
	//Write padding
	binary.Write(buf, binary.BigEndian, make([]byte, 0x30))

	binary.Write(buf, binary.BigEndian, image.paletteOffset)
	binary.Write(buf, binary.BigEndian, image.imageOffset)
	binary.Write(buf, binary.BigEndian, image.fileSize)

	if image.paletteOffset != 0 && buf.Len() < int(image.paletteOffset) {
		padding := int(image.paletteOffset) - buf.Len()
		binary.Write(buf, binary.BigEndian, make([]byte, padding))
		binary.Write(buf, binary.BigEndian, image.paletteData)
		binary.Write(buf, binary.BigEndian, image.imageData)

		err := os.WriteFile(filepath, buf.Bytes(), 0666)
		if err != nil {
			return fmt.Errorf("failed to write file due to %v to %v", err, filepath)
		}

		return nil
	}

	if buf.Len() < int(image.imageOffset) {
		padding := int(image.imageOffset) - buf.Len()
		binary.Write(buf, binary.BigEndian, make([]byte, padding))
		binary.Write(buf, binary.BigEndian, image.imageData)

		err := os.WriteFile(filepath, buf.Bytes(), 0666)
		if err != nil {
			return fmt.Errorf("failed to write file due to %v to %v", err, filepath)
		}

		return nil
	}

	return nil
}

func EncodeToKRT(rgba *image.RGBA) (*KRTImage, error) {
	image := &KRTImage{
		width:         uint32(rgba.Rect.Dx()),
		height:        uint32(rgba.Rect.Dy()),
		imageFormat:   0xF,
		blockSize:     16,
		unknown3:      0xFF, //Assume it FF for now
		paletteOffset: 0x00,
		imageOffset:   0xA0,
		fileSize:      uint32(0xA0 + (rgba.Rect.Dx() * rgba.Rect.Dy())),
		paletteData:   []byte{},
		imageData:     bytes.NewBuffer(rgba.Pix).Bytes(),
	}

	return image, nil
}

func (image *KRTImage) DecodeFromKRT() (*image.RGBA, error) {
	rgba, err := getTexture(image.width, image.height, image.imageFormat, image.imageData, image.paletteData)
	if err != nil {
		return nil, fmt.Errorf("error while decoding from KRTImage %v", err)
	}

	return rgba, nil
}

/*
	Supported Formats:
	RGB565
	RGB5A3
	RGBA8
	CI8
	CI4

	Needed Formats
	I4
	CMPR
*/
func getTexture(width uint32, height uint32, format uint32, data []byte, paletteData []byte) (*image.RGBA, error) {
	img := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))

	imagePixelIndex := 0
	imageDataIndex := 0

	if format == 0xF { //RGBA8
		blockWidth, blockHeight := 4, 4
		paletteImg := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))
		for i := 0; i < len(data) && imageDataIndex < len(data); i++ {
			tempBuf := data[imageDataIndex : imageDataIndex+64]
			imageDataIndex += 64
			for j := 0; j < 16 && imageDataIndex < len(data); j++ {

				blockSize := blockWidth * blockHeight
				blocksPerRow := int(width) / blockWidth
				block_i := imagePixelIndex % blockSize
				block_id := imagePixelIndex / blockSize
				blockCol := block_id % blocksPerRow
				blockRow := block_id / blocksPerRow
				Ix := blockCol*blockWidth + (block_i % blockWidth)
				Iy := blockRow*blockHeight + (block_i / blockWidth)

				paletteImg.Set(Ix, Iy, color.RGBA{
					R: uint8(tempBuf[1+j*2]),
					G: uint8(tempBuf[32+j*2]),
					B: uint8(tempBuf[33+j*2]),
					A: uint8(tempBuf[0+j*2]),
				})
				imagePixelIndex++
			}
		}

		return paletteImg, nil

	} else if format == 0x10 || format == 0x17 { //RGB565 or RGB5A3
		for y := 0; y < int(height); y++ {
			for x := 0; x < int(width); x++ {
				blockWidth, blockHeight := 4, 4
				pixel := binary.BigEndian.Uint16(data[imageDataIndex : imageDataIndex+2])

				blockSize := blockWidth * blockHeight
				blocksPerRow := int(width) / blockWidth
				block_i := imagePixelIndex % blockSize
				block_id := imagePixelIndex / blockSize
				blockCol := block_id % blocksPerRow
				blockRow := block_id / blocksPerRow
				Ix := blockCol*blockWidth + (block_i % blockWidth)
				Iy := blockRow*blockHeight + (block_i / blockWidth)

				if format == 0x17 {
					img.Set(Ix, Iy, color.RGBA{
						R: convert5to8(uint8((pixel >> 11))),
						G: convert6to8(uint8((pixel >> 5 & 0x3F))),
						B: convert5to8(uint8((pixel & 0x1F))),
						A: 255,
					})
				} else {
					hasAlpha := pixel & 0x8000
					if hasAlpha == 0 {
						img.Set(Ix, Iy, color.RGBA{
							R: convert4to8(uint8((pixel >> 8 & 0xF))),
							G: convert4to8(uint8((pixel >> 4 & 0xF))),
							B: convert4to8(uint8((pixel & 0xF))),
							A: convert3to8(uint8((pixel >> 12 & 0x7))),
						})
					} else {
						img.Set(Ix, Iy, color.RGBA{
							R: convert5to8(uint8((pixel >> 10 & 0x1F))),
							G: convert5to8(uint8((pixel >> 5 & 0x1F))),
							B: convert5to8(uint8((pixel & 0x1F))),
							A: 255,
						})
					}
				}

				imagePixelIndex++
				imageDataIndex += 2
			}
		}
		return img, nil
	} else if format == 0x11 || format == 0x12 { // CI4 / CI8 - RGB565 / RGB5A3
		var paletteEntries []*color.RGBA

		paletteDataIndex := 0

		for paletteDataIndex < len(paletteData) {
			pixel := binary.BigEndian.Uint16(paletteData[paletteDataIndex : paletteDataIndex+2])

			if format == 0x11 {
				paletteEntries = append(paletteEntries, &color.RGBA{
					R: convert5to8(uint8((pixel >> 11) & 0x1F)),
					G: convert6to8(uint8((pixel >> 5) & 0x3F)),
					B: convert5to8(uint8((pixel & 0x1F))),
					A: 255,
				})
			} else {
				hasAlpha := pixel & 0x8000
				if hasAlpha == 0 {
					paletteEntries = append(paletteEntries, &color.RGBA{
						R: convert4to8(uint8((pixel >> 8) & 0x0F)),
						G: convert4to8(uint8((pixel >> 4) & 0x0F)),
						B: convert4to8(uint8((pixel & 0x0F))),
						A: convert3to8(uint8((pixel >> 12) & 0x07)),
					})
				} else {
					paletteEntries = append(paletteEntries, &color.RGBA{
						R: convert5to8(uint8((pixel >> 10) & 0x1F)),
						G: convert5to8(uint8((pixel >> 5) & 0x1F)),
						B: convert5to8(uint8((pixel & 0x1F))),
						A: 255,
					})
				}
			}
			paletteDataIndex += 2
		}

		//Setup for
		bits := 0
		for y := 0; y < int(height); y++ {
			for x := 0; x < int(width); x++ {
				pixelImg := paletteEntries[uint8(data[imageDataIndex])]

				blockWidth, blockHeight := 8, 4

				blockSize := blockWidth * blockHeight
				blocksPerRow := int(width) / blockWidth
				block_i := imageDataIndex % blockSize
				block_id := imageDataIndex / blockSize
				blockCol := block_id % blocksPerRow
				blockRow := block_id / blocksPerRow
				Ix := blockCol*blockWidth + (block_i % blockWidth)
				Iy := blockRow*blockHeight + (block_i / blockWidth)

				//fmt.Printf("pixel - Index: %v, X: %v, Y: %v, Color: %v\n", uint8(texture.imageData[imageDataIndex]), Ix, Iy, pixelImg)

				img.Set(Ix, Iy, pixelImg)

				imageDataIndex++
				bits += 4

			}
		}
		return img, nil

	} else if format == 0x13 {
		var paletteEntries []*color.RGBA

		paletteDataIndex := 0

		for paletteDataIndex < len(paletteData) {
			pixel := binary.BigEndian.Uint16(paletteData[paletteDataIndex : paletteDataIndex+2])

			if format == 0x13 {
				paletteEntries = append(paletteEntries, &color.RGBA{
					R: convert5to8(uint8((pixel >> 11) & 0x1F)),
					G: convert6to8(uint8((pixel >> 5) & 0x3F)),
					B: convert5to8(uint8((pixel & 0x1F))),
					A: 255,
				})
			} else {
				hasAlpha := pixel & 0x8000
				if hasAlpha == 0 {
					paletteEntries = append(paletteEntries, &color.RGBA{
						R: convert4to8(uint8((pixel >> 8) & 0x0F)),
						G: convert4to8(uint8((pixel >> 4) & 0x0F)),
						B: convert4to8(uint8((pixel & 0x0F))),
						A: convert3to8(uint8((pixel >> 12) & 0x07)),
					})
				} else {
					paletteEntries = append(paletteEntries, &color.RGBA{
						R: convert5to8(uint8((pixel >> 10) & 0x1F)),
						G: convert5to8(uint8((pixel >> 5) & 0x1F)),
						B: convert5to8(uint8((pixel & 0x1F))),
						A: 255,
					})
				}
			}
			paletteDataIndex += 2
		}

		useSecondValue := false
		for y := 0; y < int(height); y++ {
			for x := 0; x < int(height); x++ {

				pixelImg := paletteEntries[uint8(data[imageDataIndex]>>4)]
				pixelImg2 := paletteEntries[uint8(data[imageDataIndex]&0xF)]

				blockWidth, blockHeight := 8, 8

				blockSize := blockWidth * blockHeight
				blocksPerRow := int(width) / blockWidth
				block_i := imagePixelIndex % blockSize
				block_id := imagePixelIndex / blockSize
				blockCol := block_id % blocksPerRow
				blockRow := block_id / blocksPerRow
				Ix := blockCol*blockWidth + (block_i % blockWidth)
				Iy := blockRow*blockHeight + (block_i / blockWidth)

				//fmt.Printf("pixel - Index: %v, X: %v, Y: %v, Color: %v\n", uint8(texture.imageData[imageDataIndex]), Ix, Iy, pixelImg)
				if useSecondValue {
					img.Set(Ix, Iy, pixelImg2)
					useSecondValue = false
					imageDataIndex++

				} else {
					img.Set(Ix, Iy, pixelImg)
					useSecondValue = true
				}

				imagePixelIndex++

			}
		}
		return img, nil
	} else if format == 0x16 {

		blockWidth, blockHeight := 8, 8
		useSecondValue := false
		for y := 0; y < int(height); y++ {
			for x := 0; x < int(height); x++ {

				pixelImg := uint8(data[imagePixelIndex] >> 4)
				pixelImg2 := uint8(data[imagePixelIndex] & 0xF)

				blockSize := blockWidth * blockHeight
				blocksPerRow := int(width) / blockWidth
				block_i := imagePixelIndex % blockSize
				block_id := imagePixelIndex / blockSize
				blockCol := block_id % blocksPerRow
				blockRow := block_id / blocksPerRow
				Ix := blockCol*blockWidth + (block_i % blockWidth)
				Iy := blockRow*blockHeight + (block_i / blockWidth)

				if useSecondValue {
					pixelImg2 = convert4to8(pixelImg2)
					img.Set(Ix, Iy, color.RGBA{
						R: pixelImg2,
						G: pixelImg2,
						B: pixelImg2,
						A: pixelImg2,
					})
					useSecondValue = false
					imageDataIndex++

				} else {
					pixelImg = convert4to8(pixelImg)
					img.Set(Ix, Iy, color.RGBA{
						R: pixelImg,
						G: pixelImg,
						B: pixelImg,
						A: pixelImg,
					})
					useSecondValue = true
				}

				imagePixelIndex++

			}
		}

		return img, nil
	}

	return nil, fmt.Errorf("getTexture is currently not implemented")
}

func convert3to8(v uint8) uint8 {
	return (v << 5) | (v << 2) | (v >> 1)
}

func convert4to8(v uint8) uint8 {
	return (v << 4) | v
}

func convert5to8(v uint8) uint8 {
	return (v << 3) | (v >> 2)
}

func convert6to8(v uint8) uint8 {
	return (v << 2) | (v >> 4)
}
