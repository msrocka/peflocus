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

		e := gen.flowMap.mappings[key]
		data, err := gen.reader.GetFlowData(e.OldID)
		if err != nil || data == nil {
			log.Println(" ... ERROR: flow", e.OldID,
				"mapped and used but could not it from package", err)
			continue
		}
		data, err = gen.doIt(e, data)
		if err != nil {
			log.Println(" ... ERROR: Failed to create flow",
				e.NewID, "from", e.OldID, err)
			continue
		}

		newEntry := gen.prefix + e.NewID + ".xml"
		if err = gen.writer.WriteEntry(newEntry, data); err != nil {
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
