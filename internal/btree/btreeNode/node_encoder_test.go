package btreeNode

import (
	"encoding/binary"
	"reflect"
	"slices"
	"testing"
)

// buildPayload lays out a page header exactly as decodeHeader expects it:
// [0] isLeaf | [1:3] freeKey | [3:5] freeTuple | [5:...] 2t+1 child slots
func buildPayload(t int, isLeaf uint8, freeKey, freeTuple uint16, children []uint16) []byte {
	buf := make([]byte, 5+BYTES_2*(2*t+1))
	buf[0] = isLeaf
	binary.LittleEndian.PutUint16(buf[1:3], freeKey)
	binary.LittleEndian.PutUint16(buf[3:5], freeTuple)
	for i, c := range children {
		off := 5 + i*BYTES_2
		binary.LittleEndian.PutUint16(buf[off:off+BYTES_2], c)
	}
	return buf
}

func TestDecodeHeader(t *testing.T) {
	const order = 2 // 2t+1 = 5 child slots

	full := make([]uint16, 2*order+1)
	for i := range full {
		full[i] = uint16(i + 1)
	}

	tests := []struct {
		name         string
		isLeaf       uint8
		freeKey      uint16
		freeTuple    uint16
		children     []uint16
		wantChildren []uint16
	}{
		{"leaf with no children", 1, 100, 4000, nil, []uint16{}},
		{"internal node, partially filled", 0, 120, 3900, []uint16{7, 9, 12}, []uint16{7, 9, 12}},
		{"internal node, all slots filled", 0, 50, 4096, full, full},
		{"multi-byte values", 0, 0x0102, 0xABCD, []uint16{0x0100, 0xFFFF}, []uint16{0x0100, 0xFFFF}},
		{"zero page id stops decoding", 0, 10, 20, []uint16{10, 0, 30}, []uint16{10}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			payload := buildPayload(order, tc.isLeaf, tc.freeKey, tc.freeTuple, tc.children)

			got := decodeHeader(payload, order)

			if got.isLeaf != tc.isLeaf {
				t.Errorf("isLeaf = %d, want %d", got.isLeaf, tc.isLeaf)
			}
			if got.freeKey != tc.freeKey {
				t.Errorf("freeKey = %d, want %d", got.freeKey, tc.freeKey)
			}
			if got.freeTuple != tc.freeTuple {
				t.Errorf("freeTuple = %d, want %d", got.freeTuple, tc.freeTuple)
			}
			if !slices.Equal(got.children, tc.wantChildren) {
				t.Errorf("children = %v, want %v", got.children, tc.wantChildren)
			}
		})
	}
}

// Pins the on-disk byte order so an accidental switch to BigEndian is caught.
func TestDecodeHeaderLittleEndian(t *testing.T) {
	payload := make([]byte, 5+BYTES_2*3)
	payload[1], payload[2] = 0x02, 0x01 // freeKey -> 0x0102
	payload[3], payload[4] = 0x04, 0x03 // freeTuple -> 0x0304

	got := decodeHeader(payload, 1)

	if got.freeKey != 0x0102 {
		t.Errorf("freeKey = %#x, want 0x0102", got.freeKey)
	}
	if got.freeTuple != 0x0304 {
		t.Errorf("freeTuple = %#x, want 0x0304", got.freeTuple)
	}
}

// Documents current behavior: a payload shorter than the header panics.
// If you later add bounds checking and return an error, update this test.
func TestDecodeHeaderShortPayloadPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on short payload, got none")
		}
	}()
	decodeHeader(make([]byte, 3), 2)
}

// expectedFreeKey is the independent calculation of where the key slots begin:
// 1 (isLeaf) + 2 (freeKey) + 2 (freeTuple) + 2 bytes * (2t+1) child slots = 7 + 4t
func expectedFreeKey(t int) uint16 {
	return uint16(1 + 2 + 2 + BYTES_2*(2*t+1))
}

func TestHspace(t *testing.T) {
	for _, order := range []int{1, 2, 3, 10} {
		h := &header{}
		if got, want := h.hspace(order), expectedFreeKey(order); got != want {
			t.Errorf("hspace(%d) = %d, want %d", order, got, want)
		}
	}
}

