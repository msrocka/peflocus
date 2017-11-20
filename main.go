package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/msrocka/ilcd"
)

func main() {
	zips := getZips()
	if len(zips) == 0 {
		log.Println("No zip files found.")
		return
	}

	log.Println("Found", len(zips), "zip files for conversion")
	for _, name := range zips {
		log.Println("Convert zip file", name)

		sourcePath := filepath.Join("zips", name)
		targetPath := filepath.Join("zips", "peflocus_"+name)
		deleteExisting(targetPath)

		reader, err := ilcd.NewZipReader(sourcePath)
		if err != nil {
			log.Fatalln("Failed to read zip", sourcePath, ":", err)
		}
		defer reader.Close()

		writer, err := ilcd.NewZipWriter(targetPath)
		if err != nil {
			log.Fatalln("Failed to create zip writer for", targetPath, ":", err)
		}
		defer writer.Close()

		err = reader.EachEntry(func(name string, data []byte) error {
			return writer.WriteEntry(name, data)
		})
		if err != nil {
			log.Fatalln("Failed to convert zip", err)
		}
	}
}

func getZips() []string {
	zips, err := ioutil.ReadDir("zips")
	if err != nil {
		log.Fatalln("Failed to read folder `zips`. Does it exits?")
	}

	log.Println("Scan folder zips")
	var names []string
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
		names = append(names, name)
	}
	return names
}

func deleteExisting(file string) {
	_, err := os.Stat(file)
	if os.IsNotExist(err) {
		return
	}
	log.Println("Delete existing file", file)
	if err = os.Remove(file); err != nil {
		log.Fatalln("Failed to delete existing file", file, ":", err)
	}
}
