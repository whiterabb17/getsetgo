package getsetgo

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

/*
FileEntry node describing file path, name and size - applicable in most use cases
*/
type FileEntry struct {
	Path string
	Name string
	Size int64
}

/*
ProcessingError describes error in stream processing including path, that was processed
*/
type ProcessingError struct {
	Path string
	Err  error
}

/*
Scanner returns scanner that seeks trough filesystem and returns files by provided substring
*/
type Scanner struct {
	fileStream chan *FileEntry
	errStream  chan *ProcessingError
}

/*
Search starts file scanning (in separate thread) and returns two channels initialized - FileEntry channel and ProcessingError channel

Inputs are searchPath describing path on the disk to be searched, fileNameSubstr that is a part of name searched (for example .extension) and
wg - WaitGroup used to finish app processing after search processing is done (recursively searched whole tree).
*/
func (f *Scanner) Search(searchPath string, fileNameSubstr string, wg *sync.WaitGroup) (chan *FileEntry, chan *ProcessingError) {
	if wg != nil {
		wg.Add(1)
	}

	// add buffered stream so it can fill channels before pulling (no deadlock)
	f.fileStream, f.errStream = make(chan *FileEntry, 10), make(chan *ProcessingError, 10)

	go f.run(wg, searchPath, fileNameSubstr, false)

	return f.fileStream, f.errStream
}

/*
SearchNB starts file scanning (in separate thread)

Inputs are searchPath describing path on the disk to be searched, fileNameSubstr that is a part of name searched (for example .extension) and fileEntryStream that is channel of FileEntry pointers
*/
func (f *Scanner) SearchNB(searchPath string, fileNameSubstr string, fileEntryStream chan *FileEntry) chan *ProcessingError {

	// add buffered stream so it can fill channels before pulling (no deadlock)
	f.errStream = make(chan *ProcessingError, 10)

	if fileEntryStream == nil {
		f.errStream <- &ProcessingError{
			Path: searchPath,
			Err:  fmt.Errorf("fileEntryStream is nil - no receiver defined for SearchNB() method"),
		}
		return f.errStream
	}

	// handling is up to user
	f.fileStream = fileEntryStream

	go f.run(nil, searchPath, fileNameSubstr, true)

	return f.errStream
}

/*
run function is internal runner method that works as separate goroutine
*/
func (f *Scanner) run(wg *sync.WaitGroup, path string, substr string, closeOnFinish bool) {
	defer func() {
		if wg != nil {
			wg.Done()
		}
	}()

	// start processing
	scanEntriesStream(path, strings.ToLower(substr), f.fileStream, f.errStream)

	if closeOnFinish {
		close(f.fileStream)
	}

	return
}

/*
scanEntriesStream recursively steps through directories and processes
*/
func scanEntriesStream(path, extension string, stream chan *FileEntry, errStream chan *ProcessingError) {
	dir, err := os.ReadDir(path)
	if err != nil {
		errStream <- &ProcessingError{
			Path: path,
			Err:  err,
		}
		return
	}

	for _, dirEntry := range dir {
		name := dirEntry.Name()
		if dirEntry.IsDir() {
			scanEntriesStream(filepath.Join(path, name), extension, stream, errStream)
			continue
		}

		if strings.HasSuffix(strings.ToLower(name), extension) {
			dirEntryInfo, _ := dirEntry.Info()
			stream <- &FileEntry{
				Path: path,
				Name: name,
				Size: dirEntryInfo.Size(),
			}
		}
	}
}
