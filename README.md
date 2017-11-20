# peflocus

* Converts all `*.zip` files that are located under `zips`;  ignores files that
    with `peflocus_`
* 

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