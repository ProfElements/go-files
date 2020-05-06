package jam

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type JAMHeader struct {
	Magic              string
	Unk1               uint32
	FileTableEndOffset uint32
	ArchiveNote        string
	FileNameCount      uint16
	FileExtCount       uint16
}

type FileEntry struct {
	FileNameIndex uint16
	FileExtIndex  uint16
	FileOffset    uint32
}

type JAMArchive struct {
	Header        JAMHeader
	FileNameTable []string
	FileExtTable  []string
	FileTable     []FileEntry
	FileData      []byte
}

func Read(data []byte) (*JAMArchive, error) {

	if !bytes.Equal(data[:4], []byte("FSTA")) && !bytes.Equal(data[:4], []byte("JAM2")) {
		return nil, fmt.Errorf("Missing file magic")
	}

	jam := &JAMArchive{
		Header: JAMHeader{},
	}

	jam.Header.Magic = string(data[:4])
	jam.Header.Unk1 = binary.LittleEndian.Uint32(data[4:8])
	jam.Header.FileTableEndOffset = binary.LittleEndian.Uint32(data[8:12])
	fileTableEndOffset := jam.Header.FileTableEndOffset
	jam.Header.ArchiveNote = string(data[12:28])
	jam.Header.FileNameCount = binary.LittleEndian.Uint16(data[28:30])
	fileNameCount := jam.Header.FileNameCount
	jam.Header.FileExtCount = binary.LittleEndian.Uint16(data[30:32])
	fileExtCount := jam.Header.FileExtCount

	var fileNames []string
	fileNames = make([]string, fileNameCount)

	var fileExts []string
	fileExts = make([]string, fileExtCount)

	idx := 32
	for i := 0; i < int(fileNameCount); i++ {
		fileNames[i] = string(data[idx : idx+8])
		idx = idx + 8
	}
	jam.FileNameTable = fileNames

	for i := 0; i < int(fileExtCount); i++ {
		fileExts[i] = string(data[idx : idx+4])
		idx = idx + 4
	}
	jam.FileExtTable = fileExts

	var fileEntries []FileEntry
	fileEntries = make([]FileEntry, (int(fileTableEndOffset)-idx)/8)

	for idx < int(fileTableEndOffset) {
		fileNameIndex := binary.LittleEndian.Uint16(data[idx : idx+2])
		fileExtIndex := binary.LittleEndian.Uint16(data[idx+2 : idx+4])
		fileOffset := binary.LittleEndian.Uint32(data[idx+4 : idx+8])
		fe := FileEntry{
			FileNameIndex: fileNameIndex,
			FileExtIndex:  fileExtIndex,
			FileOffset:    fileOffset,
		}
		fileEntries = append(fileEntries, fe)
		idx = idx + 8
	}

	jam.FileTable = fileEntries
	jam.FileData = data[idx:]

	return jam, nil
}

func Write(*JAMArchive) ([]byte, error) {
	return nil, fmt.Errorf("Unimplmented")
}
