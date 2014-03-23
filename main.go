package main

import (
	"log"
	"net/http"
	"os"
	"runtime"
	"runtime/debug"

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
	data.SolarSystems = nil
}

func buildJumpGates(builder *universe.UniverseBuilder) {
	for _, jumpData := range data.SolarSystemJumps {
		extension := builder.ExtendSolarSystem(jumpData.FromSolarSystemId)
		extension.AddJump(jumpgate.JumpType, jumpData.ToSolarSystemId)
	}
	data.SolarSystemJumps = nil
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
