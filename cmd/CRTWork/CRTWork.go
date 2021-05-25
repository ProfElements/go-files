package main

import (
	"flag"
	"fmt"
	"image/color"
	"image/png"
	"image"
	"os"
	"path/filepath"
	"strings"
	"bytes"
	"github.com/ProfElements/go-files/pkg/formats/cftkk/crt"
)

/*
NAME:
DESCRIPTION:
USAGE:
*/
var inputFile string
var outputFile string

func main() {

	if len(os.Args) < 2 {
		panic("We need atleast one krt file to continue.")
	}

	flag.StringVar(&inputFile, "inputfile", "", " Input file path pointing to a texture")

	flag.StringVar(&outputFile, "outputfile", "", "Output file path for the resultant png")

	flag.Parse()

	if inputFile == "" {
		inputFile = os.Args[1]
	}

	if filepath.Ext(inputFile) != "" && filepath.Ext(inputFile) != ".krt" && filepath.Ext(inputFile) != ".png" {
		panic("We need either a png or krt file to continue")
	}
	
	strPath := filepath.Base(inputFile)
	strPath = strings.Trim(strPath, filepath.Ext(strPath))

	if filepath.Ext(inputFile) == ".krt" || filepath.Ext(inputFile) == "" {
		file, err := crt.ReadKRT(inputFile)

		if err != nil {
			fmt.Printf("Something went wrong wither opening %v, %v\n", inputFile, err)
		}

		rgba, err := file.DecodeFromKRT()
		if err != nil {
			fmt.Printf("Something went wrong with decoding the KRTImage to rgba %v", err)
		}

		rgbaFile, err := os.Create(strPath + ".png")
		if err != nil {
			fmt.Printf("Something went wrong with creating the png file %v", err)
		}
		defer rgbaFile.Close()

		err = png.Encode(rgbaFile, rgba)
		if err != nil {
			fmt.Printf("Something when wrong with encoding the png file %v", err)
		}

	} else {
		raw, err := os.ReadFile(inputFile)
		if err != nil {
			fmt.Printf("something went wrong while reading the file")
		}
		rgba, err := png.Decode(bytes.NewBuffer(raw))
		if err != nil { fmt.Printf("Decoding png did something wrong") }			
		
		if rgba.ColorModel() != color.RGBAModel {
			img := image.NewRGBA(image.Rect(0, 0, rgba.Bounds().Dx(), rgba.Bounds().Dy()))
			//Convert to RGBA8
			for y := rgba.Bounds().Min.Y; y < rgba.Bounds().Max.Y; y++ {
				for x := rgba.Bounds().Min.X; x < rgba.Bounds().Max.X; x++ {
					c := color.RGBAModel.Convert(rgba.At(x, y)).(color.RGBA)
					

					img.Set(x, y, c)
				}
			}
			
			krtimage, err := crt.EncodeToKRT(img)
			if err != nil {
				fmt.Printf("Something happened while encoding png to krt")
			}

			err = krtimage.WriteKRT(strPath + ".krt")
			if err != nil {
				fmt.Printf("WRITING THE THING DIDNT WORKRKKKK")
			}
	  }
				
		pngImage, ok := rgba.(*image.RGBA) 
		if !ok {
			panic("PANICCING BECAUSE NOT RGBA")
		}

		krtimage, err := crt.EncodeToKRT(pngImage)
		if err != nil { fmt.Printf("Something wrong happened while encoding to krt") }
				
		err = krtimage.WriteKRT(strPath + ".krt")
		if err != nil {
			fmt.Printf("WRITING THE THING DIDNT WORKRKKKK")
		}
	}
}
