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

	file, err := crt.ReadKRT(inputFile)

	if err != nil {
		fmt.Printf("Something went wrong wither opening %v, %v\n", inputFile, err)
	}

	rgba, err := file.DecodeFromKRT()
	if err != nil {
		fmt.Printf("Something went wrong with decoding the KRTImage to rgba %v", err)
	}

	strPath := filepath.Base(inputFile)
	strPath = strings.Trim(strPath, filepath.Ext(strPath))

	rgbaFile, err := os.Create(strPath + ".png")
	if err != nil {
		fmt.Printf("Something went wrong with creating the png file %v", err)
	}
	defer rgbaFile.Close()

	err = png.Encode(rgbaFile, rgba)
	if err != nil {
		fmt.Printf("Something when wrong with encoding the png file %v", err)
	}

}
