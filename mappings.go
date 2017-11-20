package main

import (
	"encoding/csv"
	"log"
	"os"
	"strings"

	"github.com/beevik/etree"
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

func readMappings() *FlowMap {
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

func (m *FlowMap) reset() {
	m.used = make(map[string]bool)
}

func (m *FlowMap) onFactor(e *etree.Element) {
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
		log.Fatalln(" ... Invalid LCIA factors found")
	}
	idAttr := flowRef.SelectAttr("refObjectId")
	if idAttr == nil {
		log.Fatalln(" ... Invalid LCIA factors found")
	}
	key := getKey(location, idAttr.Value)
	mapping := m.mappings[key]
	if mapping == nil {
		log.Fatalln("Missing flow mapping for", idAttr.Value, " -> ", location)
	}
	idAttr.Value = mapping.NewID
	uriAttr := flowRef.SelectAttr("uri")
	if uriAttr != nil {
		uriAttr.Value = "../flows/" + mapping.NewID + ".xml"
	}
	m.used[key] = true
}
