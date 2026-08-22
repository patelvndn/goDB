package pager

import (
	"errors"
	"io"
	"os"
	"strconv"
)

type Page struct {
	id uint32
	data []byte
	dirty bool
}

// recall, pager does not need to know what each byte means
// sole purpose is to read and return bytes for other systems to use 
type Pager struct {
	file *os.File
	pageSize int
	pages map[uint32]*Page // cache
}

const (
	BASE_10 = 10
)

func New(fileName string, pageSize int) (*Pager, error) {

	// will create a page if it doesn't exists
	file, err := os.OpenFile(fileName, os.O_RDWR | os.O_CREATE, 0644)

	if err != nil {
		return nil, err
	}

	p := &Pager{file: file, pageSize: pageSize, pages: make(map[uint32]*Page)}
	
	return p, nil 
} 

func (p *Pager) ReadPage(id uint32) (*Page, error){
	// lets keep it simple for now
	// no metadata, just straight pages

	if _, ok := p.pages[id]; ok {
    	// Key exists in the map
		return p.pages[id], nil 
	}

	info, err := p.file.Stat()

	if err != nil {
		return nil, err
	}

	if (info.Size()) < (int64(id+1) * int64((p.pageSize))) {
		return nil, errors.New("error: no page exists for that id, largest page id is " + strconv.FormatInt((info.Size() / int64(p.pageSize)-1), BASE_10))
	}

	// now seek and read bytes
	_, err = p.file.Seek(int64(p.pageSize) * int64(id), io.SeekStart)

	if err != nil {
		return nil, err
	}
	data := make([]byte, p.pageSize)
	bytesRead, err := p.file.Read(data)

	if err != nil {
		return nil, err
	}

	if bytesRead != int(p.pageSize) {
		return nil, errors.New("error: wasn't able to read all data to the page")
	}

	page := &Page{
		data: data,
		id: id,
		dirty: false,
	}
	
	p.pages[id] = page
	return page, nil

}

func (p *Pager) WritePage(id uint32, newData []byte) (*Page, error){
	_, err := p.ReadPage(id)

	if err != nil {
		return nil, err
	}

	if len(newData) > p.pageSize {
		return nil, errors.New("error: data exceeds page size")
	}

	if len(newData) < p.pageSize {
		newData = append(newData, make([]byte, p.pageSize - len(newData))...)
	}

	// create a copy in case flush fails
	candidate := &Page{id: id, data: newData, dirty: true}
	err = p.flush(candidate)

	if err != nil {
		return nil, err
	}

	candidate.dirty = false
	p.pages[id] = candidate

	return candidate, nil
}

func (p *Pager) AllocatePage() (*Page, error) {
	
	info, err := p.file.Stat()

	if err != nil {
		return nil, err
	}

	prevFileSize := info.Size()

	// get page id
	id := uint32(info.Size()) / uint32(p.pageSize)

	// go to end of file and allocate pageSize number of bytes
	err = allocateBytes(p, prevFileSize)
	if err != nil{
		return nil, err
	}
	
	page := &Page{
		id: uint32(id),
		data: make([]byte, p.pageSize),
		dirty: false,
	} 
	
	p.pages[id] = page
	return page, nil
}

// persist file in physical storage
func (p *Pager) Sync() error {
	return p.file.Sync()
}

// close the file
func (p *Pager) Close() error {
	
	var err error

	for _, page := range(p.pages){
	
		if page.dirty {
			err = p.flush(page)
		}

		if err != nil {
			return err
		}
	}

	err = p.Sync()

	if err != nil {
		return err
	}

	return p.file.Close()
}

func allocateBytes(p *Pager, prevFileSize int64) error {
	data := make([]byte, p.pageSize)

	if _, err := p.file.Seek(0, io.SeekEnd); err != nil {
		return err
	}

	bytesWritten, err := p.file.Write(data)
	if err != nil {
		return err
	}

	if bytesWritten != int(p.pageSize) {
		return errors.New("error: wasn't able to allocate enough data to a page")
	}

	info, err := p.file.Stat()

	if err != nil {
		return err
	}

	if info.Size() != prevFileSize+int64(p.pageSize) {
		return errors.New("error: wasn't able to allocate enough data to a page")
	}

	return nil
}

func (p *Pager) flush(page *Page) error {
	
	if _, err := p.file.Seek(int64(page.id)*int64(p.pageSize), io.SeekStart); err != nil {
		return err
	}

	bytesWritten, err := p.file.Write(page.data)

	if bytesWritten != p.pageSize {
		// todo: roll back if this happens
		return errors.New("error: not all bytes were written, partial write may have occurred")
	} else if err != nil {
		return err
	}

	return nil 
}
