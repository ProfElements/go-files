package crt

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
)

/*
NAME: Creature From the Krust Krab Texture `KRT`
EXTENSION: .krt or none.
DESCRIPTION: The headerless texture format used in gcp archives in Spongebob Squarepants: Creature From the Krust Krab
*/

/*Binary Structure
padding []byte 0x20 in size of zeroes
width          uint32
height         uint32
unknown1       uint32 //Probably Pixel Format
unknown2       uint16
unknown3       uint8
unknown4       uint8
unknown5       uint32 looks to be bit flags
unknown6       uint32 looks to be bit flags aswell
imageDataSize  uint32 (width*height)
padding        []byte 0x30 in size of zeroes
metaDataOffset uint32 reads until imageDataStart
imageDataStartOffset uint32
imageDataEndOffset   uint32
padding[]      []byte padding until imageDataStart
*/

type KRTTexture struct {
	width       uint32
	height      uint32
	imageFormat uint32
	blockSize   uint16
	mipMapData  []byte
	imageData   []byte
}

func Read(Data []byte) (*KRTTexture, error) {
	texture := &KRTTexture{}

	index := 0

	if len(Data) < 0xA0 {
		return nil, fmt.Errorf("length of data is not long enough to be considered a KRT")
	}

	index += 0x20

	texture.width = binary.BigEndian.Uint32(Data[index : index+4])

	index += 4

	texture.height = binary.BigEndian.Uint32(Data[index : index+4])

	index += 4

	texture.imageFormat = binary.BigEndian.Uint32(Data[index : index+4])
	fmt.Printf("%v\n", texture.imageFormat)

	index += 4

	texture.blockSize = binary.BigEndian.Uint16(Data[index : index+2])

	index += 12

	if texture.width*texture.height != binary.BigEndian.Uint32(Data[index:index+4]) {
		return nil, fmt.Errorf("wrong height width this is wrongggg %v, %v, %v", texture.width, texture.height, binary.BigEndian.Uint32(Data[index:index+4]))

	}

	index += 0x30
	mipMapOffset := binary.BigEndian.Uint32(Data[index : index+4])
	index += 4
	imageDataOffset := binary.BigEndian.Uint32(Data[index : index+4])
	fmt.Printf("\n%x\n", imageDataOffset)

	if mipMapOffset != 0 {
		texture.mipMapData = Data[mipMapOffset:imageDataOffset]
	}
	texture.imageData = Data[imageDataOffset:]

	return texture, nil
}

func Decode(data *KRTTexture) (*image.RGBA, error) {

	//if data.imageFormat != 16 {
	//	return nil, fmt.Errorf("does not currently support any other pixel formats besides 16 format: %v", data.imageFormat)
	//}

	imageDataIndex := 0
	imgIndex := 0
	imageDataPixel := getPixelFromTextureFormat(data.imageFormat, data.imageData, imageDataIndex)
	if imageDataPixel == nil {
		return nil, fmt.Errorf("texture format not recognized or not yet implemented")
	}
	img := image.NewRGBA(image.Rect(0, 0, int(data.width), int(data.height)))

	//blockWidth := 4
	//blockHeight := 4

	blockWidth, blockHeight := getBlockSizeFromTextureFormat(data.imageFormat)

	for y := 0; y < int(data.height); y++ {
		for x := 0; x < int(data.width); x++ {

			blockSize := blockWidth * blockHeight
			blocksPerRow := int(data.width) / blockWidth

			block_i := imgIndex % blockSize
			block_id := imgIndex / blockSize
			blockCol := block_id % blocksPerRow
			blockRow := block_id / blocksPerRow
			Ix := blockCol*blockWidth + (block_i % blockWidth)
			Iy := blockRow*blockHeight + (block_i / blockWidth)

			var imageDataPixel32 uint32

			switch v := imageDataPixel.(type) {
			case uint8:
				imageDataPixel32 = uint32(imageDataPixel.(uint8))
				_ = v
			case uint16:
				imageDataPixel32 = uint32(imageDataPixel.(uint16))
			case uint32:
				imageDataPixel32 = imageDataPixel.(uint32)
			}

			pixelColor := getColorFromTextureFormat(data.imageFormat, imageDataPixel32)
			if pixelColor == nil {
				pixelColor = getColorFromTexturePalette(data.imageFormat, data.mipMapData, imageDataPixel32)
			}

			img.Set(Ix, Iy, pixelColor)
			imageDataIndex += 2
			imgIndex++
			if imageDataIndex < len(data.imageData) {
				switch v := getPixelFromTextureFormat(data.imageFormat, data.imageData, imageDataIndex).(type) {
				case uint8:
					imageDataPixel = uint32(getPixelFromTextureFormat(data.imageFormat, data.imageData, imageDataIndex).(uint8))
					_ = v
				case uint16:
					imageDataPixel = uint32(getPixelFromTextureFormat(data.imageFormat, data.imageData, imageDataIndex).(uint16))
				case uint32:
					imageDataPixel = getPixelFromTextureFormat(data.imageFormat, data.imageData, imageDataIndex).(uint32)
				}
			}

		}
	}

	//img, err := unswizzleImg(data.imageFormat, img)
	// err != nil {
	//	fmt.Printf("Something went wrong while unswizzling the image")
	//}

	return img, nil
}

