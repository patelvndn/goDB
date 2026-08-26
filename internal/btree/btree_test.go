package btree

import "testing"

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
	if !bt.root.isLeaf {
		t.Error("a brand new tree's root should be a leaf")
	}
	if len(bt.root.keys) != 0 {
		t.Errorf("a brand new tree's root should have no keys, got %v", bt.root.keys)
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
	keys := []int{10, 20, 5, 15, 25, 1, 30}
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

	_, ok := bt.InsertNode(10)
	if ok {
		t.Error("inserting a key that already exists should report ok=false")
	}
}

// --- Internal invariant: a leaf's own keys slice should reflect what was inserted ---
// This isolates *where* insert is going wrong, independent of whether search
// happens to work.

func TestInsertPopulatesLeafKeys(t *testing.T) {
	bt := New(2)
	bt.InsertNode(10)

	if len(bt.root.keys) != 1 {
		t.Fatalf("expected root.keys to have 1 entry after one insert, got %v", bt.root.keys)
	}
	if bt.root.keys[0] != 10 {
		t.Errorf("expected root.keys[0] == 10, got %d", bt.root.keys[0])
	}
}

func TestInsertKeepsKeysSorted(t *testing.T) {
	bt := New(2)
	for _, k := range []int{30, 10, 20} {
		bt.InsertNode(k)
	}

	got := bt.root.keys
	for i := 1; i < len(got); i++ {
		if got[i-1] > got[i] {
			t.Errorf("keys not sorted: %v", got)
			break
		}
	}
}