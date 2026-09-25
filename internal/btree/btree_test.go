package btree

import (
	"math/rand"
	"testing"

	"github.com/patelvndn/goDB/internal/btree/btreeNode"
)

type Node = btreeNode.Node

// ---- invariant checker ----
// Walks the tree after every operation and fails loudly at the first
// broken invariant, instead of surfacing as a mysterious "not found"
// three operations later.

func validateTree(t *testing.T, bt *Btree) {
	t.Helper()
	if bt.root == nil {
		t.Fatalf("root is nil")
	}
	depth := leafDepth(bt.root, 0)
	validateNode(t, bt.root, bt.t, true, depth, 0)
}

func leafDepth(n *Node, d int) int {
	if n.IsLeaf {
		return d
	}
	return leafDepth(n.Children[0], d+1)
}


func findIndex (keys []int, key int) int {
	for i, k := range keys {
		if key <= k {
			return i
		}
	}
	return len(keys)
}


func validateNode(t *testing.T, n *Node, order int, isRoot bool, wantLeafDepth, depth int) {
	t.Helper()

	max := 2*order - 1
	if len(n.Keys) > max {
		t.Fatalf("node has %d keys, exceeds max %d: %v", len(n.Keys), max, n.Keys)
	}
	if !isRoot && len(n.Keys) < order-1 {
		t.Fatalf("non-root node underflowed: %d keys, min %d: %v", len(n.Keys), order-1, n.Keys)
	}

	for i := 1; i < len(n.Keys); i++ {
		if n.Keys[i-1] >= n.Keys[i] {
			t.Fatalf("keys not strictly sorted: %v", n.Keys)
		}
	}

	if n.IsLeaf {
		if depth != wantLeafDepth {
			t.Fatalf("leaf at depth %d, expected %d (tree unbalanced)", depth, wantLeafDepth)
		}
		if len(n.Children) != 0 {
			t.Fatalf("leaf has %d children, expected 0", len(n.Children))
		}
		return
	}

	if len(n.Children) != len(n.Keys)+1 {
		t.Fatalf("node has %d keys but %d children (want %d): keys=%v",
			len(n.Keys), len(n.Children), len(n.Keys)+1, n.Keys)
	}

	for i, child := range n.Children {
		if len(child.Keys) == 0 {
			t.Fatalf("child %d has zero keys", i)
		}
		if i > 0 && child.Keys[0] <= n.Keys[i-1] {
			t.Fatalf("child %d first key %d not > separator %d", i, child.Keys[0], n.Keys[i-1])
		}
		if i < len(n.Keys) && child.Keys[len(child.Keys)-1] >= n.Keys[i] {
			t.Fatalf("child %d last key %d not < separator %d", i, child.Keys[len(child.Keys)-1], n.Keys[i])
		}
		validateNode(t, child, order, false, wantLeafDepth, depth+1)
	}
}

// --- findIndex: pure function, easiest to pin down first ---

func TestFindIndex(t *testing.T) {
	cases := []struct {
		name string
		keys []int
		key  int
		want int
	}{
		{"empty slice", []int{}, 5, 0},
		{"key smaller than all", []int{10, 20, 30}, 5, 0},
		{"key larger than all", []int{10, 20, 30}, 35, 3},
		{"key equals first", []int{10, 20, 30}, 10, 0},
		{"key equals middle", []int{10, 20, 30}, 20, 1},
		{"key equals last", []int{10, 20, 30}, 30, 2},
		{"key falls between elements", []int{10, 20, 30}, 15, 1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := findIndex(c.keys, c.key)
			if got != c.want {
				t.Errorf("findIndex(%v, %d) = %d, want %d", c.keys, c.key, got, c.want)
			}
		})
	}
}

// --- New / tree construction ---

func TestNew(t *testing.T) {
	bt := New(2)
	if bt == nil {
		t.Fatal("New returned nil")
	}
	if bt.root == nil {
		t.Fatal("root is nil")
	}
	if !bt.root.IsLeaf {
		t.Error("a brand new tree's root should be a leaf")
	}
	if len(bt.root.Keys) != 0 {
		t.Errorf("a brand new tree's root should have no keys, got %v", bt.root.Keys)
	}
}

