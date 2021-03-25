package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ProfElements/go-files/pkg/formats/cftkk/fetm"
)

/*
NAME:
DESCRIPTION:
USAGE:
*/
var inputFile string
var outputFile string
var isJson bool
var isRaw bool

func main() {

	if len(os.Args) < 2 {
		panic("We need atleast one fetm file to continue.")
	}

	flag.StringVar(&inputFile, "inputfile", "", " Input file path pointing to either a fetm file or a json file.")

	flag.StringVar(&outputFile, "outputfile", "", "Output file path for the resultant fetm or json file.")

	flag.BoolVar(&isRaw, "raw", true, "A flag to specify whether to output raw or constructed json")
	flag.BoolVar(&isJson, "json", false, "Whether to convert json to fetm")

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

	var file *fetm.FETM
	var err error

	if isJson {
		file, err = fetm.DecodeFromJSON(data, isRaw)
		data, err = file.Write()
		if err != nil {
			fmt.Printf("something happened when encoding to json")
		}

		strPath := filepath.Base(inputFile)
		strPath = strings.Trim(strPath, filepath.Ext(strPath))
		fmt.Printf(strPath + ".fetm")

		if outputFile == "" {
			os.WriteFile(strPath+".fetm", data, 777)
		} else {
			os.WriteFile(outputFile, data, 777)
		}

	} else {
		file, err = fetm.Read(data)
		json, err := file.EncodeToJSON(isRaw)

		if err != nil {
			fmt.Printf("something happened when encoding to json")
		}

		strPath := filepath.Base(inputFile)
		strPath = strings.TrimSuffix(strPath, filepath.Ext(strPath))
		fmt.Printf(strPath + ".json")

		if outputFile == "" {
			os.WriteFile(strPath+".json", json, 777)
		} else {
			os.WriteFile(outputFile, json, 777)
		}
	}

	if err != nil {
		fmt.Print(err)
	}

}
