package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/ProfElements/go-files/formats/jam"
)

/*
NAME: JAMWork
DESCRIPTION: A basic manipulation tool for .jam archives.
USAGE: JAMwork [-u, -p, --unpack, --pack] [JAM_ARCHIVE, JAM_DIRECTORY]
*/

func main() {
	args := os.Args[1:]

	/*
		args[0] should be one of these: -u, -p, --unpack, --pack
		args[1] should be a path to either a jam archive, or a directory
	*/

	switch operation := args[0]; operation {
	case "-u":
		unpack()
	case "--unpack":
		unpack()
	case "-p":
		workFile := &jam.Work{}
		files, err := ioutil.ReadDir(args[1])
		if err != nil {
			fmt.Printf("Reading directory failed.")

		}
		fileData := make([]jam.WorkFile, len(files))
		for i, f := range files {
			filePath := args[1] + strings.TrimSuffix(f.Name(), filepath.Ext(f.Name())) + filepath.Ext(f.Name())
			if err != nil {
				fmt.Printf("getting absolute path of files didnt work.")
			}

			binFile, err := ioutil.ReadFile(filePath)
			if err != nil {
				fmt.Printf("Reading file failed. %v \n", err)
			}

			fileData[i].FileName = strings.TrimSuffix(f.Name(), filepath.Ext(f.Name()))
			fileData[i].FileExt = strings.Trim(filepath.Ext(f.Name()), ".")
			fileData[i].Data = binFile
		}
		workFile.Files = fileData
		jamFile, err := jam.Encode(workFile)
		if err != nil {
			fmt.Printf("Encoding of data did not work.")
		}

		file, err := jam.Write(jamFile)
		if err != nil {
			fmt.Printf("Writing of jam file did not work")
		}

		err = ioutil.WriteFile("test.jam", file, 0644)
	}

}

func unpack() {
	args := os.Args[1:]
	directory := strings.TrimSuffix(filepath.Base(args[1]), filepath.Ext(filepath.Base(args[1])))

	file, err := ioutil.ReadFile(args[1])
	if err != nil {
		fmt.Printf("Loading of the jam archive didn't work.")
	}
	jamFile, err := jam.Read(file)
	if err != nil {
		fmt.Printf("Reading of the jam archive didn't work.")
	}
	workFile, err := jam.Decode(jamFile)
	if err != nil {
		fmt.Printf("Decoding of the jam archive didn't work.")
	}

	_ = os.Mkdir(directory, 0644)
	for i := 0; i < len(workFile.Files); i++ {
		if workFile.Files[i].FileName == "" || workFile.Files[i].FileExt == "nil" || len(workFile.Files[i].Data) == 0 {

		} else {

			output_file := filepath.Join(directory, strings.Trim(workFile.Files[i].FileName, "\x00")+"."+strings.Trim(workFile.Files[i].FileExt, "\x00"))

			fmt.Printf("File: "+strings.Trim(workFile.Files[i].FileName, "\x00")+"."+strings.Trim(workFile.Files[i].FileExt, "\x00")+" Data Size: %x"+"\n", len(workFile.Files[i].Data))

			err := ioutil.WriteFile(output_file, workFile.Files[i].Data, 0644)
			if err != nil {
				fmt.Printf("Writing the extracted files to the output path didnt work :(  %v\n ", err)
			}
		}
	}
}
