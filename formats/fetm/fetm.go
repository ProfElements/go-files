package fetm

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
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
	World variantData
	Raw   []variantData
}

type sector struct {
	Raw []variantData
}

type node struct {
	NodeType      variantData
	NodeFactory   variantData
	EntityClass   variantData
	DependentData []variantData
}

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
		return nil, fmt.Errorf("data is not long enough to be considered a FETM\n")
	}

	if !bytes.Equal(data[:3], []byte{0x01, 0x7C, 0x07}) {
		return nil, fmt.Errorf("this file does not have FETM header magic\n")
	}

	readIndex := 0

	for len(data) > readIndex {

		curIdent := identifer(uint8(data[readIndex]))

		//fmt.Printf("curIdent %v\n", curIdent)

		switch curIdent {
		case s8:
			fetm.Raw = append(fetm.Raw, variantData{curIdent, int8(data[readIndex+1])})
			readIndex += 2
		case u8:
			fetm.Raw = append(fetm.Raw, variantData{curIdent, uint8(data[readIndex+1])})
			readIndex += 2
		case s16:
			var val int16
			buf := bytes.NewReader(data[readIndex : readIndex+1])
			binary.Read(buf, binary.BigEndian, &val)

			fetm.Raw = append(fetm.Raw, variantData{curIdent, val})
			readIndex += 2
		case u16:
			var val uint16
			buf := bytes.NewReader(data[readIndex+1 : readIndex+3])
			binary.Read(buf, binary.BigEndian, &val)

			fetm.Raw = append(fetm.Raw, variantData{curIdent, val})
			readIndex += 3
		case u32:
			var val uint32
			buf := bytes.NewReader(data[readIndex+1 : readIndex+5])
			binary.Read(buf, binary.BigEndian, &val)

			fetm.Raw = append(fetm.Raw, variantData{curIdent, val})
			readIndex += 5
		case hex:
			var val uint32
			buf := bytes.NewReader(data[readIndex+1 : readIndex+5])
			binary.Read(buf, binary.BigEndian, &val)

			fetm.Raw = append(fetm.Raw, variantData{curIdent, val})
			readIndex += 5
		case f32:
			var val float32
			buf := bytes.NewReader(data[readIndex+1 : readIndex+5])
			binary.Read(buf, binary.BigEndian, &val)

			fetm.Raw = append(fetm.Raw, variantData{curIdent, val})
			readIndex += 5
		case str:
			strData := findStr(data, readIndex)

			fetm.Raw = append(fetm.Raw, variantData{curIdent, strData})
			if len(strData) == 1 {
				readIndex++
			} else {
				readIndex += len(strData) + 2
			}

		default:
			return nil, fmt.Errorf("unknown idenifier %v at %x", curIdent, readIndex)
		}

	}

	fetm.Header.Magic = fetm.Raw[0]
	fetm.Header.Note = fetm.Raw[1]
	fetm.Header.SectionCount = fetm.Raw[2]
	fetm.Header.FileSize = fetm.Raw[3]
	fetm.Header.Raw = fetm.Raw[0:3]

	if reflect.TypeOf(fetm.Raw[4].Data).String() == "string" && fetm.Raw[4].Data.(string) == "world" {
		fetm.World.World = fetm.Raw[4]

		for idx, value := range fetm.Raw {
			if reflect.TypeOf(value.Data).String() == "string" && value.Data.(string) == "World Sector" {
				fetm.World.Raw = fetm.Raw[4 : idx-2]
				//This includes nodes as of right due to not being able to look at source to figure out how they find nodes
				fetm.Sector.Raw = fetm.Raw[idx-2:]
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
	var buffer bytes.Buffer

	for _, value := range data.Raw {
		binary.Write(&buffer, binary.BigEndian, byte(value.Ident))
		switch value.Ident {
		case s8:
			binary.Write(&buffer, binary.BigEndian, int8(value.Data.(float64)))
		case u8:
			binary.Write(&buffer, binary.BigEndian, uint8(value.Data.(float64)))
		case s16:
			binary.Write(&buffer, binary.BigEndian, int16(value.Data.(float64)))
		case u16:
			binary.Write(&buffer, binary.BigEndian, uint16(value.Data.(float64)))
		case u32:
			binary.Write(&buffer, binary.BigEndian, uint32(value.Data.(float64)))
		case hex:
			hex, _ := strconv.ParseUint(value.Data.(string), 10, 0)
			binary.Write(&buffer, binary.BigEndian, uint32(hex))
		case f32:
			binary.Write(&buffer, binary.BigEndian, float32(value.Data.(float64)))
		case str:
			binary.Write(&buffer, binary.BigEndian, []byte(value.Data.(string)))
			binary.Write(&buffer, binary.BigEndian, byte(0x00))
		default:
		}
	}

	return buffer.Bytes(), nil
}

/*
DecodeFromJSON ...
Read data from a json file and decode it to FETM structure or error
*/
func DecodeFromJSON(data []byte, isRaw bool) (*FETM, error) {
	fetm := FETM{}

	if isRaw {
		err := json.Unmarshal(data, &fetm.Raw)
		if err != nil {
			return nil, fmt.Errorf("something went wrong when decoding from json %v", err)
		}
	} else {
		err := json.Unmarshal(data, &fetm)
		if err != nil {
			return nil, fmt.Errorf("something went wrong when decoding from json %v", err)
		}
	}

	return &fetm, nil
}

/*
EncodeToJSON ...
Encode and write FETM data to JSON file or error
*/
func (data *FETM) EncodeToJSON(isRaw bool) ([]byte, error) {
	if isRaw {
		raw, err := json.Marshal(data.Raw)
		if err != nil {
			return nil, fmt.Errorf("something went wrong when encoding the json")
		}
		return raw, nil
	} else {
		raw, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("something went wrong when encoding the json")
		}
		return raw, nil
	}

}

func findStr(data []byte, startIndex int) string {
	startIndex++
	currentIndex := startIndex
	for len(data) > currentIndex {
		if data[currentIndex] == 0x00 {
			return string(data[startIndex:currentIndex])
		}
		currentIndex++
	}

	return ""
}

//Look at ParseWorldBlock, ParseSectorBlock, and ParseNodeList
