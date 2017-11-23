package main

import (
	"log"
	"path/filepath"

	"github.com/msrocka/ilcd"
)

// CalcImpact contains the information of an LCIA category for the calculation.
type CalcImpact struct {
	UUID    string
	Name    string
	Method  string
	Factors map[string]float64
}

// NewCalcImpact initializes a LCIA category for the calculation. It returns nil
// if something went wrong.
func NewCalcImpact(m *ilcd.Method) *CalcImpact {
	if m == nil || m.Info == nil {
		return nil
	}
	name := ""
	if m.Info.Name != nil {
		name = m.Info.Name.Get("en")
	}
	name += "; " + m.Info.ImpactIndicator
	if len(m.Factors) == 0 {
		log.Println("ERROR: No LCIA factors in method", name)
		return nil
	}
	impact := &CalcImpact{
		UUID:    m.Info.UUID,
		Name:    name,
		Method:  m.Info.Methodology,
		Factors: make(map[string]float64)}
	for _, f := range m.Factors {
		impact.addFactor(&f)
	}
	log.Println("INFO: Loaded method", name, "with", len(impact.Factors), "factors")
	return impact
}

func (impact *CalcImpact) addFactor(f *ilcd.ImpactFactor) {
	if impact == nil || f == nil || f.Flow == nil || f.MeanValue == 0.0 {
		return
	}
	value := f.MeanValue
	if f.Direction == "Input" {
		value = -f.MeanValue
	}
	flowID := NormKey(f.Flow.UUID)
	location := NormKey(f.Location)
	if location == "" {
		impact.addFactorValue(flowID, value)
	} else {
		impact.addFactorValue(location+"/"+flowID, value)
	}
}

func (impact *CalcImpact) addFactorValue(key string, value float64) {
	oldVal, contains := impact.Factors[key]
	if contains {
		log.Println(" ... ERROR: multiple LCIA factors for same flow",
			key, ": ", oldVal, value)
		return
	}
	impact.Factors[key] = value
}

// Calculator calculates the LCIA results of a set of ILCD packages.
type Calculator struct {
	workdir string
}

// NewCalculator initializes a new calculator from the given args
func NewCalculator(args *Args) *Calculator {
	return &Calculator{workdir: args.WorkDir}
}

// Run executes the calculation.
func (c *Calculator) Run() {
	zips := GetZipNames(c.workdir)
	for _, name := range zips {
		log.Println("INFO: Calculate the results of", name)
		path := filepath.Join(c.workdir, name)
		reader, err := ilcd.NewZipReader(path)
		if err != nil {
			log.Println("ERROR: Failed to read zip", name, err)
			continue
		}
		impacts := c.readImpacts(reader)
		if len(impacts) == 0 {
			log.Println("INFO: No LCIA methods found in package",
				name, "nothing to caluclate")
			continue
		}
	}
}

func (c *Calculator) readImpacts(r *ilcd.ZipReader) []*CalcImpact {
	var impacts []*CalcImpact
	err := r.EachMethod(func(method *ilcd.Method) bool {
		impact := NewCalcImpact(method)
		if impact != nil {
			impacts = append(impacts, impact)
		}
		return true
	})
	if err != nil {
		log.Println("ERROR: failed to read LCIA methods", err)
		return nil
	}
	return impacts
}
