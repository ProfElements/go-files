package crt

import (
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

func Decode(data *KRTTexture) (*image.NRGBA, error) {

	//if data.imageFormat != 16 {
	//	return nil, fmt.Errorf("does not currently support any other pixel formats besides 16 format: %v", data.imageFormat)
	//}

	imageDataIndex := 0
	imgIndex := 0
	imageDataPixel := binary.BigEndian.Uint16(data.imageData[imageDataIndex : imageDataIndex+2])

	img := image.NewNRGBA(image.Rect(0, 0, int(data.width), int(data.height)))

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

			pixelColor := getColorFromTextureFormat(data.imageFormat, imageDataPixel)

			img.Set(Ix, Iy, pixelColor)
			imageDataIndex += 2
			imgIndex++
			if imageDataIndex < len(data.imageData) {
				imageDataPixel = binary.BigEndian.Uint16(data.imageData[imageDataIndex : imageDataIndex+2])
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

func getColorFromTextureFormat(imageFormat uint32, pixel uint16) color.NRGBA {
	switch imageFormat {
	case 0xf:
		// GX_TF_RGBA8	    0x6
	case 0x10:
		//GX_TF_RGB5A3		0x5
		return color.NRGBA{
			R: uint8(((pixel & 0x7C00) >> 10) << 3),
			G: uint8(((pixel & 0x3E0) >> 5) << 3),
			B: uint8((pixel & 0x1F) << 3),
			A: 255,
		}
	case 0x11:
		//GX_TF_CI8			0x9
	case 0x12:
		//GX_TF_CI8			0x9
	case 0x13:
		//Says it just a palette or something :shrug:
	case 0x15:
		//GX_TF_CMPR		0xE
	case 0x16:
		// GX_TF_I4			0x0
	case 0x17:
		//GX_TF_RGB565		0x4
		return color.NRGBA{
			R: uint8(((pixel & 0x7C00) >> 10) << 3),
			G: uint8(((pixel & 0x3E0) >> 5) << 3),
			B: uint8((pixel & 0x1F) << 3),
			A: 255,
		}
	default:
		return color.NRGBA{
			R: 255,
			G: 255,
			B: 255,
			A: 255,
		}

	}
	return color.NRGBA{
		R: 255,
		G: 255,
		B: 255,
		A: 255,
	}
}
