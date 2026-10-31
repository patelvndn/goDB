package btreeNode

import (
	"errors"
	"slices"

	"github.com/patelvndn/goDB/internal/pager"
)

type pgr = pager.Pager
type page = pager.Page

type Node struct {
	IsLeaf bool
	PageId uint16
	Children []uint16
	Keys []uint32 // these might not always be integers... but for now they will be
	Values []string // will be string for now... eventually they'll just be bytes and then be decoded further up on a schema level
}

func New(pager *pgr) *Node {
	n :=  &Node{IsLeaf: true}
	
	p, err := pager.AllocatePage()
	if err != nil {
		return nil
	}
	n.PageId = p.Id

	return n
}

func findIndex (keys []uint32, key uint32) int {
	for i, k := range keys {
		if key <= k {
			return i
		}
	}
	return len(keys)
}

func (n *Node) Insert(key uint32, t int, pager *pgr) (*Node, error) {
	promoted, right, split, err := n.insert(key, t, pager);

	if err != nil {
		return nil, err
	}

	var newRoot *Node
	var left uint16

	if split {
		left = n.PageId
		newRoot = New(pager)
		newRoot.Keys = []uint32{promoted}
		newRoot.Children = []uint16{left, right.PageId}
		n = newRoot
	}
	return n, nil 
}


func (n *Node) insert(key uint32, t int, pager *pgr) (uint32, *Node, bool, error) {
	i := findIndex(n.Keys, key) 
	// this is the node with the key
	if 0 <= i && i < len(n.Keys) && n.Keys[i] == key {
		return 0, nil, false, errors.New("Key already exists in node")
	}

	if n.IsLeaf {
		// insert key into index i 
		n.Keys = slices.Insert(n.Keys, i, key)
		n.Values = slices.Insert(n.Values, i, "nityaISchud")

		// write page to persist data
		_, err := pager.WritePage(n.PageId, encode(n, t))
		if err != nil {
			return 0, nil, false, err
		}

	} else {
		// propogate the split back up, if it happened
		var child *Node
		var err error
		
		if child, err = DecodeChild(n.Children[i], t, pager); err != nil {
			return 0, nil, false, err
		}

		promoted, right, split, err := child.insert(key,t, pager);
		if err != nil {
			return 0, nil, false, err
		}

		if split {
			// n is not a leaf therefore it must have children and therefore it definitely has at least len(n.keys)+1 children 
			n.Keys = slices.Insert(n.Keys, i, promoted)
			n.Children = slices.Insert(n.Children, i+1, right.PageId)
		}

	}

	// split when we have 2t keys
	if len(n.Keys) > 2*t-1 {
        return n.split(t, pager)
    } 
	return 0, nil, false, nil
}

func (n *Node)split(t int, pager *pgr) (uint32, *Node, bool, error)  {
	// splitting a node means half the keys get shifted to a new node (right node)
	// and the middle key (t) gets promoted up so 0 ... t-1 = left has t keys and right will have t-1 keys
	// n turns into left
	promoted := n.Keys[t]
	
	// 0,1, ... t-1
	leftKeys := make([]uint32, t) // make list with size t
	copy(leftKeys, n.Keys[:t]) 
	
	// t+1, ... 2t
	rightKeys := make([]uint32, t-1) 
	copy(rightKeys, n.Keys[t+1:]) // dont copy promoted element
	
	n.Keys = leftKeys
	
	// need to split children too, but how?
	// node could be a leaf still aka it won't have any children
	right := New(pager)
	if right == nil {
		return 0, nil, false, errors.New("Error creating node")
	}

	right.Keys = rightKeys
	
	if !n.IsLeaf {
		rightChildren := make([]uint16, len(n.Children)-(t+1))
		copy(rightChildren, n.Children[t+1:],)
		right.Children = rightChildren
		
		leftChildren := make([]uint16, t+1)
		copy(leftChildren, n.Children[:t+1])
		n.Children = leftChildren
		
		right.IsLeaf = false
	}
	pager.WritePage(right.PageId, encode(right, t))

	return promoted, right, true, nil
}

// search within the node itself to find out where the key exists
func (n *Node) Search(key uint32, t int, pager *pgr) (*Node, int, bool) {
	i := findIndex(n.Keys, key)		
	// this is the node with the key
	if i < len(n.Keys) && n.Keys[i] == key {
		return n, i, true
	}

	// this node is a leaf, no more children to search
	if n.IsLeaf {
		return nil, 0, false
	}

	// recursively search the children that would have the key
	child, err := DecodeChild(n.Children[i],t, pager)
	if err!= nil {
		return nil, 0, false
	}
	return child.Search(key, t, pager)
}

