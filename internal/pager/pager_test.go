package pager

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

const testPageSize = 16

// newTestPager gives every test a fresh, isolated .db file so tests can't
// bleed state into each other.
func newTestPager(t *testing.T) (*Pager, string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")
	p, err := New(path, testPageSize)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	return p, path
}

// --- New ---

func TestNew_CreatesFileIfNotExists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "new.db")

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("test setup invalid: file already exists")
	}

	if _, err := New(path, testPageSize); err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file to exist on disk after New(), got stat error: %v", err)
	}
}

func TestNew_OpensExistingFileWithoutTruncating(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "existing.db")

	seed := bytes.Repeat([]byte{0x11}, testPageSize)
	if err := os.WriteFile(path, seed, 0644); err != nil {
		t.Fatalf("failed to seed file: %v", err)
	}

	if _, err := New(path, testPageSize); err != nil {
		t.Fatalf("New() returned error opening existing file: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file after New(): %v", err)
	}
	if !bytes.Equal(got, seed) {
		t.Errorf("expected New() to preserve existing file contents, got %x want %x", got, seed)
	}
}

// --- AllocatePage ---

func TestAllocatePage_GrowsFileBySinglePageSize(t *testing.T) {
	p, path := newTestPager(t)

	before, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}

	if _, err := p.AllocatePage(0); err != nil {
		t.Fatalf("AllocatePage() returned error: %v", err)
	}

	after, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}

	if grew := after.Size() - before.Size(); grew != int64(testPageSize) {
		t.Errorf("expected file to grow by %d bytes, grew by %d", testPageSize, grew)
	}
}

func TestAllocatePage_ReturnsDistinctIDsAcrossCalls(t *testing.T) {
	p, _ := newTestPager(t)

	seen := make(map[uint32]bool)
	for i := 0; i < 5; i++ {
		page, err := p.AllocatePage(0)
		if err != nil {
			t.Fatalf("AllocatePage() call %d returned error: %v", i, err)
		}
		if seen[page.id] {
			t.Errorf("AllocatePage() call %d returned id %d, which was already handed out", i, page.id)
		}
		seen[page.id] = true
	}

	if len(seen) != 5 {
		t.Errorf("expected 5 distinct page ids after 5 allocations, got %d distinct ids: %v", len(seen), seen)
	}
}

// --- ReadPage ---

func TestReadPage_ReturnsDataForJustAllocatedPage(t *testing.T) {
	p, _ := newTestPager(t)

	allocated, err := p.AllocatePage(0)
	if err != nil {
		t.Fatalf("AllocatePage() returned error: %v", err)
	}

	page, err := p.ReadPage(allocated.id)
	if err != nil {
		t.Fatalf("ReadPage(%d) returned error for a page that was just allocated: %v", allocated.id, err)
	}

	if page.id != allocated.id {
		t.Errorf("expected page id %d, got %d", allocated.id, page.id)
	}
	if len(page.data) != testPageSize {
		t.Errorf("expected page data length %d, got %d", testPageSize, len(page.data))
	}
}

func TestReadPage_UnallocatedIDOnEmptyFileReturnsError(t *testing.T) {
	p, _ := newTestPager(t)

	if _, err := p.ReadPage(0); err == nil {
		t.Fatal("expected ReadPage() on an empty file to return an error, got nil")
	}
}

func TestReadPage_IDBeyondEndOfFileReturnsError(t *testing.T) {
	p, _ := newTestPager(t)

	if _, err := p.AllocatePage(0); err != nil {
		t.Fatalf("AllocatePage() failed: %v", err)
	}

	// file only has 1 page on disk; id 5 is far past the end of the file.
	if _, err := p.ReadPage(5); err == nil {
		t.Fatal("expected ReadPage() for an id past the end of the file to return an error, got nil")
	}
}

func TestReadPage_DistinctPagesReturnDistinctData(t *testing.T) {
	p, _ := newTestPager(t)

	firstAlloc, err := p.AllocatePage(0)
	if err != nil {
		t.Fatalf("AllocatePage() failed: %v", err)
	}
	secondAlloc, err := p.AllocatePage(0)
	if err != nil {
		t.Fatalf("AllocatePage() failed: %v", err)
	}
	if firstAlloc.id == secondAlloc.id {
		t.Fatalf("test setup invalid: both allocations returned the same id %d", firstAlloc.id)
	}

	dataA := bytes.Repeat([]byte{0xAA}, testPageSize)
	dataB := bytes.Repeat([]byte{0xBB}, testPageSize)

	if _, err := p.WritePage(firstAlloc.id, dataA); err != nil {
		t.Fatalf("WritePage(%d) failed: %v", firstAlloc.id, err)
	}
	if _, err := p.WritePage(secondAlloc.id, dataB); err != nil {
		t.Fatalf("WritePage(%d) failed: %v", secondAlloc.id, err)
	}

	readA, err := p.ReadPage(firstAlloc.id)
	if err != nil {
		t.Fatalf("ReadPage(%d) failed: %v", firstAlloc.id, err)
	}
	readB, err := p.ReadPage(secondAlloc.id)
	if err != nil {
		t.Fatalf("ReadPage(%d) failed: %v", secondAlloc.id, err)
	}

	if !bytes.Equal(readA.data, dataA) {
		t.Errorf("page %d: expected data %x, got %x", firstAlloc.id, dataA, readA.data)
	}
	if !bytes.Equal(readB.data, dataB) {
		t.Errorf("page %d: expected data %x, got %x", secondAlloc.id, dataB, readB.data)
	}
}

