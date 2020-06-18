package main

/*
NAME:JAMWork
DESCRIPTION:A simple packer and unpacked for .jam files
USAGE: ./JAMWork.exe -p/-u <file/folder>
*/

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/ProfElements/go-files/formats/jam"
)

func main() {
	args := os.Args[1:]
	//args[0] is the .jam file input
	//args[1] should be one of four values, encode, decode, read, or write
	//args[2] should be output
	file, err := ioutil.ReadFile(args[0])
	if err != nil {
		fmt.Printf("Something went wrong.")
	}
	switch operation := args[1]; operation {
	case "read":
		jam, err := jam.Read(file)
		if err != nil {
			fmt.Printf("Something went wrong.")
		}
		print(jam.Header.Magic)
	case "decode":

		jamFile, err := jam.Read(file)
		if err != nil {
			fmt.Printf("Something went wrong.")
		}
		file, err := jam.Decode(jamFile)
		if err != nil {
			fmt.Printf("Something went wrong.")
		}
		for i := 0; i < len(file.Files); i++ {
			if file.Files[i].FileName == "" || file.Files[i].FileExt == "" || (strings.Trim(file.Files[i].FileExt, "\x00") == "TPL" && len(file.Files[i].Data) == 0) {

			} else {
				fmt.Printf(strings.Trim(file.Files[i].FileName, "\x00")+"."+strings.Trim(file.Files[i].FileExt, "\x00")+" Offset: %x"+" DataSize: %x"+"\n", jamFile.FileTable[i].FileOffset, len(file.Files[i].Data))
				export, err := os.Create(strings.Trim(file.Files[i].FileName, "\x00") + "." + strings.Trim(file.Files[i].FileExt, "\x00"))
				if err != nil {
					fmt.Printf("Something went wrong")
				}
				export.Write(file.Files[i].Data)
				export.Close()
			}
		}
	case "write":
		jamFile, err := jam.Read(file)
		if err != nil {
			fmt.Printf("Something went wrong.")
		}
		file, err := jam.Write(jamFile)
		if err != nil {
			fmt.Printf("%v", err)
		}
		err = ioutil.WriteFile("file.jam", file, 0777)
		if err != nil {
			fmt.Printf("I Fucked up.")
		}
	}
}