/*
func read(data []byte) (*KRTTexture, error) { return nil, nil }
func write(*File) ([]byte, error)           { return nil, nil }
func decode(*File) *Work                    { return nil }
func encode(*Work) *KRTTexture              { return nil }
*/

/*
//#define GX_TF_I4			0x0
//#define GX_TF_I8			0x1
//#define GX_TF_IA4			0x2
//#define GX_TF_IA8			0x3

//#define GX_TF_RGB5A3		0x5
//#define GX_TF_RGBA8	    0x6
//#define GX_TF_CI4			0x8
//#define GX_TF_CI8			0x9
//#define GX_TF_CI14		0xa

//#define GX_TL_IA8			0x00
//#define GX_TL_RGB565		0x01
//#define GX_TL_RGB5A3		0x02
*/
func getBlockSizeFromTextureFormat(imageFormat uint32) (width int, height int) {
	switch imageFormat {
	case 0xf:
		// GX_TF_RGBA8	    0x6
		return 4, 4
	case 0x10:
		//GX_TF_RGB5A3		0x5
		return 4, 4
	case 0x11:
		//GX_TF_CI8			0x9
		return 8, 4
	case 0x12:
		//GX_TF_CI8			0x9
		return 8, 4
	case 0x13:
		//Says it just a palette or something :shrug:
		return 0, 0
	case 0x15:
		//GX_TF_CMPR		0xE
		return 8, 8
	case 0x16:
		// GX_TF_I4			0x0
		return 8, 8
	case 0x17:
		//GX_TF_RGB565		0x4
		return 4, 4
	default:
		return 0, 0
	}
}

func getColorFromTextureFormat(imageFormat uint32, pixel uint32) *color.RGBA {
	switch imageFormat {
	case 0xf:
		// GX_TF_RGBA8	    0x6
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.LittleEndian, pixel)
		colorData := buf.Bytes()

		return &color.RGBA{
			R: colorData[0],
			G: colorData[1],
			B: colorData[2],
			A: colorData[3],
		}
	case 0x10:
		//GX_TF_RGB5A3		0x5
		return &color.RGBA{
			R: uint8(((pixel & 0x7C00) >> 10) << 3),
			G: uint8(((pixel & 0x3E0) >> 5) << 3),
			B: uint8((pixel & 0x1F) << 3),
			A: 255,
		}
	case 0x11:
		//GX_TF_CI8			0x9
		return nil
	case 0x12:
		//GX_TF_CI8			0x9
		return nil
	case 0x13:
		return nil
		//Says it just a palette or something :shrug:
	case 0x15:
		return nil
	case 0x16:
		// GX_TF_I4			0x0
		return nil
	case 0x17:
		//GX_TF_RGB565		0x4
		return &color.RGBA{
			R: uint8(((pixel & 0x7C00) >> 10) << 3),
			G: uint8(((pixel & 0x3E0) >> 5) << 3),
			B: uint8((pixel & 0x1F) << 3),
			A: 255,
		}
	default:
		return &color.RGBA{
			R: 255,
			G: 000,
			B: 255,
			A: 255,
		}

	}

}

func getPixelFromTextureFormat(imageFormat uint32, imageData []byte, imageDataOffset int) (interfacing interface{}) {
	switch imageFormat {
	case 0xf:
		_, ok := interfacing.(uint32)
		if !ok {
			fmt.Printf("INTERFACE DOING SOME SHIT")
		}

		return binary.BigEndian.Uint32(imageData[imageDataOffset : imageDataOffset+4])
	case 0x10:
		//GX_TF_RGB5A3		0x5
		return binary.BigEndian.Uint16(imageData[imageDataOffset : imageDataOffset+2])
	case 0x11:
		//GX_TF_CI8			0x9
		return uint8(imageData[imageDataOffset])
	case 0x12:
		//GX_TF_CI8			0x9
		return uint8(imageData[imageDataOffset])
	case 0x13:
		return nil
	case 0x15:
		//GX_TF_CMPR		0xE
		return uint8(imageData[imageDataOffset])
	case 0x16:
		// GX_TF_I4			0x0
		return uint8(imageData[imageDataOffset])
	case 0x17:
		//GX_TF_RGB565		0x4
		return binary.BigEndian.Uint16(imageData[imageDataOffset : imageDataOffset+2])
	default:
		return nil
	}
}

func getColorFromTexturePalette(imageFormat uint32, paletteData []byte, imagePixel interface{}) *color.RGBA {
	return &color.RGBA{
		R: 255,
		G: 000,
		B: 255,
		A: 255,
	}
}
