package btreeNode

import (
	"encoding/binary"
	"reflect"
	"unsafe"
)

// this is the same for each node, won't change even with generalized schemas
type header struct {
	isLeaf    uint8
	freeKey   uint16
	freeTuple uint16
	children  []uint16 // not expensive to copy so might as well
}

// contains the key ID and the pointer of where to go to
type slot struct {
	start uint16
	size  uint16
	key   uint32
}

const (
	PAGE_SIZE uint16 = 4096
	BYTES_2   int    = 2
	KEY_BYTES int    = 4
)

/*
	we have a specific schema to follow

		n.children will be pageIds with uint16
		n.keys will be uint32

	everything is encoded in little endian

	we know the schema since the schema lives here
	pages get decoded into headers first
*/

func maxNumChildren(t int) int {
	return 2*t + 1
}

func (h *header) hspace(t int) uint16 {
	sizeOfChild := int(reflect.TypeOf(h.children).Elem().Size())
	return h.mspace() + uint16(sizeOfChild*maxNumChildren(t))
}

func (h *header) mspace() uint16 {
	return uint16(binary.Size(h.isLeaf) + binary.Size(h.freeKey) + binary.Size(h.freeTuple))
}

func encode(n *Node, t int) []byte {

	// we have to recalculate the header... each time
	// we know where free key and freetuple will live

	// free key starts at  1 + 2 + 2 + 2 * (2*t+1) -- can have at most 2t+1 children and each child is 2 bytes
	// 5 + 4t + 2
	// 7 + 4t

	payload := make([]byte, PAGE_SIZE)
	header := encodeHeader(payload, n, t)
	encodeKeysAndValues(payload, n, header, t) // requires encoding the slots
	// need to rewrite certain values in the header here
	// so it seems like we should encode the keys and values first and then encode the header after
	// when we know where certain freeKey and freeValue are, otherwise we encoded in stale values.
	encodeFreePointers(payload, header)
	return payload
}

func encodeFreePointers(payload []byte, h *header) {
	binary.LittleEndian.PutUint16(payload[1:1+binary.Size(h.freeKey)], h.freeKey)
	binary.LittleEndian.PutUint16(payload[1+binary.Size(h.freeKey):1+binary.Size(h.freeKey)+binary.Size(h.freeTuple)], h.freeTuple)
}

func encodeHeader(payload []byte, n *Node, t int) *header {
	var h *header = &header{}
	var isLeaf uint8 = 0
	if n.IsLeaf {
		isLeaf += 1
	}
	payload[0] = isLeaf

	// free key is headerSize + slotSize * length of n.keys
	// what is header size? we are given t
	// remember total number of possible children are... if we split on size > 2t -1 so if we split on 2t keys then at most we will have 2t -1  nodes and therefroe 2t-1 children
	freeKey := h.hspace(t) //
	binary.LittleEndian.PutUint16(payload[1:1+binary.Size(freeKey)], freeKey)

	freeTuple := PAGE_SIZE
	binary.LittleEndian.PutUint16(payload[1+binary.Size(freeKey):1+binary.Size(freeKey)+binary.Size(freeTuple)], freeTuple)

	start := binary.Size(isLeaf) + binary.Size(freeKey) + binary.Size(freeTuple)
	stop := 2*t + 1
	// so children will be from 0 to 0 + 2t+1
	encodeChildren(payload, n.Children, start, stop)

	h.isLeaf = isLeaf
	h.freeKey = freeKey
	h.freeTuple = freeTuple
	h.children = n.Children
	return h
}

func encodeChildren(payload []byte, children []uint16, start int, stop int) {
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

func encodeKeysAndValues(payload []byte, n *Node, h *header, t int) {
	// for each k-v pair we need to calculate and write
	// the start
	// the size of bytes of data (value) written
	// key

	// this gives us a slot which we write to freeKey in header
	// then we write the data starting from freeTuple. so freeTuple - size(value) : freeTuple
	keys := n.Keys
	values := n.Values

	ft := h.freeTuple
	kstart := h.hspace(t)
	for i := range n.Keys {
		k := keys[i]
		v := values[i]

		s := slot{
			start: ft,
			size:  uint16(len(v)),
			key:   k,
		}

		// now write slot until fk
		// hspace -- fk is reserved so keys, fk
		binary.LittleEndian.PutUint16(payload[kstart:kstart+2], s.start)
		binary.LittleEndian.PutUint16(payload[kstart+2:kstart+4], s.size)
		binary.LittleEndian.PutUint32(payload[kstart+4:kstart+8], s.key)
		kstart += uint16(unsafe.Sizeof(s))

		// now we have to write pageSize ->
		vPayload := []byte(v)

		copy(payload[s.start-s.size:s.start], vPayload)
		ft = s.start - s.size
	}
	h.freeTuple = ft
	h.freeKey = kstart
}

// no values yet
func decode(payload []byte, t int) *Node {

	h := decodeHeader(payload, t)

	node := &Node{IsLeaf: true}
	if h.isLeaf == 0 {
		node.IsLeaf = false
	}

	node.Children = h.children

	keySlotsStart := 7 + 4*t
	slots := decodeKeys(payload[keySlotsStart:h.freeKey])

	keys := make([]uint32, len(slots))
	for i, s := range slots {
		keys[i] = s.key
	}

	node.Keys = keys
	node.Values = decodeValues(payload, slots)

	return node
}

func decodeValues(payload []byte, slots []slot) []string {
	values := make([]string, len(slots))
	for i, s := range slots {
		values[i] = string(payload[s.start-s.size : s.start])

	}
	return values
}

func decodeHeader(payload []byte, t int) *header {
	h := &header{}
	h.isLeaf = uint8(payload[0])
	h.freeKey = binary.LittleEndian.Uint16(payload[1:3])   // pointer offset we right to for keys
	h.freeTuple = binary.LittleEndian.Uint16(payload[3:5]) // pointer offset we right to for tuples

	children := make([]uint16, 0)
	for i := 5; i < 5+BYTES_2*(2*t+1); i += BYTES_2 {
		pageId := binary.LittleEndian.Uint16(payload[i : i+BYTES_2])
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

	if len(payload)%8 != 0 {
		panic("payload corrupted, should be multiple of 8")
	}

	keys := make([]slot, 0)

	for i := 0; i < len(payload); i += int(size) {
		keys = append(keys, slot{
			start: binary.LittleEndian.Uint16(payload[i : i+2]),
			size:  binary.LittleEndian.Uint16(payload[i+2 : i+4]),
			key:   binary.LittleEndian.Uint32(payload[i+4 : i+8]),
		})
	}

	return keys
}
