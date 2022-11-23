package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/mateuszmierzwinski/filescanner"

	"log"
	"path/filepath"
	"sync"
)

const snooze time.Duration = 600

var (
	dest           string
	ext            string
	header         string
	tempFolderName string
	collection     string
	sepr           string
	errors         string = ""
	cpy            bool   = false
	fileList       bool   = false
)

func sCopy(fPath string, fName string, dPath string, wg *sync.WaitGroup) {
	if wg != nil {
		wg.Add(1)
	}
	fileName := fPath + "/" + fName
	destPath := dPath + "/" + fName
	fin, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer fin.Close()

	fout, err := os.Create(destPath)
	if err != nil {
		log.Fatal(err)
	}
	defer fout.Close()

	_, err = io.Copy(fout, fin)

	if err != nil {
		log.Fatal(err)
	}
	wg.Done()
}

func sArch(source, target string) error {
	// 1. Create a ZIP file and zip.Writer
	f, err := os.Create(target)
	if err != nil {
		return err
	}
	defer f.Close()

	writer := zip.NewWriter(f)
	defer writer.Close()

	// 2. Go through all the files of the source
	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 3. Create a local file header
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		// set compression
		header.Method = zip.Deflate

		// 4. Set relative path of a file as the header name
		header.Name, err = filepath.Rel(filepath.Dir(source), path)
		if err != nil {
			return err
		}
		if info.IsDir() {
			header.Name += "/"
		}

		// 5. Create writer for the file header and save content of the file
		headerWriter, err := writer.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(headerWriter, f)
		return err
	})
}

func catchParent(fullP string) string {
	parent := strings.Split(fullP, sepr)
	c := len(parent) - 1
	return parent[c]
}

func tracker(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s", name, elapsed)
	if fileList {
		fList, _ := os.Create(dest)
		header += elapsed.String() + "\n\n"
		fList.WriteString(header + collection)
		if errors != "" {
			eList, _ := os.Create("errors.log")
			eList.WriteString(errors)
		}
	}
}

func race() string {
	defer tracker(time.Now(), "Execution")
	if runtime.GOOS == "windows" {
		sepr = "\\"
	} else {
		sepr = "/"
	}
	sprint()
	return ""
}

func sprint() {
	if len(os.Args) == 4 {
		dest = os.Args[3]
		if strings.Contains(os.Args[3], ".") {
			fileList = true
			if _, er := os.Stat(dest); !os.IsNotExist(er) {
				os.Remove(dest)
			}
		} else {
			cpy = true
			if strings.Contains(dest, ".zip") {
				tempFolderName = strings.Replace(dest, ".zip", "", 1)
			} else {
				tempFolderName = dest
				dest = dest + ".zip"
			}
			if _, err := os.Stat(tempFolderName); os.IsNotExist(err) {
				os.Mkdir(tempFolderName, 0775)
			}
		}
	}
	/*	Logic to Search for Directories instead of files
		if strings.HasPrefix(os.Args[2], "d.") {
			ext = strings.Replace(os.Args[2], "d.", "", -1)
			getDir = true
		} else {
			// logic
		}
	*/
	path := os.Args[1]
	ext = os.Args[2]
	wg := sync.WaitGroup{}
	f := filescanner.Scanner{}
	if strings.Contains(ext, ",") {
		exts := strings.Split(ext, ",")
		if fileList {
			header += "Hunting \n		Extentions: " + ext + "	|		Path: " + path + "		| Total Time Elapsed: "
		}
		log.Println("Hunting \n		Extentions: " + ext + "	|		Path: " + path + "\n\n")
		for _, target := range exts {
			resStream, errStream := f.Search(path, target, &wg)
			// Found files handling
			go func() {
				for {
					foundFile := <-resStream
					cPath := filepath.Join(foundFile.Path, foundFile.Name)
					log.Printf("Found file: %s", cPath)
					if fileList {
						collection += "Path: " + cPath + "\n"
					}
					if foundFile.Path != dest {
						sp := catchParent(foundFile.Path)
						if cpy {
							if _, err := os.Stat(sp); os.IsNotExist(err) {
								os.Mkdir(dest+sepr+sp, 0775)
								go sCopy(foundFile.Path, foundFile.Name, dest+"\\"+sp, &wg)
							}
							if _, err := os.Stat(sp); !os.IsNotExist(err) {
								go sCopy(foundFile.Path, foundFile.Name, dest+"\\"+sp, &wg)
							}
						}
					}
				}
			}()

			// Errors handling
			go func() {
				for {
					errorOccured := <-errStream
					if fileList {
						errors += fmt.Sprintf("Processing error at %s: %s", errorOccured.Path, errorOccured.Err.Error()) + "\n"
					}
					log.Printf("Processing error at %s: %s", errorOccured.Path, errorOccured.Err.Error())
				}
			}()
		}
	} else {
		log.Println("Hunting Extention: " + ext + "\n\n")
		resStream, errStream := f.Search(path, ext, &wg)
		if fileList {
			header += "Hunting \n		Extention: " + ext + "		| Path: " + path + "		| Total Time Elapsed: "
		}
		// Found files handling
		go func() {
			for {
				foundFile := <-resStream
				cPath := filepath.Join(foundFile.Path, foundFile.Name)
				log.Printf("Found file: %s", cPath)
				if fileList {
					collection += "Path: " + cPath + "\n"
				}
				if foundFile.Path != dest {
					sp := catchParent(foundFile.Path)
					if cpy {
						if _, err := os.Stat(sp); os.IsNotExist(err) {
							os.Mkdir(dest+sepr+sp, 0775)
							go sCopy(foundFile.Path, foundFile.Name, dest+"\\"+sp, &wg)
						}
						if _, err := os.Stat(sp); !os.IsNotExist(err) {
							go sCopy(foundFile.Path, foundFile.Name, dest+"\\"+sp, &wg)
						}
					}
				}
			}
		}()

		// Errors handling
		go func() {
			for {
				errorOccured := <-errStream
				log.Printf("Processing error at %s: %s", errorOccured.Path, errorOccured.Err.Error())
				if fileList {
					errors += fmt.Sprintf("Processing error at %s: %s", errorOccured.Path, errorOccured.Err.Error()) + "\n"
				}
			}
		}()
	}
	// this wait group waits for FileScanner to finish searching
	wg.Wait()
	time.Sleep(snooze)
	if cpy {
		sArch(tempFolderName, dest)
		os.RemoveAll(dest)
	}
}

func main() {
	if len(os.Args) == 1 {
		fmt.Println("  __    ____ _____  __   ____ _____  __    ___  \n", "/ /`_ | |_   | |  ( (` | |_   | |  / /`_ / / \\ \n", "\\_\\_/ |_|__  |_|  _)_) |_|__  |_|  \\_\\_/ \\_\\_/ ")
		fmt.Println("\n Required args: \n                    <path>  /home/username / C:\\\\  \n", "     <fileName/extention>  log \n", "\n  Optional arguments: \n             <file/folder>  outfile.txt / folderName\n\n", "eg. getsetgo C:\\\\ payslip payslips.txt")
		os.Exit(0)
	}
	race()
}
