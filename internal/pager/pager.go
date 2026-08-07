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

	if uint32(info.Size()) > id * uint32(p.pageSize) {
		return nil, errors.New("error: no page exists for that id, largest page id is " + strconv.Itoa(len(p.pages)))
	}

	if err != nil {
		return nil, err
	}

	// now seek and read bytes
	p.file.Seek(0, p.pageSize * int(id))
	data := make([]byte, p.pageSize)
	bytesRead, err := p.file.Read(data)

	if bytesRead != int(p.pageSize) {
		return nil, errors.New("error: wasn't able to read all data to the page")
	}

	if err != nil {
		return nil, err
	}

	page := &Page{
		data: data,
		id: id,
		dirty: false,
	}
	
	return page, nil

}

func (p *Pager) WritePage(id uint32, newData []byte) (*Page, error){
	page, err := p.ReadPage(id)

	if err != nil {
		return nil, err
	}

	if len(newData) > p.pageSize {
		return nil, errors.New("error: data exceeds page size")
	}

	if len(newData) < p.pageSize {
		newData = append(newData, make([]byte, p.pageSize - len(newData))...)
	}

	page.data = newData
	page.dirty = true // we can flush rn tbh
	
	p.pages[id] = page

	return page, nil
}

func (p *Pager) AllocatePage(id uint32) (*Page, error) {
	pageId := len(p.pages)

	// get page ID
	// go to end of file and allocate pageSize number of bytes

	data := make([]byte, p.pageSize)
	p.file.Seek(0, io.SeekEnd)
	bytesWritten, err := p.file.Write(data)

	if bytesWritten != int(p.pageSize) {
		return nil, errors.New("error: wasn't able to allocate enough data to a page")
	}

	if err != nil {
		return nil, err
	}

	page := &Page{
		id: uint32(pageId),
		data: make([]byte, p.pageSize),
		dirty: false,
	} 

	return page, nil
}