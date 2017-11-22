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

	flowPrefix := ""
	err = reader.EachEntry(func(name string, data []byte) error {
		if flowPrefix == "" && ilcd.IsFlowPath(name) {
			flowPrefix = strings.Split(name, "flows")[0] + "flows/"
		}
		converted, err := m.flowMap.MapFlows(name, data)
		if err != nil {
			return err
		}
		return writer.WriteEntry(name, converted)
	})
	if err != nil {
		log.Println("ERROR: Failed to convert zip", err)
		return
	}

	gen := FlowGen{
		flowMap: m.flowMap,
		prefix:  flowPrefix,
		reader:  reader,
		writer:  writer}
	gen.Generate()

	m.flowMap.ResetStats()
}
