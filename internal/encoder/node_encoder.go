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
*/

func Encode(n *Node, t int) ([]byte, error) {
	childrenLength := PAGE_ID_BYTES * (2*t-1)
	keyLength := KEY_BYTES * len(n.Keys)

	payload := make([]byte, 1 + childrenLength + keyLength)

	if n.IsLeaf {
		payload[0] = 1
	}

	for _, c := range n.Children {
		binary.LittleEndian.AppendUint16(payload, c)
	}
	
	// we reserve 4 bytes for rle which is generous tbh
	var keySize uint32  = uint32(len(n.Keys))

	binary




}

func Decode(bytes []byte) (*Node, error) {

}