// --- SearchNode on an empty tree ---

func TestSearchEmptyTree(t *testing.T) {
	bt := New(2)
	_, _, found := bt.SearchNode(42)
	if found {
		t.Error("searching an empty tree should never find a key")
	}
}

// --- Insert then search: the core round-trip every B-tree needs ---

func TestInsertThenSearchSingleKey(t *testing.T) {
	bt := New(2)
	bt.InsertNode(10)
	_, _, found := bt.SearchNode(10)
	if !found {
		t.Error("expected to find key 10 immediately after inserting it")
	}
}

func TestInsertMultipleThenSearchEach(t *testing.T) {
	bt := New(2)
	keys := []int{10, 20, 5, 15, 25, 1, 30, 3, 2, 4, 44, 67, 87, 546, 12}
	for _, k := range keys {
		bt.InsertNode(k)
	}
	for _, k := range keys {
		_, _, found := bt.SearchNode(k)
		if !found {
			t.Errorf("expected to find key %d after inserting all of %v", k, keys)
		}
	}
}

func TestSearchKeyNeverInserted(t *testing.T) {
	bt := New(2)
	bt.InsertNode(10)
	bt.InsertNode(20)

	_, _, found := bt.SearchNode(99)
	if found {
		t.Error("search for a key that was never inserted should return false")
	}
}

// --- Duplicate handling ---

func TestInsertDuplicateKeyIsRejected(t *testing.T) {
	bt := New(2)
	bt.InsertNode(10)

	err := bt.InsertNode(10)
	if err == nil {
		t.Error("inserting a key that already exists should report ok=false")
	}
}

// --- Internal invariant: a leaf's own keys slice should reflect what was inserted ---
// This isolates *where* insert is going wrong, independent of whether search
// happens to work.

func TestInsertPopulatesLeafKeys(t *testing.T) {
	bt := New(2)
	bt.InsertNode(10)

	if len(bt.root.Keys) != 1 {
		t.Fatalf("expected root.keys to have 1 entry after one insert, got %v", bt.root.Keys)
	}
	if bt.root.Keys[0] != 10 {
		t.Errorf("expected root.keys[0] == 10, got %d", bt.root.Keys[0])
	}
}

func TestInsertKeepsKeysSorted(t *testing.T) {
	bt := New(2)
	for _, k := range []int{30, 10, 20} {
		bt.InsertNode(k)
	}

	got := bt.root.Keys
	for i := 1; i < len(got); i++ {
		if got[i-1] > got[i] {
			t.Errorf("keys not sorted: %v", got)
			break
		}
	}
}

// ---- basic behavior ----

func TestRemove_EmptyTree(t *testing.T) {
	bt := New(2)
	if err := bt.RemoveNode(5); err == nil {
		t.Errorf("expected error removing from empty tree, got nil")
	}
}

func TestRemove_NonExistentKey(t *testing.T) {
	bt := New(2)
	for _, k := range []int{10, 20, 30} {
		bt.InsertNode(k)
	}
	if err := bt.RemoveNode(99); err == nil {
		t.Errorf("expected error removing key not present, got nil")
	}
}

func TestRemove_SingleKey(t *testing.T) {
	bt := New(2)
	bt.InsertNode(10)

	if err := bt.RemoveNode(10); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, _, found := bt.SearchNode(10); found {
		t.Errorf("expected 10 to be removed")
	}
}

func TestRemove_LeafNoUnderflow(t *testing.T) {
	// t=2: a leaf can hold up to 3 keys, so removing one from a
	// 3-key leaf shouldn't need to borrow or merge.
	bt := New(2)
	for _, k := range []int{10, 20, 30} {
		bt.InsertNode(k)
	}

	if err := bt.RemoveNode(20); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, _, found := bt.SearchNode(20); found {
		t.Errorf("expected 20 to be removed")
	}
	for _, k := range []int{10, 30} {
		if _, _, found := bt.SearchNode(k); !found {
			t.Errorf("expected %d to still be found", k)
		}
	}
	validateTree(t, bt)
}