func TestEncodeHeader(t *testing.T) {
	tests := []struct {
		name     string
		order    int
		isLeaf   bool
		children []uint16
	}{
		{"leaf, no children", 2, true, nil},
		{"leaf, empty non-nil children", 2, true, []uint16{}},
		{"internal, partially filled", 2, false, []uint16{7, 9, 12}},
		{"internal, all slots filled", 2, false, []uint16{1, 2, 3, 4, 5}},
		{"multi-byte page ids", 2, false, []uint16{0x0100, 0xFFFF}},
		{"order 1", 1, false, []uint16{4, 8, 15}},
		{"order 3", 3, false, []uint16{1, 2, 3, 4, 5, 6, 7}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			n := &Node{IsLeaf: tc.isLeaf, Children: tc.children}
			payload := make([]byte, PAGE_SIZE)

			h := encodeHeader(payload, n, tc.order)

			wantFreeKey := expectedFreeKey(tc.order)

			// --- returned header ---
			wantLeaf := uint8(0)
			if tc.isLeaf {
				wantLeaf = 1
			}
			if h.isLeaf != wantLeaf {
				t.Errorf("header.isLeaf = %d, want %d", h.isLeaf, wantLeaf)
			}
			if h.freeKey != wantFreeKey {
				t.Errorf("header.freeKey = %d, want %d", h.freeKey, wantFreeKey)
			}
			if h.freeTuple != PAGE_SIZE {
				t.Errorf("header.freeTuple = %d, want %d", h.freeTuple, PAGE_SIZE)
			}
			if !slices.Equal(h.children, tc.children) {
				t.Errorf("header.children = %v, want %v", h.children, tc.children)
			}

			// --- raw bytes written to the page ---
			if payload[0] != wantLeaf {
				t.Errorf("payload[0] = %d, want %d", payload[0], wantLeaf)
			}
			if got := binary.LittleEndian.Uint16(payload[1:3]); got != wantFreeKey {
				t.Errorf("encoded freeKey = %d, want %d", got, wantFreeKey)
			}
			if got := binary.LittleEndian.Uint16(payload[3:5]); got != PAGE_SIZE {
				t.Errorf("encoded freeTuple = %d, want %d", got, PAGE_SIZE)
			}
			for i, c := range tc.children {
				off := 5 + i*BYTES_2
				if got := binary.LittleEndian.Uint16(payload[off : off+BYTES_2]); got != c {
					t.Errorf("encoded child[%d] = %d, want %d", i, got, c)
				}
			}

			// everything from the end of the children region onward must be untouched
			for i := int(wantFreeKey); i < len(payload); i++ {
				if payload[i] != 0 {
					t.Fatalf("payload[%d] = %d, want 0 (wrote past header region)", i, payload[i])
				}
			}

			// --- round trip through decodeHeader ---
			got := decodeHeader(payload, tc.order)
			if got.isLeaf != wantLeaf || got.freeKey != wantFreeKey || got.freeTuple != PAGE_SIZE {
				t.Errorf("round trip scalars = (%d, %d, %d), want (%d, %d, %d)",
					got.isLeaf, got.freeKey, got.freeTuple, wantLeaf, wantFreeKey, PAGE_SIZE)
			}
			if !slices.Equal(got.children, tc.children) {
				t.Errorf("round trip children = %v, want %v", got.children, tc.children)
			}
		})
	}
}

// Pins the on-disk byte order for the encoder, mirroring TestDecodeHeaderLittleEndian.
func TestEncodeHeaderLittleEndian(t *testing.T) {
	const order = 2
	n := &Node{Children: []uint16{0x0102}}
	payload := make([]byte, PAGE_SIZE)

	encodeHeader(payload, n, order)

	// child 0 = 0x0102 -> low byte first
	if payload[5] != 0x02 || payload[6] != 0x01 {
		t.Errorf("child bytes = % x, want 02 01", payload[5:7])
	}
	// freeTuple = 4096 = 0x1000 -> 00 10
	if payload[3] != 0x00 || payload[4] != 0x10 {
		t.Errorf("freeTuple bytes = % x, want 00 10", payload[3:5])
	}
}

func TestSizeCheck(t *testing.T) {
	h := &header{}
	t.Log(binary.Size(reflect.TypeOf(h.children).Elem())) // -1 or 2?
	t.Log(h.hspace(2))                                    // 15 expected
}