func TestReadPage_RepeatedCallsReturnSameCachedPointer(t *testing.T) {
	p, _ := newTestPager(t)

	allocated, err := p.AllocatePage(0)
	if err != nil {
		t.Fatalf("AllocatePage() failed: %v", err)
	}
	data := bytes.Repeat([]byte{0x42}, testPageSize)
	if _, err := p.WritePage(allocated.id, data); err != nil {
		t.Fatalf("WritePage() failed: %v", err)
	}

	first, err := p.ReadPage(allocated.id)
	if err != nil {
		t.Fatalf("ReadPage() failed: %v", err)
	}
	second, err := p.ReadPage(allocated.id)
	if err != nil {
		t.Fatalf("ReadPage() failed: %v", err)
	}

	if first != second {
		t.Errorf("expected repeated ReadPage() calls to return the same cached *Page, got different pointers: %p vs %p", first, second)
	}
}

// --- WritePage ---

func TestWritePage_ThenReadPageRoundTrips(t *testing.T) {
	p, _ := newTestPager(t)

	allocated, err := p.AllocatePage(0)
	if err != nil {
		t.Fatalf("AllocatePage() failed: %v", err)
	}

	want := []byte("hello world!!!!!") // 16 bytes == testPageSize
	if len(want) != testPageSize {
		t.Fatalf("test data length %d does not match testPageSize %d", len(want), testPageSize)
	}

	if _, err := p.WritePage(allocated.id, want); err != nil {
		t.Fatalf("WritePage() returned error: %v", err)
	}

	got, err := p.ReadPage(allocated.id)
	if err != nil {
		t.Fatalf("ReadPage() returned error: %v", err)
	}
	if !bytes.Equal(got.data, want) {
		t.Errorf("expected data %q, got %q", want, got.data)
	}
}

func TestWritePage_PersistsToDiskAcrossPagerInstances(t *testing.T) {
	p, path := newTestPager(t)

	allocated, err := p.AllocatePage(0)
	if err != nil {
		t.Fatalf("AllocatePage() failed: %v", err)
	}

	want := bytes.Repeat([]byte{0xCD}, testPageSize)
	if _, err := p.WritePage(allocated.id, want); err != nil {
		t.Fatalf("WritePage() failed: %v", err)
	}

	// simulate closing and reopening the database: a fresh Pager, same file,
	// with no in-memory cache carried over.
	p2, err := New(path, testPageSize)
	if err != nil {
		t.Fatalf("failed to reopen pager on same file: %v", err)
	}

	got, err := p2.ReadPage(allocated.id)
	if err != nil {
		t.Fatalf("ReadPage() on reopened pager failed: %v", err)
	}
	if !bytes.Equal(got.data, want) {
		t.Errorf("data did not persist to disk: expected %x, got %x", want, got.data)
	}
}

func TestWritePage_PadsShortDataToPageSize(t *testing.T) {
	p, _ := newTestPager(t)

	allocated, err := p.AllocatePage(0)
	if err != nil {
		t.Fatalf("AllocatePage() failed: %v", err)
	}

	short := []byte("hi")
	page, err := p.WritePage(allocated.id, short)
	if err != nil {
		t.Fatalf("WritePage() failed: %v", err)
	}

	if len(page.data) != testPageSize {
		t.Fatalf("expected page data padded to %d bytes, got %d", testPageSize, len(page.data))
	}
	if !bytes.Equal(page.data[:len(short)], short) {
		t.Errorf("expected data to start with %q, got %q", short, page.data[:len(short)])
	}
	for i := len(short); i < testPageSize; i++ {
		if page.data[i] != 0 {
			t.Errorf("expected padding byte at index %d to be 0, got %d", i, page.data[i])
		}
	}
}

func TestWritePage_RejectsDataLargerThanPageSize(t *testing.T) {
	p, _ := newTestPager(t)

	allocated, err := p.AllocatePage(0)
	if err != nil {
		t.Fatalf("AllocatePage() failed: %v", err)
	}

	tooBig := make([]byte, testPageSize+1)
	if _, err := p.WritePage(allocated.id, tooBig); err == nil {
		t.Error("expected WritePage() to return an error for oversized data, got nil")
	}
}

func TestWritePage_OnUnallocatedIDReturnsError(t *testing.T) {
	p, _ := newTestPager(t)

	if _, err := p.WritePage(0, make([]byte, testPageSize)); err == nil {
		t.Fatal("expected WritePage() on an unallocated page to return an error, got nil")
	}
}