func TestRemove_RootCollapsesToChild(t *testing.T) {
	// Force at least one split, then remove enough keys that the
	// root should shrink back down to a single leaf.
	bt := New(2)
	for _, k := range []int{10, 20, 30, 40, 50} {
		bt.InsertNode(k)
	}
	if bt.root.IsLeaf {
		t.Fatalf("setup invalid: expected root to have split already")
	}

	for _, k := range []int{10, 20, 30, 40} {
		if err := bt.RemoveNode(k); err != nil {
			t.Fatalf("unexpected error removing %d: %v", k, err)
		}
		validateTree(t, bt)
	}

	if !bt.root.IsLeaf {
		t.Errorf("expected root to have collapsed to a leaf, still internal with keys %v", bt.root.Keys)
	}
	if _, _, found := bt.SearchNode(50); !found {
		t.Errorf("expected 50 to survive")
	}
}

// ---- property-style tests: these are what actually exercise
// borrow-left, borrow-right, and merge, without hand-picking cases ----

func TestRemove_AllKeysAscending(t *testing.T) {
	bt := New(2)
	keys := []int{10, 20, 5, 15, 25, 1, 30, 3, 2, 4, 44, 67, 87, 546, 12}
	for _, k := range keys {
		bt.InsertNode(k)
	}

	sorted := append([]int(nil), keys...)
	sortInts(sorted)

	for idx, k := range sorted {
		if err := bt.RemoveNode(k); err != nil {
			t.Fatalf("unexpected error removing %d: %v", k, err)
		}
		validateTree(t, bt)
		if _, _, found := bt.SearchNode(k); found {
			t.Fatalf("key %d still found after removal", k)
		}
		for _, remaining := range sorted[idx+1:] {
			if _, _, found := bt.SearchNode(remaining); !found {
				t.Fatalf("key %d missing after removing %d", remaining, k)
			}
		}
	}
}

func TestRemove_AllKeysDescending(t *testing.T) {
	bt := New(2)
	keys := []int{10, 20, 5, 15, 25, 1, 30, 3, 2, 4, 44, 67, 87, 546, 12}
	for _, k := range keys {
		bt.InsertNode(k)
	}

	sorted := append([]int(nil), keys...)
	sortInts(sorted)
	reverse(sorted)

	for idx, k := range sorted {
		if err := bt.RemoveNode(k); err != nil {
			t.Fatalf("unexpected error removing %d: %v", k, err)
		}
		validateTree(t, bt)
		for _, remaining := range sorted[idx+1:] {
			if _, _, found := bt.SearchNode(remaining); !found {
				t.Fatalf("key %d missing after removing %d", remaining, k)
			}
		}
	}
}

func TestRemove_RandomOrder(t *testing.T) {
	r := rand.New(rand.NewSource(42)) // fixed seed: reproducible on failure

	keys := make([]int, 200)
	for i := range keys {
		keys[i] = i
	}
	r.Shuffle(len(keys), func(i, j int) { keys[i], keys[j] = keys[j], keys[i] })

	bt := New(2)
	for _, k := range keys {
		bt.InsertNode(k)
		validateTree(t, bt)
	}

	removeOrder := append([]int(nil), keys...)
	r.Shuffle(len(removeOrder), func(i, j int) { removeOrder[i], removeOrder[j] = removeOrder[j], removeOrder[i] })

	present := make(map[int]bool)
	for _, k := range keys {
		present[k] = true
	}

	for _, k := range removeOrder {
		if err := bt.RemoveNode(k); err != nil {
			t.Fatalf("unexpected error removing %d: %v", k, err)
		}
		delete(present, k)
		validateTree(t, bt)

		if _, _, found := bt.SearchNode(k); found {
			t.Fatalf("key %d still found after removal", k)
		}
		// spot-check a handful of survivors rather than all 200 every iteration
		checked := 0
		for survivor := range present {
			if _, _, found := bt.SearchNode(survivor); !found {
				t.Fatalf("key %d missing after removing %d", survivor, k)
			}
			checked++
			if checked >= 10 {
				break
			}
		}
	}
}

// ---- small local helpers (or swap in "sort" / "slices" from stdlib) ----

func sortInts(s []int) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j-1] > s[j]; j-- {
			s[j-1], s[j] = s[j], s[j-1]
		}
	}
}

func reverse(s []int) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

