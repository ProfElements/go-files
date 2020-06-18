package jam

/*
Name: JAM
Extension: .jam
Description: An archive used by a few high voltage gamecube/wii games.
             Its a very simple archive structure. The High Voltage
             games that use it are:
             The Grim Adventures of Billy and Mandy,
             Charlie and the Chocolate Factory,
             and Codename: Kids Next Door Operation V.I.D.E.O.G.A.M.E
*/

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/ProfElements/go-files/formats/tpl"
)

type fileEntry struct {
	fileNameIdx uint16
	fileExtIdx  uint16
	FileOffset  uint32
}
type Header struct {
	Magic              uint32
	unk1               uint32
	fileTableEndOffset uint32
	ArchiveNote        string
	fileNameCount      uint16
	fileExtCount       uint16
}
type File struct {
	Header        Header
	fileNameTable []string
	fileExtTable  []string
	FileTable     []fileEntry
	Files         []byte
	Data          []byte
}

//----------//
type WorkFile struct {
	FileName string
	FileExt  string
	Data     []byte
}

type Work struct {
	Files []WorkFile
}

func Read(data []byte) (*File, error) {
	jam := &File{}

	if len(data) < 32 {
		return nil, fmt.Errorf("Data size is too small to be a .jam file")
	}

	if !bytes.Equal(data[:4], []byte("FSTA")) && !bytes.Equal(data[:4], []byte("JAM2")) {
		return nil, fmt.Errorf("Wrong file magic, It isn't `FSTA` or `JAM2`")
	}

	jam.Header.Magic = binary.LittleEndian.Uint32(data[:4])
	jam.Header.unk1 = binary.LittleEndian.Uint32(data[4:8])
	jam.Header.fileTableEndOffset = binary.LittleEndian.Uint32(data[8:12])
	jam.Header.ArchiveNote = string(data[12:28])
	jam.Header.fileNameCount = binary.LittleEndian.Uint16(data[28:30])
	jam.Header.fileExtCount = binary.LittleEndian.Uint16(data[30:32])

	var fileNames []string
	var fileExts []string

	fileNames = make([]string, jam.Header.fileNameCount)
	fileExts = make([]string, jam.Header.fileExtCount)

	idx := 32
	for i := 0; i < len(fileNames); i++ {
		fileNames[i] = string(data[idx : idx+8])
		idx += 8
	}

	for i := 0; i < len(fileExts); i++ {
		fileExts[i] = string(data[idx : idx+4])
		idx += 4
	}

	jam.fileNameTable = fileNames
	jam.fileExtTable = fileExts

	var fileEntries []fileEntry
	fileEntries = make([]fileEntry, ((int(jam.Header.fileTableEndOffset)-idx)/8)+1)

	for i := 0; idx < int(jam.Header.fileTableEndOffset); i++ {
		fileEntry := fileEntry{
			fileNameIdx: binary.LittleEndian.Uint16(data[idx : idx+2]),
			fileExtIdx:  binary.LittleEndian.Uint16(data[idx+2 : idx+4]),
			FileOffset:  binary.LittleEndian.Uint32(data[idx+4 : idx+8]),
		}
		idx += 8
		fileEntries[i] = fileEntry
	}

	jam.FileTable = fileEntries
	jam.Files = data[idx:]
	jam.Data = data

	return jam, nil

}

