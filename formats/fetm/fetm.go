package fetm

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"
)

type identifer uint8

const (
	s8  identifer = 0x00
	u8  identifer = 0x01
	s16 identifer = 0x02
	u16 identifer = 0x03
	u32 identifer = 0x04
	hex identifer = 0x05
	f32 identifer = 0x06
	str identifer = 0x07
)

//Make this structure more workable for encoding to json.
type variantData struct {
	Ident identifer
	Data  interface{}
}

type header struct {
	Magic        variantData
	Note         variantData
	SectionCount variantData
	FileSize     variantData
	Raw          []variantData
}

type world struct {
	Raw []variantData
}

type sector struct {
	Raw []variantData
}

type node struct{}

/*
FETM ...
FETM file structure to read and write to and from files.
*/
type FETM struct {
	Header header
	World  world
	Sector sector
	Nodes  []variantData //make this nodes later
	Raw    []variantData
}

/*
Read ...
Read data from file and return a FETM structure pointer, or error
*/
func Read(data []byte) (*FETM, error) {
	fetm := &FETM{}

	if len(data) < 3 {
		return nil, fmt.Errorf("data is not long enough to be considered a FETM")
	}

	if !bytes.Equal(data[:3], []byte{0x01, 0x7C, 0x07}) {
		return nil, fmt.Errorf("this file does not have FETM header magic")
	}

	readIndex := 0

	for len(data) > readIndex {

		if uint8(data[readIndex]) > 7 {
			return nil, fmt.Errorf("unkown identifer from file")
		}

		curIdent := identifer(uint8(data[readIndex]))
		readIndex++

		switch curIdent {
		case s8:
			fetm.Raw = append(fetm.Raw, variantData{curIdent, int8(data[readIndex])})
			readIndex++
		case u8:
			fetm.Raw = append(fetm.Raw, variantData{curIdent, uint8(data[readIndex])})
			readIndex++
		case s16:
			var val int16
			buf := bytes.NewReader(data[readIndex : readIndex+1])
			binary.Read(buf, binary.BigEndian, &val)

			fetm.Raw = append(fetm.Raw, variantData{curIdent, val})
			readIndex++
		case u16:
			var val uint16
			buf := bytes.NewReader(data[readIndex : readIndex+1])
			binary.Read(buf, binary.BigEndian, &val)

			fetm.Raw = append(fetm.Raw, variantData{curIdent, val})
			readIndex++
		case u32:
			var val uint32
			buf := bytes.NewReader(data[readIndex : readIndex+3])
			binary.Read(buf, binary.BigEndian, &val)

			fetm.Raw = append(fetm.Raw, variantData{curIdent, val})
			readIndex += 3
		case f32:
			var val float32
			buf := bytes.NewReader(data[readIndex : readIndex+3])
			binary.Read(buf, binary.BigEndian, &val)

			fetm.Raw = append(fetm.Raw, variantData{curIdent, val})
		case str:
			fetm.Raw = append(fetm.Raw, variantData{curIdent, findStr(data, readIndex)})
		default:
			return nil, fmt.Errorf("unkown identifer from file")
		}

	}

	fetm.Header.Magic = fetm.Raw[0]
	fetm.Header.Note = fetm.Raw[1]
	fetm.Header.SectionCount = fetm.Raw[2]
	fetm.Header.FileSize = fetm.Raw[3]
	fetm.Header.Raw = fetm.Raw[0:3]

	if reflect.TypeOf(fetm.Raw[4].Data).String() == "string" && fetm.Raw[4].Data.(string) == "world" {
		for idx, value := range fetm.Raw {
			if reflect.TypeOf(value.Data).String() == "string" && value.Data.(string) == "sector" {
				fetm.World.Raw = fetm.Raw[4 : idx-1]
				//This includes nodes as of right due to not being able to look at source to figure out how they find nodes
				fetm.Sector.Raw = fetm.Raw[idx:]
			}
		}
	}

	return fetm, nil
}

/*
Write ...
Write FETM data to file or error
*/
func (data *FETM) Write() ([]byte, error) {
	return nil, fmt.Errorf("write is not currently implmented")
}

/*
DecodeFromJSON ...
Read data from a json file and decode it to FETM structure or error
*/
func DecodeFromJSON(data []byte) (*FETM, error) {
	return nil, fmt.Errorf("DecodeFromJSON is not currently implemented")
}

/*
EncodeToJSON ...
Encode and write FETM data to JSON file or error
*/
func (data *FETM) EncodeToJSON() ([]byte, error) {
	return nil, fmt.Errorf("EncodeToJSON is not currently implemented")
}

func findStr(data []byte, startIndex int) string {
	curIndex := startIndex
	for len(data) > curIndex {
		if data[curIndex] == 0x00 {
			return string(data[startIndex:curIndex])
		}
		curIndex++
	}

	return ""
}
