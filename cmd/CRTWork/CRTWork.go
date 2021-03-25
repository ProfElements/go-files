package main

import (
	"flag"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"strings"

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

	file, err := os.ReadFile(inputFile)

	if err != nil {
		fmt.Printf("Something went wrong wither opening %v, %v\n", inputFile, err)
	}

	parseData(file)

}

func parseData(data []byte) {

	file, err := crt.Read(data)
	if err != nil {
		fmt.Printf("Something went wrong while reading the data %v\n", err)
	}

	image, err := crt.Decode(file)
	if err != nil {
		fmt.Printf("Something went wrong while decoding the data into a image %v\n", err)
	}

	strPath := filepath.Base(inputFile)
	strPath = strings.Trim(strPath, filepath.Ext(strPath))

	imageFile, err := os.Create(strPath + ".png")

	if err != nil {
		fmt.Printf("Something went wrong while creating the image file %v\n", err)
	}

	err = png.Encode(imageFile, image)
	if err != nil {
		fmt.Printf("Something went wrong while encoding the file to png %v\n", err)
	}
	err = imageFile.Close()
	if err != nil {
		fmt.Printf("Something went wrong while trying to close the image file %v\n", err)
	}

}
