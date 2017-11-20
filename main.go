package main

import (
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
)

func getZips() []string {
	zips, err := ioutil.ReadDir("zips")
	if err != nil {
		log.Fatalln("Failed to read folder `zips`. Does it exits?")
	}

	log.Println("Scan folder zips")
	var paths []string
	for _, zip := range zips {
		name := zip.Name()
		if zip.IsDir() && !strings.HasSuffix(name, ".zip") {
			log.Println("ignore file", name)
			continue
		}
		if strings.HasPrefix(name, "peflocus_") {
			log.Println("ignore file", name, "(this may be overwritten)")
			continue
		}
		paths = append(paths, filepath.Join("zips", name))
	}
	return paths
}

func main() {
	zips := getZips()
	if len(zips) == 0 {
		log.Println("No zip files found.")
		return
	}

	log.Println("Found", len(zips), "zip files for conversion")
	for _, path := range zips {
		log.Println("Convert zip file", path)
	}
}
