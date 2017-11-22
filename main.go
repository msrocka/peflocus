package main

import (
	"log"
	"path/filepath"
	"strings"

	"github.com/msrocka/ilcd"
)

func main() {
	log.SetFlags(0)
	args := ReadArgs()
	switch args.Command {
	case "map":
		NewFlowMapper(args).Run()
	default:
		log.Fatalln("ERROR: Unknown command", args.Command)
	}
}

func runConversion(name string, flowMap *FlowMap) {
	sourcePath := filepath.Join("zips", name)
	targetPath := filepath.Join("zips", "peflocus_"+name)
	DeleteExisting(targetPath)

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
