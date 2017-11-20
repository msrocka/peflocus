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
	log.SetPrefix("")
	flowMap := ReadFlowMap()

	zips := getZips()
	if len(zips) == 0 {
		log.Println("No zip files found.")
		return
	}

	log.Println("Found", len(zips), "zip files for conversion")
	for _, name := range zips {
		log.Println("Convert zip file", name)
		runConversion(name, flowMap)
	}
}

func runConversion(name string, flowMap *FlowMap) {
	sourcePath := filepath.Join("zips", name)
	targetPath := filepath.Join("zips", "peflocus_"+name)
	deleteExisting(targetPath)

	// create the reader and writer
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

	flowPrefix := ""
	err = reader.EachEntry(func(name string, data []byte) error {
		if flowPrefix == "" && ilcd.IsFlowPath(name) {
			flowPrefix = strings.Split(name, "flows")[0] + "flows/"
		}
		converted, err := flowMap.MapFlows(name, data)
		if err != nil {
			return err
		}
		return writer.WriteEntry(name, converted)
	})
	if err != nil {
		log.Fatalln("Failed to convert zip", err)
	}

	gen := FlowGen{
		flowMap: flowMap,
		prefix:  flowPrefix,
		reader:  reader,
		writer:  writer}
	gen.Generate()

	// TODO: write stats
	flowMap.ResetStats()
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
		if zip.IsDir() || !strings.HasSuffix(name, ".zip") {
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