func Write(data *File) ([]byte, error) {
	buffer := &bytes.Buffer{}
	_, err := buffer.WriteString("FSTA")
	if err != nil {
		return nil, err
	}

	err = binary.Write(buffer, binary.LittleEndian, uint32(data.Header.unk1))
	if err != nil {
		return nil, err
	}

	err = binary.Write(buffer, binary.LittleEndian, uint32(data.Header.fileTableEndOffset))
	if err != nil {
		return nil, err
	}

	_, err = buffer.WriteString("JMWK")
	if err != nil {
		return nil, err
	}

	err = binary.Write(buffer, binary.LittleEndian, []byte("\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00"))
	if err != nil {
		return nil, err
	}

	err = binary.Write(buffer, binary.LittleEndian, uint16(data.Header.fileNameCount))
	if err != nil {
		return nil, err
	}

	err = binary.Write(buffer, binary.LittleEndian, uint16(data.Header.fileExtCount))
	if err != nil {
		return nil, err
	}

	for idx := 0; idx < int(data.Header.fileNameCount); idx++ {
		err = binary.Write(buffer, binary.LittleEndian, binary.LittleEndian.Uint64([]byte(data.fileNameTable[idx])))
		if err != nil {
			return nil, err
		}
	}

	for idx := 0; idx < int(data.Header.fileExtCount); idx++ {
		err = binary.Write(buffer, binary.LittleEndian, binary.LittleEndian.Uint32([]byte(data.fileExtTable[idx])))
		if err != nil {
			return nil, err
		}
	}

	err = binary.Write(buffer, binary.LittleEndian, data.FileTable[:len(data.FileTable)-2])
	if err != nil {
		return nil, err
	}

	err = binary.Write(buffer, binary.LittleEndian, data.Files)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil

}
func Decode(data *File) (*Work, error) {
	work := &Work{}

	var workFiles []WorkFile
	workFiles = make([]WorkFile, len(data.FileTable))

	for i := 0; i < len(workFiles); i++ {
		if int(data.FileTable[i].fileNameIdx) > len(data.fileNameTable) || int(data.FileTable[i].fileExtIdx) > len(data.fileExtTable) || int(data.FileTable[i].fileExtIdx) == 0 || int(data.FileTable[i].FileOffset) < int(data.Header.fileTableEndOffset) || int(data.FileTable[i].FileOffset) > len(data.Data) {

		} else {
			tempExt := data.fileExtTable[data.FileTable[i].fileExtIdx]
			tempOffset := data.FileTable[i].FileOffset
			workFile := WorkFile{
				FileName: data.fileNameTable[data.FileTable[i].fileNameIdx],
				FileExt:  data.fileExtTable[data.FileTable[i].fileExtIdx],
				Data:     getData(tempExt, tempOffset, data),
			}
			workFiles[i] = workFile
		}
	}
	work.Files = workFiles
	return work, nil
}

func Encode(data *Work) (*File, error) {
	file := &File{}
	file.Header.Magic = binary.LittleEndian.Uint32([]byte("FSTA"))
	file.Header.unk1 = binary.LittleEndian.Uint32([]byte("\x00\x00\x00\x00"))
	file.Header.ArchiveNote = "JMWK"

}

//----------//
func getData(fileExt string, fileOffset uint32, data *File) []byte {

	switch Extension := strings.Trim(fileExt, "\x00"); Extension {
	case "GGG":
		fileSize := binary.BigEndian.Uint32(data.Data[fileOffset+12 : fileOffset+16])
		fileData := data.Data[fileOffset : fileOffset+fileSize]
		return fileData
	case "GKA":
		fileSize := binary.BigEndian.Uint32(data.Data[fileOffset+12 : fileOffset+16])
		fileData := data.Data[fileOffset : fileOffset+fileSize]
		return fileData
	case "GMD":
		fileSize := binary.BigEndian.Uint32(data.Data[fileOffset+12 : fileOffset+16])
		fileData := data.Data[fileOffset : fileOffset+fileSize]
		return fileData
	case "GMS":
		fileSize := binary.BigEndian.Uint32(data.Data[fileOffset+12 : fileOffset+16])
		fileData := data.Data[fileOffset : fileOffset+fileSize]
		return fileData
	case "GSL":
		fileSize := binary.BigEndian.Uint32(data.Data[fileOffset+20 : fileOffset+24])
		fileData := data.Data[fileOffset : fileOffset+fileSize]
		return fileData
	case "TGA":
		return nil
	case "TPL":
		tpl, err := tpl.Read(data.Data[fileOffset:])
		if err != nil {
			fmt.Printf("SOmething went wrong")
		}
		if tpl == nil {
			return []byte{}
		}
		return tpl.Data
		//return nil
	default:
		idx := fileOffset
		for int(idx) != len(data.Data) && data.Data[idx] != byte(0xFF) || int(fileOffset) > len(data.Data) {
			idx++
		}
		fileData := data.Data[fileOffset:idx]
		return fileData
	}

}
