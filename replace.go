package main

import (
	"log"

	"github.com/beevik/etree"
	"github.com/msrocka/ilcd"
)

func replaceFlows(entry string, data []byte) ([]byte, error) {
	if ilcd.IsMethodPath(entry) {
		return replaceInMethod(data)
	}
	if ilcd.IsProcessPath(entry) {
		return replaceInProcess(data)
	}
	return data, nil
}

func replaceInMethod(data []byte) ([]byte, error) {
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
	return data, nil
}

func replaceInProcess(data []byte) ([]byte, error) {

	return data, nil
}
