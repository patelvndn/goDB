package btreeNode

import (
	"encoding/binary"
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