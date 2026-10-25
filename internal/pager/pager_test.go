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

	if _, err := p.AllocatePage(); err != nil {
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

	seen := make(map[uint16]bool)
	for i := 0; i < 5; i++ {
		page, err := p.AllocatePage()
		if err != nil {
			t.Fatalf("AllocatePage() call %d returned error: %v", i, err)
		}
		if seen[page.Id] {
			t.Errorf("AllocatePage() call %d returned id %d, which was already handed out", i, page.Id)
		}
		seen[page.Id] = true
	}

	if len(seen) != 5 {
		t.Errorf("expected 5 distinct page ids after 5 allocations, got %d distinct ids: %v", len(seen), seen)
	}
}

// --- ReadPage ---

func TestReadPage_ReturnsDataForJustAllocatedPage(t *testing.T) {
	p, _ := newTestPager(t)

	allocated, err := p.AllocatePage()
	if err != nil {
		t.Fatalf("AllocatePage() returned error: %v", err)
	}

	page, err := p.ReadPage(allocated.Id)
	if err != nil {
		t.Fatalf("ReadPage(%d) returned error for a page that was just allocated: %v", allocated.Id, err)
	}

	if page.Id != allocated.Id {
		t.Errorf("expected page id %d, got %d", allocated.Id, page.Id)
	}
	if len(page.Data) != testPageSize {
		t.Errorf("expected page data length %d, got %d", testPageSize, len(page.Data))
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

	if _, err := p.AllocatePage(); err != nil {
		t.Fatalf("AllocatePage() failed: %v", err)
	}

	// file only has 1 page on disk; id 5 is far past the end of the file.
	if _, err := p.ReadPage(5); err == nil {
		t.Fatal("expected ReadPage() for an id past the end of the file to return an error, got nil")
	}
}

func TestReadPage_DistinctPagesReturnDistinctData(t *testing.T) {
	p, _ := newTestPager(t)

	firstAlloc, err := p.AllocatePage()
	if err != nil {
		t.Fatalf("AllocatePage() failed: %v", err)
	}
	secondAlloc, err := p.AllocatePage()
	if err != nil {
		t.Fatalf("AllocatePage() failed: %v", err)
	}
	if firstAlloc.Id == secondAlloc.Id {
		t.Fatalf("test setup invalid: both allocations returned the same id %d", firstAlloc.Id)
	}

	dataA := bytes.Repeat([]byte{0xAA}, testPageSize)
	dataB := bytes.Repeat([]byte{0xBB}, testPageSize)

	if _, err := p.WritePage(firstAlloc.Id, dataA); err != nil {
		t.Fatalf("WritePage(%d) failed: %v", firstAlloc.Id, err)
	}
	if _, err := p.WritePage(secondAlloc.Id, dataB); err != nil {
		t.Fatalf("WritePage(%d) failed: %v", secondAlloc.Id, err)
	}

	readA, err := p.ReadPage(firstAlloc.Id)
	if err != nil {
		t.Fatalf("ReadPage(%d) failed: %v", firstAlloc.Id, err)
	}
	readB, err := p.ReadPage(secondAlloc.Id)
	if err != nil {
		t.Fatalf("ReadPage(%d) failed: %v", secondAlloc.Id, err)
	}

	if !bytes.Equal(readA.Data, dataA) {
		t.Errorf("page %d: expected data %x, got %x", firstAlloc.Id, dataA, readA.Data)
	}
	if !bytes.Equal(readB.Data, dataB) {
		t.Errorf("page %d: expected data %x, got %x", secondAlloc.Id, dataB, readB.Data)
	}
}

func TestReadPage_RepeatedCallsReturnSameCachedPointer(t *testing.T) {
	p, _ := newTestPager(t)

	allocated, err := p.AllocatePage()
	if err != nil {
		t.Fatalf("AllocatePage() failed: %v", err)
	}
	data := bytes.Repeat([]byte{0x42}, testPageSize)
	if _, err := p.WritePage(allocated.Id, data); err != nil {
		t.Fatalf("WritePage() failed: %v", err)
	}

	first, err := p.ReadPage(allocated.Id)
	if err != nil {
		t.Fatalf("ReadPage() failed: %v", err)
	}
	second, err := p.ReadPage(allocated.Id)
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

	allocated, err := p.AllocatePage()
	if err != nil {
		t.Fatalf("AllocatePage() failed: %v", err)
	}

	want := []byte("hello world!!!!!") // 16 bytes == testPageSize
	if len(want) != testPageSize {
		t.Fatalf("test data length %d does not match testPageSize %d", len(want), testPageSize)
	}

	if _, err := p.WritePage(allocated.Id, want); err != nil {
		t.Fatalf("WritePage() returned error: %v", err)
	}

	got, err := p.ReadPage(allocated.Id)
	if err != nil {
		t.Fatalf("ReadPage() returned error: %v", err)
	}
	if !bytes.Equal(got.Data, want) {
		t.Errorf("expected data %q, got %q", want, got.Data)
	}
}

func TestWritePage_PersistsToDiskAcrossPagerInstances(t *testing.T) {
	p, path := newTestPager(t)

	allocated, err := p.AllocatePage()
	if err != nil {
		t.Fatalf("AllocatePage() failed: %v", err)
	}

	want := bytes.Repeat([]byte{0xCD}, testPageSize)
	if _, err := p.WritePage(allocated.Id, want); err != nil {
		t.Fatalf("WritePage() failed: %v", err)
	}

	// simulate closing and reopening the database: a fresh Pager, same file,
	// with no in-memory cache carried over.
	p2, err := New(path, testPageSize)
	if err != nil {
		t.Fatalf("failed to reopen pager on same file: %v", err)
	}

	got, err := p2.ReadPage(allocated.Id)
	if err != nil {
		t.Fatalf("ReadPage() on reopened pager failed: %v", err)
	}
	if !bytes.Equal(got.Data, want) {
		t.Errorf("data did not persist to disk: expected %x, got %x", want, got.Data)
	}
}

func TestWritePage_PadsShortDataToPageSize(t *testing.T) {
	p, _ := newTestPager(t)

	allocated, err := p.AllocatePage()
	if err != nil {
		t.Fatalf("AllocatePage() failed: %v", err)
	}

	short := []byte("hi")
	page, err := p.WritePage(allocated.Id, short)
	if err != nil {
		t.Fatalf("WritePage() failed: %v", err)
	}

	if len(page.Data) != testPageSize {
		t.Fatalf("expected page data padded to %d bytes, got %d", testPageSize, len(page.Data))
	}
	if !bytes.Equal(page.Data[:len(short)], short) {
		t.Errorf("expected data to start with %q, got %q", short, page.Data[:len(short)])
	}
	for i := len(short); i < testPageSize; i++ {
		if page.Data[i] != 0 {
			t.Errorf("expected padding byte at index %d to be 0, got %d", i, page.Data[i])
		}
	}
}

func TestWritePage_RejectsDataLargerThanPageSize(t *testing.T) {
	p, _ := newTestPager(t)

	allocated, err := p.AllocatePage()
	if err != nil {
		t.Fatalf("AllocatePage() failed: %v", err)
	}

	tooBig := make([]byte, testPageSize+1)
	if _, err := p.WritePage(allocated.Id, tooBig); err == nil {
		t.Error("expected WritePage() to return an error for oversized data, got nil")
	}
}

func TestWritePage_OnUnallocatedIDReturnsError(t *testing.T) {
	p, _ := newTestPager(t)

	if _, err := p.WritePage(0, make([]byte, testPageSize)); err == nil {
		t.Fatal("expected WritePage() on an unallocated page to return an error, got nil")
	}
}