package main

import (
	"errors"
	"log"
	"strings"

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
func (gen *FlowGen) Generate(forMapped bool) {
	log.Println("INFO: Generate new flows")
	generated := 0

	for key := range gen.flowMap.used {

		var mapEntry *FlowMapEntry
		if forMapped {
			mapEntry = gen.flowMap.mappings[key]
		} else {
			mapEntry = gen.flowMap.unmappings[key]
		}
		if mapEntry == nil {
			panic("This should not happen ?" + key)
		}

		var existingID string
		var genID string
		location := ""
		if forMapped {
			existingID = mapEntry.OldID
			genID = mapEntry.NewID
			location = mapEntry.Location
		} else {
			existingID = mapEntry.NewID
			genID = mapEntry.OldID
		}

		flowEntry := gen.reader.FindDataSet(ilcd.FlowDataSet, existingID)
		if flowEntry == nil {
			log.Println(" ... ERROR: flow", existingID,
				"mapped and used but could not find it in package")
			continue
		}
		data, err := flowEntry.Read()
		if err != nil {
			log.Println(" ... ERROR: Failed to read flow", flowEntry.Path(), err)
			continue
		}

		data, err = gen.doIt(genID, location, data)
		if err != nil {
			log.Println(" ... ERROR: Failed to create flow", genID,
				"from", existingID, err)
			continue
		}

		newEntry := gen.prefix + genID + ".xml"
		if err = gen.writer.Write(newEntry, data); err != nil {
			log.Println(" ... ERROR: Failed to write new flow", newEntry, err)
			continue
		}

		generated++
	}
	log.Println(" ... generated", generated, "new flows")
}

// if location="" it is expected that this function runs in `unmapped` mode,
// which will remove flow locations from the name; other wise the location code
// is added to the name -> this is a quickly hack to get this done o_O
func (gen *FlowGen) doIt(id string, location string, data []byte) ([]byte, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(data); err != nil {
		return nil, err
	}
	uuid := doc.FindElement("./flowDataSet/flowInformation/dataSetInformation/UUID")
	if uuid == nil {
		return nil, errors.New("No UUID element")
	}
	uuid.SetText(id)
	npath := "./flowDataSet/flowInformation/dataSetInformation/name/baseName"
	for _, elem := range doc.FindElements(npath) {
		name := elem.Text()
		if location != "" {
			elem.SetText(name + " - " + location)
		} else {
			locIdx := strings.LastIndex(name, " - ")
			if locIdx > 0 {
				name = name[:locIdx]
				elem.SetText(name)
			}
		}
	}
	return doc.WriteToBytes()
}
