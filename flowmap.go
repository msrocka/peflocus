package main

import (
	"encoding/csv"
	"log"
	"os"
	"strings"

	"github.com/beevik/etree"
	"github.com/msrocka/ilcd"
)

func getKey(location, uuid string) string {
	key := strings.TrimSpace(location) + "/" + strings.TrimSpace(uuid)
	return strings.ToLower(key)
}

// FlowMapEntry represents a row in the mapping file
type FlowMapEntry struct {
	Location string
	OldID    string
	NewID    string
}

func (e *FlowMapEntry) key() string {
	return getKey(e.Location, e.OldID)
}

// FlowMap contains the flow mappings.
type FlowMap struct {
	mappings map[string]*FlowMapEntry
	used     map[string]bool
}

// ReadFlowMap reads the flow mappings from the file flow_mapping.csv
func ReadFlowMap() *FlowMap {
	log.Println("Read flow mappings from flow_mapping.csv")
	file, err := os.Open("flow_mapping.csv")
	if err != nil {
		log.Fatalln("Failed to read mapping file flow_mapping.csv", err)
	}
	defer file.Close()
	rows, err := csv.NewReader(file).ReadAll()
	if err != nil {
		log.Fatalln("Failed to read mapping file flow_mapping.csv", err)
	}
	fm := FlowMap{
		mappings: make(map[string]*FlowMapEntry),
		used:     make(map[string]bool)}
	for i, row := range rows {
		if i == 0 {
			continue
		}
		e := FlowMapEntry{OldID: row[0], Location: row[1], NewID: row[2]}
		fm.mappings[e.key()] = &e
	}
	return &fm
}

// ResetStats clears the mapping statistics
func (m *FlowMap) ResetStats() {
	m.used = make(map[string]bool)
}

// MapFlows maps the flows in the given data set if it is an LCIA method or
// process
func (m *FlowMap) MapFlows(zipEntry string, data []byte) ([]byte, error) {
	if ilcd.IsMethodPath(zipEntry) {
		return m.MapMethod(data)
	}
	if ilcd.IsProcessPath(zipEntry) {
		return m.MapProcess(data)
	}
	return data, nil
}

// MapMethod maps the flows in the given LCIA method data set.
func (m *FlowMap) MapMethod(data []byte) ([]byte, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(data); err != nil {
		return nil, err
	}
	uuidElem := doc.FindElement("./LCIAMethodDataSet/LCIAMethodInformation/dataSetInformation/UUID")
	if uuidElem != nil {
		log.Println("Replace flows in LCIA method", uuidElem.Text())
	}
	factors := doc.FindElements("./LCIAMethodDataSet/characterisationFactors/factor")
	log.Println(" ... check", len(factors), "factors")
	for _, factor := range factors {
		m.MapFlow(factor)
	}
	return doc.WriteToBytes()
}

// MapProcess maps the flows in the given process data set.
func (m *FlowMap) MapProcess(data []byte) ([]byte, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(data); err != nil {
		return nil, err
	}
	uuidElem := doc.FindElement("./processDataSet/processInformation/dataSetInformation/UUID")
	if uuidElem != nil {
		log.Println("Replace flows in process", uuidElem.Text())
	}
	exchanges := doc.FindElements("./processDataSet/exchanges/exchange")
	log.Println(" ... check", len(exchanges), "exchanges")
	for _, e := range exchanges {
		m.MapFlow(e)
	}
	return doc.WriteToBytes()
}

// MapFlow maps the flow information
func (m *FlowMap) MapFlow(e *etree.Element) {
	locElem := e.FindElement("./location")
	if locElem == nil {
		return
	}
	location := strings.TrimSpace(locElem.Text())
	if location == "" {
		return
	}
	flowRef := e.FindElement("./referenceToFlowDataSet")
	if flowRef == nil {
		log.Println(" ... ERROR: no flow reference found")
		return
	}
	idAttr := flowRef.SelectAttr("refObjectId")
	if idAttr == nil {
		log.Println(" ... ERROR: no flow reference found")
		return
	}
	key := getKey(location, idAttr.Value)
	mapping := m.mappings[key]
	if mapping == nil {
		log.Println(" ... ERROR: Missing flow mapping for",
			idAttr.Value, " -> ", location)
		// TODO: generate UUID and mapping!
		return
	}
	idAttr.Value = mapping.NewID
	uriAttr := flowRef.SelectAttr("uri")
	if uriAttr != nil {
		uriAttr.Value = "../flows/" + mapping.NewID + ".xml"
	}
	m.used[key] = true
}