func (n *Node) Remove(key uint32, t int, pager *pgr) (*Node, error) {
	
	if err := n.remove(key, t, pager); err != nil {
		return n, err
	}

	if len(n.Keys) == 0 && !n.IsLeaf {
		child, err := DecodeChild(n.Children[0], t, pager)
		if err != nil {
			return nil, err
		}
		return child, nil
	}

	return n, nil
}

func (n *Node) remove(key uint32, t int, pager *pgr) (error) {

	i := findIndex(n.Keys, key)
	found := false
	if i < len(n.Keys) && n.Keys[i] == key {
		found = true
	}

	if n.IsLeaf {
		// key is not in internal node
		if found {
			n.Keys = slices.Delete(n.Keys,i,i+1)
			return nil
		} 
		return errors.New("Key not found error")
	}

	left, err := DecodeChild(n.Children[i],t, pager)
	if err != nil {
		return err
	}

	right, err := DecodeChild(n.Children[i+1], t, pager)
	if err != nil {
		return err
	}

	if found {
		if len(left.Keys) >= t {
			pred := left.maxKey(t, pager)
			n.Keys[i] = pred // this is the 'delete' -> rewrite i and then remove pred from left
			return left.remove(pred, t, pager)
		} else if len(right.Keys) >= t {
			succ := right.minKey(t, pager)
			n.Keys[i] = succ
			return right.remove(succ, t, pager)
		}
		merge(n,i, t, pager) // so when both left and right are < t so merging gets us  < 2t, then in the merge we have to add i to it, then we remove it from the chil
		return left.remove(key, t, pager)
	} 

	if len(left.Keys) < t {
		i = n.fill(i, t, pager) // we need to balance our tree by adding keys
	}

	return left.remove(key, t, pager)
}

// ensures Children[i] has >= t keys; returns the (possibly shifted) child index
func (n *Node) fill(i int, t int, pager *pgr) int {

	child, _ := DecodeChild(n.Children[i],t, pager)
	left_sib, _ := DecodeChild(n.Children[i-1], t, pager)
	right_sib, _ := DecodeChild(n.Children[i+1], t, pager)

	switch {
		case i > 0 && len(left_sib.Keys) >= t:
			
			sib := left_sib
			child.Keys = slices.Insert(child.Keys, 0, n.Keys[i-1])

			if !child.IsLeaf {
				last := len(sib.Children) - 1
				child.Children = slices.Insert(child.Children, 0, sib.Children[last])
				sib.Children = sib.Children[:last]
			}

			n.Keys[i-1] = sib.Keys[len(sib.Keys)-1]
			sib.Keys = sib.Keys[:len(sib.Keys)-1]
		case i < len(n.Keys) && len(right_sib.Keys) >= t:
			sib := right_sib
			child.Keys = append(child.Keys, n.Keys[i])

			if !child.IsLeaf {
				child.Children = append(child.Children, sib.Children[0])
				sib.Children = sib.Children[1:]
			}

			n.Keys[i] = sib.Keys[0]
			sib.Keys = sib.Keys[1:]
		case i < len(n.Keys):
			merge(n, i, t, pager)
		default:
			merge(n, i-1, t, pager)
			return i - 1
	}
	return i
}


func (n *Node) maxKey(t int, pager *pgr) uint32 {
	for !n.IsLeaf{
		n, _ = DecodeChild(n.Children[len(n.Children)-1],t, pager)
	}
	return n.Keys[len(n.Keys) - 1]
}

func (n *Node) minKey(t int, pager *pgr) uint32 {
	for !n.IsLeaf{
		n, _ = DecodeChild(n.Children[0],t, pager)
	}
	return n.Keys[0]
}

// a handle merge function is probably needed? maybe this can live in btree? 
func merge (n *Node, i int, t int, pager *pgr) {
	// we want to merge left and right together and remove the node at index i 
	left, _ := DecodeChild(n.Children[i],t, pager)
	right, _ := DecodeChild(n.Children[i+1], t, pager)

	left.Keys = append(left.Keys, n.Keys[i])
	left.Keys = append(left.Keys, right.Keys...)
	left.Children = append(left.Children, right.Children...)

	n.Children = slices.Delete(n.Children, i+1, i+2) // delete the right child
	n.Keys = slices.Delete(n.Keys,i,i+1) // delete the key at index i 

	
}

func DecodeChild(pageId uint16, t int, pager *pgr) (*Node, error){
	page, err := pager.ReadPage(pageId)
	if err != nil {
		return nil, err
	}
	return decode(page.Data,t), nil
}
