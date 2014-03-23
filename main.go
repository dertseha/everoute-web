package main

import (
	encodingJson "encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"runtime"
	"runtime/debug"

	"github.com/gorilla/rpc"
	rpcJson "github.com/gorilla/rpc/json"

	"github.com/dertseha/everoute/travel/capabilities/jumpdrive"
	"github.com/dertseha/everoute/travel/capabilities/jumpgate"
	"github.com/dertseha/everoute/travel/rules/security"
	"github.com/dertseha/everoute/travel/rules/transitcount"
	"github.com/dertseha/everoute/universe"
)

type Table struct {
	Names []string        `json:"names"`
	Data  [][]interface{} `json:"data"`
}

func readTableFile(fileName string) *Table {
	var table Table
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		panic(fmt.Sprintf("Failed to read file <%s>", err))
	}

	err = encodingJson.Unmarshal(content, &table)
	if err != nil {
		panic(fmt.Sprintf("Failed to decode file <%s>", err))
	}

	return &table
}

func getTableColumnIndices(table *Table) map[string]int {
	columns := make(map[string]int)

	for index, name := range table.Names {
		columns[name] = index
	}

	return columns
}

func buildSolarSystems(builder *universe.UniverseBuilder) {
	table := readTableFile("solarSystems.json")
	columns := getTableColumnIndices(table)

	for _, system := range table.Data {
		regionId := universe.Id(reflect.ValueOf(system[columns["regionId"]]).Float())
		constellationId := universe.Id(reflect.ValueOf(system[columns["constellationId"]]).Float())
		solarSystemId := universe.Id(reflect.ValueOf(system[columns["solarSystemId"]]).Float())
		//solarSystemName := reflect.ValueOf(system[columns["solarSystemName"]]).String()
		x := reflect.ValueOf(system[columns["x"]]).Float()
		y := reflect.ValueOf(system[columns["y"]]).Float()
		z := reflect.ValueOf(system[columns["z"]]).Float()
		trueSec := universe.TrueSecurity(reflect.ValueOf(system[columns["security"]]).Float())
		galaxyId := universe.NewEdenId

		if regionId >= 11000000 {
			galaxyId = universe.WSpaceId
		}

		builder.AddSolarSystem(solarSystemId, constellationId, regionId, galaxyId, universe.NewSpecificLocation(x, y, z), trueSec)
	}
}

func buildJumpGates(builder *universe.UniverseBuilder) {
	table := readTableFile("solarSystemJumps.json")
	columns := getTableColumnIndices(table)

	for _, jumpData := range table.Data {
		var fromSolarSystemId = universe.Id(reflect.ValueOf(jumpData[columns["fromSolarSystemId"]]).Float())
		var toSolarSystemId = universe.Id(reflect.ValueOf(jumpData[columns["toSolarSystemId"]]).Float())
		var extension = builder.ExtendSolarSystem(fromSolarSystemId)

		extension.AddJump(jumpgate.JumpType, toSolarSystemId)
	}
}

func prepareUniverse() *universe.UniverseBuilder {
	builder := universe.New().Extend()

	buildSolarSystems(builder)
	buildJumpGates(builder)
	transitcount.ExtendUniverse(builder)
	security.ExtendUniverse(builder)

	jumpdrive.ExtendUniverse(builder, 17.0)

	return builder
}

func initRuntime() {
	numCpu := runtime.NumCPU()
	maxThreads := 250 // Heroku limit: 256

	log.Printf("Initializing runtime to use %d CPUs and %d threads", numCpu, maxThreads)
	runtime.GOMAXPROCS(numCpu)
	debug.SetMaxThreads(maxThreads)
}

func main() {
	initRuntime()

	log.Printf("Building universe...")
	universeBuilder := prepareUniverse()
	universe := universeBuilder.Build()

	log.Printf("Starting server...")
	rpcServer := rpc.NewServer()
	rpcServer.RegisterCodec(rpcJson.NewCodec(), "application/json")
	service := NewRouteService(universe)
	rpcServer.RegisterService(service, "Route")

	http.Handle("/", rpcServer)
	http.ListenAndServe(":3000", nil)
}
