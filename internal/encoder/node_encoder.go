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
	
	childrenLengthBytes := PAGE_ID_BYTES * (2*t-1)
	payload := make([]byte, 1 + childrenLengthBytes)
	
	if n.IsLeaf {
		payload[0] = 1
	}
	
	for _, c := range n.Children {
		binary.LittleEndian.AppendUint16(payload, c)
	}
	
	keyLength := KEY_BYTES * len(n.Keys)
	// we reserve 4 bytes for rle which is 
	var keySize uint32  = uint32(len(n.Keys))

	for _, k := range n.Keys {
		binary.LittleEndian.AppendUint32(k)
	}
}

func Decode(bytes []byte) (*Node, error) {

}