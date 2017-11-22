# peflocus
`peflocus` is a command line tool to work with the ILCD data sets of the
[PEF pilots](http://ec.europa.eu/environment/eussd/smgp/ef_pilots.htm#pef) (and
also other ILCD packages). It has a set of sub-commands that are described
below. The general usage of the tool is (where command and options are
placeholders; see the command descriptions below):

```bash
peflocus [command] [options]
```

`peflocus` writes log messages to the stderr output. Thus, you can pipe them
into a file via:

```bash
peflocus [command] [options] 2> [path/to/logfile]
```

## The `map` command
The PEF data sets are partly regionalized via the `location` element in
exchanges of processes and characterization factors of LCIA method data sets.
This means that a flow can occur multiple times in exchanges of a process
or in characterization factors of an LCIA method but with different location
codes. The code snippets below show examples of a regionalized exchange and
characterization factors:

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

[openLCA](http://www.openlca.org/) (which has another approach for
[regionalization](https://www.openlca.org/wp-content/uploads/2016/08/Regionalized-LCIA-in-openLCA.pdf))
and other tools may not support these `location` elements. The `map` command
takes a mapping file with of the format `(flow-UUID, location) -> new-flow-UUID`
and applies this to a set of ILCD zip files in a folder. It assigns the new
UUIDs to the exchanges and characterization factors and also creates new flows
for these new IDs. For each zip file `x.zip` it will create a file
`peflocus_x.zip` where these mappings are applied.

The mapping file should be an `utf-8` encoded CSV file (with comma as column
separator) with the following colums: 

* UUID of the ILCD flow
* Location code
* New UUID of the flow

The first line of the file is ignored.

The map command has the following options:

* `-workdir` => The path to the folder with the ILCD zip packages; defaults to
  `zips`
* `-mapfile` => The path to the mapping file that should be used; defaults to
  `flow_mapping.csv`

Thus, the command `peflocus map` is the same as:

```
peflocus map -wordir zips -mapfile flow_mapping.csv
```