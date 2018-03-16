package main

import (
	"log"
	"path/filepath"
	"strings"

	"github.com/msrocka/ilcd"
)

// FlowMapper applies a flow mapping to the ILCD packages in a working directory.
type FlowMapper struct {
	workdir string
	mapfile string

	flowMap *FlowMap
}

// NewFlowMapper initializes a new flow mapper from the given arguments.
func NewFlowMapper(args *Args) *FlowMapper {
	return &FlowMapper{workdir: args.WorkDir, mapfile: args.MapFile}
}

// Run executes the flow mapping.
func (m *FlowMapper) Run() {
	m.flowMap = ReadFlowMap(m.mapfile)
	zips := GetZipNames(m.workdir)
	for _, name := range zips {
		sourcePath := filepath.Join(m.workdir, name)
		targetPath := filepath.Join(m.workdir, "peflocus_"+name)
		DeleteExisting(targetPath)
		log.Println("INFO: map flows in", sourcePath, "to", targetPath)
		m.doIt(sourcePath, targetPath)
	}
}

func (m *FlowMapper) doIt(sourcePath, targetPath string) {

	// create the reader and writer
	reader, err := ilcd.NewZipReader(sourcePath)
	if err != nil {
		log.Println("ERROR: Failed to read zip", sourcePath, ":", err)
		return
	}
	defer reader.Close()
	writer, err := ilcd.NewZipWriter(targetPath)
	if err != nil {
		log.Println("ERROR: Failed to create zip writer for",
			targetPath, ":", err)
		return
	}
	defer writer.Close()

	// map the flows in the data sets
	flowPrefix := ""
	reader.Map(writer, func(zipFile *ilcd.ZipFile) (string, []byte) {
		path := zipFile.Path()
		data, err := zipFile.Read()
		if err != nil {
			log.Println("ERROR: Failed to read entry", path, err)
			return "", nil
		}
		if flowPrefix == "" && ilcd.IsFlowPath(path) {
			flowPrefix = strings.Split(path, "flows")[0] + "flows/"
		}
		converted, err := m.flowMap.MapFlows(path, data)
		if err != nil {
			log.Println("ERROR: Failed to map flows in", path, err)
			return "", nil
		}
		return path, converted
	})

	gen := FlowGen{
		flowMap: m.flowMap,
		prefix:  flowPrefix,
		reader:  reader,
		writer:  writer}
	gen.Generate(true)

	m.flowMap.ResetStats()
}