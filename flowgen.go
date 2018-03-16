package main

import (
	"errors"
	"log"

	"github.com/beevik/etree"
	"github.com/msrocka/ilcd"
)

// FlowGen generates the new flow data sets after a mapping.
type FlowGen struct {
	prefix  string
	reader  *ilcd.ZipReader
	writer  *ilcd.ZipWriter
	flowMap *FlowMap
}

// Generate creates the mapped flow in the target package.
func (gen *FlowGen) Generate() {
	log.Println("INFO: Generate new flows")
	generated := 0

	for key := range gen.flowMap.used {

		mapEntry := gen.flowMap.mappings[key]

		flowEntry := gen.reader.FindDataSet(ilcd.FlowDataSet, mapEntry.OldID)
		if flowEntry == nil {
			log.Println(" ... ERROR: flow", mapEntry.OldID,
				"mapped and used but could not find it in package")
			continue
		}
		data, err := flowEntry.Read()
		if err != nil {
			log.Println(" ... ERROR: Failed to read flow", flowEntry.Path(), err)
			continue
		}

		data, err = gen.doIt(mapEntry, data)
		if err != nil {
			log.Println(" ... ERROR: Failed to create flow", mapEntry.NewID,
				"from", mapEntry.OldID, err)
			continue
		}

		newEntry := gen.prefix + mapEntry.NewID + ".xml"
		if err = gen.writer.Write(newEntry, data); err != nil {
			log.Println(" ... ERROR: Failed to write new flow", newEntry, err)
			continue
		}

		generated++
	}
	log.Println(" ... generated", generated, "new flows")
}

func (gen *FlowGen) doIt(e *FlowMapEntry, data []byte) ([]byte, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(data); err != nil {
		return nil, err
	}

	uuid := doc.FindElement("./flowDataSet/flowInformation/dataSetInformation/UUID")
	if uuid == nil {
		return nil, errors.New("No UUID element")
	}
	uuid.SetText(e.NewID)

	npath := "./flowDataSet/flowInformation/dataSetInformation/name/baseName"
	for _, elem := range doc.FindElements(npath) {
		name := elem.Text()
		elem.SetText(name + " - " + e.Location)
	}

	return doc.WriteToBytes()
}
