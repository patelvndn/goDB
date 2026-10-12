package encoder

import (
	"encoding/binary"

	"github.com/patelvndn/goDB/internal/btree/btreeNode"
)

type Node = btreeNode.Node
var PAGE_ID_BYTES int = 2
var KEY_BYTES int = 4
var PAGE_SIZE int = 4000 // 4 kb page size
/*
	we have a specific schema to follow
		n.children will be pageIds with uint16
		n.keys will be uint32

	everything is encoded in little endian 
*/
func Encode(n *Node, t int) ([]byte) {
	payload := make([]byte, PAGE_SIZE)

	encodeLeafFlag(payload, n.IsLeaf)
	encodeChildren(payload, n.Children, 2)
	encodeKeys(payload, n.Keys, 2 + PAGE_ID_BYTES*(2*t+1))

	return payload
}

func encodeKeys(bytes []byte, keys []uint32, start int) {
		for _, key := range keys {
		binary.LittleEndian.PutUint32(bytes[start:start+binary.Size(key)], key)
		start += binary.Size(key)
	}
}

// if it is a leaf set flag positive
func encodeLeafFlag(bytes []byte, isLeaf bool) {
	var flag uint8 = 1
	if !isLeaf {
		flag = 0
	}
	bytes[0] = byte(flag)
}

func encodeChildren(bytes []byte, pageIds []uint16, start int){
	// start at i and go to childrenLengthBytes
	for _, pageId := range pageIds {
		binary.LittleEndian.PutUint16(bytes[start:start+binary.Size(pageId)], pageId)
		start += binary.Size(pageId)
	}
}

func Decode(payload []byte, t int) (*Node, error) {
	node := btreeNode.New()

	keyBytesStart := 2 + PAGE_ID_BYTES*(2*t+1)
	
	node.IsLeaf = decodeLeafFlag(payload)
	node.Children = decodeChildren(payload[2:keyBytesStart])
	node.Keys = decodeKeys(payload[keyBytesStart:])

	return node, nil 
}

// returns True if we are a leaf, a leaf has a nonzero flag
func decodeLeafFlag(payload []byte) bool {
	flag := uint8(payload[0])
	return flag != 0 

}

func decodeChildren(payload []byte) []uint16 {
	// keep reading bytes until we reach the limit or we read a 0 
	// we stop at 0 since no child an ever have pageID 0 
	children := make([]uint16, 0)
	for i := 0; i < len(payload); i += PAGE_ID_BYTES {
		children = append(children, binary.LittleEndian.Uint16(payload[i:i+PAGE_ID_BYTES]))
	}
	return children
}

func decodeKeys(payload []byte) []uint32 {
	keys := make([]uint32, 0)
	for i := 0; i < len(payload); i += KEY_BYTES {
		keys = append(keys, binary.LittleEndian.Uint32(payload[i:i+KEY_BYTES]))
	}
	return keys
}