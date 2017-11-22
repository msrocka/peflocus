package main

import (
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/msrocka/ilcd"
)

// DeleteExisting deletes the given file if it exists.
func DeleteExisting(file string) {
	_, err := os.Stat(file)
	if os.IsNotExist(err) {
		return
	}
	log.Println("INFO: Delete existing file", file)
	if err = os.Remove(file); err != nil {
		log.Println("ERROR: Failed to delete existing file", file, ":", err)
	}
}

// GetZipNames returns the names of the zip files in the given folder. It
// excludes all zip files that start with the prefix `peflocus_`.
func GetZipNames(folder string) []string {
	zips, err := ioutil.ReadDir(folder)
	if err != nil {
		log.Println("ERROR: Failed to read zip files from folder", folder, err)
		return nil
	}

	log.Println("INFO: Get zip files from", folder)
	var names []string
	for _, zip := range zips {
		name := zip.Name()
		if zip.IsDir() || !strings.HasSuffix(name, ".zip") {
			log.Println(" ... ignore file", name)
			continue
		}
		if strings.HasPrefix(name, "peflocus_") {
			log.Println(" ... ignore file", name, "(this may be overwritten)")
			continue
		}
		names = append(names, name)
	}
	log.Println(" ... found", len(names), "files")
	return names
}

// GetPathType returns the data set type of the given path.
func GetPathType(path string) ilcd.DataSetType {
	p := strings.ToLower(path)
	isXML := strings.HasSuffix(p, ".xml")
	for _, t := range ilcd.DataSetTypes() {
		if strings.Contains(p, t.Folder()) {
			if t == ilcd.ExternalDocType || isXML {
				return t
			}
		}
	}
	return ilcd.UnknownType
}
