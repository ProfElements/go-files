package fetm


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
	Data []byte 
}

type header struct {
	Magic variantData
	Note  variantData
	SectionCount variantData 
	FileSize variantData 
}

type world struct {}

type sector struct {}

type node struct {}

type FETM struct {
	Header header
	World  world
	Sector sector
	Nodes  []node
	Raw    []variantData
}

