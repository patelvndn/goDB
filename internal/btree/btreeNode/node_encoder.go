package btreeNode

import (
	"encoding/binary"
	"unsafe"
)

// this is the same for each node, won't change even with generalized schemas
type header struct {
	isLeaf uint8;
	freeKey uint16;
	freeTuple uint16;
	children []uint16; // not expensive to copy so might as well 
}

// contains the key ID and the pointer of where to go to 
type slot struct {
	start uint16 
	size uint16
	key uint32
}

const (
	PAGE_SIZE uint16 = 4096
	BYTES_2 int = 2
	KEY_BYTES int = 4
)

/*
	we have a specific schema to follow
		n.children will be pageIds with uint16
		n.keys will be uint32

	everything is encoded in little endian 

	we know the schema since the schema lives here
	pages get decoded into headers first
*/
func encode(n *Node, t int) ([]byte) {

	// we have to recalculate the header... each time	
	// we know where free key and freetuple will live 

	// free key starts at  1 + 2 + 2 + 2 * (2*t+1) -- can have at most 2t+1 children and each child is 2 bytes
	// 5 + 4t + 2
	// 7 + 4t 


	payload := make([]byte, PAGE_SIZE)

	header := encodeHeader(payload, n, t)
	encodeKeysAndValues(payload, n, header, t) // requires encoding the slots // we en

	return payload
}

func encodeHeader(payload []byte, n *Node, t int) *header {
	var childSize uint16
	hspace := 5 + uint16(binary.Size(childSize)*(2*t+1))

	var isLeaf uint8 = 0 
	if n.IsLeaf {
		isLeaf += 1
	}
	payload[0] = isLeaf

	// free key is headerSize + slotSize * length of n.keys
	// what is header size? we are given t 
	// remember total number of possible children are... if we split on size > 2t -1 so if we split on 2t keys then at most we will have 2t -1  nodes and therefroe 2t-1 children 
	freeKey := hspace + uint16(unsafe.Sizeof(slot{})) * uint16(len(n.Keys))
	binary.LittleEndian.PutUint16(payload[1:1+binary.Size(freeKey)], freeKey)

	freeTuple := PAGE_SIZE
	binary.LittleEndian.PutUint16(payload[1+binary.Size(freeKey):1+2*binary.Size(freeKey)], freeTuple)

	// so children will be from 0 to 0 + 2t+1 
	encodeChildren(payload, n.Children, 2*t+1)
	return &header{
		isLeaf: isLeaf,
		freeKey: freeKey,
		freeTuple: freeTuple,
		children: n.Children,
	}
}

func encodeChildren(payload []byte, children []uint16, stop int) {
	start := 5
	for i := range stop {
		var c uint16 = 0
		if i < len(children) {
			c = children[i]
		}

		binary.LittleEndian.PutUint16(payload[start:start+binary.Size(c)], c)
		start += binary.Size(c)

		if c == 0 {
			// early exit if no more children
			break
		}
	}
}

func encodeKeysAndValues(payload []byte, n *Node, h *header, t int){
	// for each k-v pair we need to calculate and write
	// the start 
	// the size of bytes of data (value) written
	// key

	// this gives us a slot which we write to freeKey in header
	// then we write the data starting from freeTuple. so freeTuple - size(value) : freeTuple
	keys := n.Keys
	values := n.Values

	ft := h.freeTuple

	var childSize uint16
	hspace := 5 + uint16(binary.Size(childSize)*(2*t+1))
	for i := range(n.Keys){
		k := keys[i]
		v := values[i]

		s := slot{
			start: ft,
			size: uint16(binary.Size(v)),
			key: k,
		}

		// now write slot until fk 
		// hspace -- fk is reserved so keys, fk 
		binary.LittleEndian.PutUint16(payload[hspace:hspace+2], s.start)
		binary.LittleEndian.PutUint16(payload[hspace+2:hspace+4], s.size)
		binary.LittleEndian.PutUint32(payload[hspace+4:hspace+8], s.key)
		hspace += uint16(unsafe.Sizeof(s))

		// now we have to write pageSize -> 
		copy(payload[ft-uint16(len(v)):ft],[]byte(v))
	}

	h.freeKey = hspace

}

// no values yet
func decode(payload []byte, t int) *Node {
	h := decodeHeader(payload,t)

	node := &Node{IsLeaf: true}
	if h.isLeaf == 0 {
		node.IsLeaf=false
	}

	node.Children = h.children

	keySlotsStart := 7 + 4*t
	slots := decodeKeys(payload[keySlotsStart:h.freeKey])

	keys := make([]uint32, len(slots))
	for i, s := range(slots){
		keys[i] = s.key
	}
	node.Keys = keys
	// node.Values = decodeValues(payload[freeTuple:]) // something like this 
	return node
}


func decodeHeader(payload []byte, t int) *header{

	h := &header{}
	h.isLeaf = uint8(payload[0])
	h.freeKey = binary.LittleEndian.Uint16(payload[1:3]) // pointer offset we right to for keys
	h.freeTuple = binary.LittleEndian.Uint16(payload[3:5]) // pointer offset we right to for tuples

	children := make([]uint16, 0)
	for i := 5; i < 5 + BYTES_2*(2*t+1); i += BYTES_2 {
		pageId := binary.LittleEndian.Uint16(payload[i:i+BYTES_2])
		if pageId != 0 {
			children = append(children, pageId)
		} else {
			break
		}
	}
	h.children = children

	return h
}

func decodeKeys(payload []byte) []slot {
	size := unsafe.Sizeof(slot{})

	if len(payload) % 8 != 0 {
		panic("payload corrupted, should be multiple of 8")
	}

	keys := make([]slot, 0)

	for i := 0; i < len(payload); i += int(size) {
		keys = append(keys, slot{
			start: binary.LittleEndian.Uint16(payload[i:i+2]),
			size: binary.LittleEndian.Uint16(payload[i+2:i+4]),
			key: binary.LittleEndian.Uint32(payload[i+4:i+8]),
		})
	}

	return keys
}