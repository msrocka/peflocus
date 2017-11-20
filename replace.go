package main

import (
	"log"

	"github.com/beevik/etree"
	"github.com/msrocka/ilcd"
)

func replaceFlows(entry string, data []byte, flowMap *FlowMap) ([]byte, error) {
	if ilcd.IsMethodPath(entry) {
		return replaceInMethod(data, flowMap)
	}
	if ilcd.IsProcessPath(entry) {
		return replaceInProcess(data, flowMap)
	}
	return data, nil
}

func replaceInMethod(data []byte, flowMap *FlowMap) ([]byte, error) {
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
		flowMap.MapFlow(factor)
	}
	return doc.WriteToBytes()
}

func replaceInProcess(data []byte, flowMap *FlowMap) ([]byte, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(data); err != nil {
		return nil, err
	}
	uuidElem := doc.FindElement("./processDataSet/processInformation/dataSetInformation/UUID")
	if uuidElem != nil {
		log.Println("Replace flows in process method", uuidElem.Text())
	}
	exchanges := doc.FindElements("./processDataSet/exchanges/exchange")
	log.Println(" ... check", len(exchanges), "exchanges")
	for _, e := range exchanges {
		flowMap.MapFlow(e)
	}
	return doc.WriteToBytes()
}
