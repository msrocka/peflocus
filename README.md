# peflocus
In the ILCD data sets of the [PEF pilots](http://ec.europa.eu/environment/eussd/smgp/ef_pilots.htm#pef)
elemantary flows are partly regionalized via the `location` element in inputs
and outputs of processes or in characterization factors of LCIA method data sets.
[openLCA](http://www.openlca.org/) has another approach for
[regionalization](https://www.openlca.org/wp-content/uploads/2016/08/Regionalized-LCIA-in-openLCA.pdf)
and does not support these `location` elements in process exchanges and LCIA
factors.

`peflocus` is a command line tool that maps flows in such regionalized exchanges
and LCIA factors to new flows. It takes an ILCD zip packages, links such
exchanges and LCIA factors to flows from a mapping file, and adds the used flows
to that package.

## Usage
Put the `peflocus` executable next to a CSV file with the flow mappings
`flow_mapping.csv` and a folder `zips` that contains the ILCD zip files that
you want to convert. Then, just open a command line and run the executable. For
each zip file in the zips folder it will then create a new zip file with a
`peflocus_` prefix where the flow mappings are applied.

## ...

```xml
<factor>
  <referenceToFlowDataSet
    refObjectId="03b56eb6-cc68-4251-9317-06878cb27dff"
    type="flow data set"
    uri="../flows/03b56eb6-cc68-4251-9317-06878cb27dff.xml"
    version="03.00.000">
    <common:shortDescription xml:lang="en">from arable, irrigated,</common:shortDescription>
  </referenceToFlowDataSet>
  <location>AD</location>
  <exchangeDirection>Input</exchangeDirection>
  <meanValue>-128</meanValue>
</factor>
```

```xml
<exchange dataSetInternalID="1">
  <referenceToFlowDataSet
    type="flow data set"
    refObjectId="e3abf13f-3bb9-4e52-b72b-9bd276625c55"
    version="01.00.000"
    uri="../flows/e3abf13f-3bb9-4e52-b72b-9bd276625c55">
    <common:shortDescription xml:lang="en">1,1,1,2-Tetrachloroethane</common:shortDescription>
  </referenceToFlowDataSet>
	<location>PL</location>
  <exchangeDirection>Output</exchangeDirection>
  <meanAmount>1.0</meanAmount>
  <resultingAmount>1.0</resultingAmount>
</exchange>
```