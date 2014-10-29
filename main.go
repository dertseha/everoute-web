package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/gorilla/rpc"
	rpcJson "github.com/gorilla/rpc/json"

	"github.com/dertseha/everoute/travel/capabilities/jumpdrive"
	"github.com/dertseha/everoute/travel/capabilities/jumpgate"
	"github.com/dertseha/everoute/travel/rules/security"
	"github.com/dertseha/everoute/travel/rules/transitcount"
	"github.com/dertseha/everoute/universe"

	"github.com/dertseha/everoute-web/data"
)

func buildSolarSystems(builder *universe.UniverseBuilder) {
	for _, system := range data.SolarSystems {
		trueSec := universe.TrueSecurity(system.Security)
		galaxyId := universe.NewEdenId

		if system.RegionId >= 11000000 {
			galaxyId = universe.WSpaceId
		}

		builder.AddSolarSystem(system.SolarSystemId, system.ConstellationId, system.RegionId, galaxyId,
			universe.NewSpecificLocation(system.X, system.Y, system.Z), trueSec)
	}
}

func getSolarSystemIdsByName() map[string]universe.Id {
	result := make(map[string]universe.Id)

	for _, system := range data.SolarSystems {
		result[system.Name] = system.SolarSystemId
	}

	return result
}

func getJumpGateDestinationName(gate data.JumpGateData) string {
	destNameStart := strings.Index(gate.Name, "(") + 1
	destNameEnd := strings.Index(gate.Name, ")")

	return gate.Name[destNameStart:destNameEnd]
}

func getJumpGateKey(fromSolarSystemId, toSolarSystemId universe.Id) string {
	return fmt.Sprintf("%d->%d", fromSolarSystemId, toSolarSystemId)
}

func getJumpGateLocations() map[string]universe.Location {
	result := make(map[string]universe.Location)
	solarSystemIdsByName := getSolarSystemIdsByName()

	for _, gate := range data.JumpGates {
		destName := getJumpGateDestinationName(gate)
		key := getJumpGateKey(gate.SolarSystemId, solarSystemIdsByName[destName])
		location := universe.NewSpecificLocation(gate.X, gate.Y, gate.Z)

		result[key] = location
	}

	return result
}

func buildJumpGates(builder *universe.UniverseBuilder) {
	jumpGateLocations := getJumpGateLocations()

	for _, jumpData := range data.SolarSystemJumps {
		extension := builder.ExtendSolarSystem(jumpData.FromSolarSystemId)
		jumpBuilder := extension.BuildJump(jumpgate.JumpType, jumpData.ToSolarSystemId)

		jumpBuilder.From(jumpGateLocations[getJumpGateKey(jumpData.FromSolarSystemId, jumpData.ToSolarSystemId)])
		jumpBuilder.To(jumpGateLocations[getJumpGateKey(jumpData.ToSolarSystemId, jumpData.FromSolarSystemId)])
	}
}

func dropUnusedData() {
	data.SolarSystems = nil
	data.SolarSystemJumps = nil
	data.JumpGates = nil
}

func prepareUniverse() *universe.UniverseBuilder {
	builder := universe.New().Extend()

	buildSolarSystems(builder)
	buildJumpGates(builder)
	transitcount.ExtendUniverse(builder)
	security.ExtendUniverse(builder)

	jumpdrive.ExtendUniverse(builder, 5.0)

	dropUnusedData()

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

	log.Printf("Initializing server...")
	rpcServer := rpc.NewServer()
	rpcServer.RegisterCodec(rpcJson.NewCodec(), "application/json")
	service := NewRouteService(universe)
	rpcServer.RegisterService(service, "Route")

	http.Handle("/", rpcServer)
	serverPort := os.Getenv("PORT")
	if serverPort == "" {
		serverPort = "3000"
	}
	log.Printf("Starting server on port <%s>...", serverPort)
	http.ListenAndServe(":"+serverPort, nil)
}
