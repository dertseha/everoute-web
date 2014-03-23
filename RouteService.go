package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"sort"

	"github.com/dertseha/everoute/travel"
	"github.com/dertseha/everoute/travel/capabilities"
	"github.com/dertseha/everoute/travel/capabilities/jumpdrive"
	"github.com/dertseha/everoute/travel/capabilities/jumpgate"
	"github.com/dertseha/everoute/travel/rules"
	"github.com/dertseha/everoute/travel/rules/jumpdistance"
	"github.com/dertseha/everoute/travel/rules/security"
	"github.com/dertseha/everoute/travel/rules/transitcount"
	"github.com/dertseha/everoute/travel/search"
	"github.com/dertseha/everoute/universe"

	"github.com/dertseha/everoute-web/api"
)

type routeSearchResultCollector struct {
	route *search.Route
}

func (collector *routeSearchResultCollector) Collect(route *search.Route) {
	collector.route = route
}

type RouteService struct {
	universe universe.Universe
}

func NewRouteService(universe universe.Universe) *RouteService {
	service := &RouteService{
		universe: universe}

	return service
}

func (service *RouteService) Find(r *http.Request, request *api.RouteFindRequest, response *api.RouteFindResponse) (err error) {
	defer func() {
		if panic := recover(); panic != nil {
			errorText := fmt.Sprintf("Failed to find route: \"%s\"", panic)
			log.Print(errorText)
			err = errors.New(errorText)
		}
	}()

	capability := getTravelCapability(service.universe, &request.Capabilities)
	rule := getTravelRule(request.Rules)
	starts := getStartSystems(service.universe, &request.Route.From)
	var collector = &routeSearchResultCollector{}
	var searchDone = make(chan int)

	builder := search.NewRouteFinder(capability, rule, starts, collector, func() { searchDone <- 1 })
	for _, waypoint := range request.Route.Via {
		builder.AddWaypoint(getOptimizedSystemSearchCriterion(service.universe, waypoint.SolarSystem, rule, request.Route.Avoid))
	}
	if request.Route.To != nil {
		builder.ForDestination(getOptimizedSystemSearchCriterion(service.universe, request.Route.To.SolarSystem, rule, request.Route.Avoid))
	}

	builder.Build()

	<-searchDone

	response.Path = make([]api.PathEntry, 0)
	if collector.route != nil {
		steps := collector.route.Steps()
		for _, step := range steps {
			pathEntry := api.PathEntry{SolarSystem: step.SolarSystemId()}
			response.Path = append(response.Path, pathEntry)
		}
	}

	return
}

type priorizedTravelRule struct {
	priority uint
	rule     travel.TravelRule
}

type priorizedTravelRules []*priorizedTravelRule

func (rules priorizedTravelRules) Len() int {
	return len(rules)
}

func (rules priorizedTravelRules) Swap(i, j int) {
	rules[i], rules[j] = rules[j], rules[i]
}

func (rules priorizedTravelRules) Less(i, j int) bool {
	return rules[i].priority < rules[j].priority
}

func getTravelRule(ruleset *api.TravelRuleset) travel.TravelRule {
	list := make([]travel.TravelRule, 0)
	hasTransitCount := false
	priorizedRules := make(priorizedTravelRules, 0)

	addRule := func(priority uint, rule travel.TravelRule) {
		entry := &priorizedTravelRule{priority: priority, rule: rule}
		priorizedRules = append(priorizedRules, entry)
	}

	if ruleset != nil {
		if ruleset.TransitCount != nil {
			addRule(ruleset.TransitCount.Priority, transitcount.Rule())
			hasTransitCount = true
		}
		if ruleset.MinSecurity != nil {
			addRule(ruleset.MinSecurity.Priority, security.MinRule(ruleset.MinSecurity.Limit))
		}
		if ruleset.MaxSecurity != nil {
			addRule(ruleset.MaxSecurity.Priority, security.MaxRule(ruleset.MaxSecurity.Limit))
		}
		if ruleset.JumpDistance != nil {
			addRule(ruleset.JumpDistance.Priority, jumpdistance.Rule())
		}
	}
	sort.Sort(priorizedRules)
	for _, entry := range priorizedRules {
		list = append(list, entry.rule)
	}
	if !hasTransitCount {
		list = append(list, transitcount.Rule())
	}

	return rules.TravelRuleset(list...)
}

func getTravelCapability(universe universe.Universe, requestedCapabilities *api.TravelCapabilities) travel.TravelCapability {
	list := make([]travel.TravelCapability, 0)

	if requestedCapabilities.JumpGate != nil {
		list = append(list, jumpgate.JumpGateTravelCapability(universe))
	}
	if requestedCapabilities.JumpDrive != nil {
		list = append(list, jumpdrive.JumpDriveTravelCapability(universe, requestedCapabilities.JumpDrive.DistanceLimit))
	}

	return capabilities.CombiningTravelCapability(list...)
}

func getOptimizedSystemSearchCriterion(universe universe.Universe, solarSystemId universe.Id, rule travel.TravelRule, avoid *api.AvoidEntry) search.SearchCriterion {
	criteria := make([]search.SearchCriterion, 0)

	criteria = append(criteria, search.DestinationSystemSearchCriterion(universe.SolarSystem(solarSystemId).Id()))
	criteria = append(criteria, search.CostAwareSearchCriterion(rule))
	if avoid != nil {
		criteria = append(criteria, search.SystemAvoidingSearchCriterion(avoid.SolarSystems...))
	}

	return search.CombiningSearchCriterion(criteria...)
}

func getStartSystems(universe universe.Universe, from *api.FromEntry) []travel.Path {
	starts := make([]travel.Path, 0)

	for _, solarSystemId := range from.SolarSystems {
		solarSystem := universe.SolarSystem(solarSystemId)
		path := travel.NewPath(travel.NewStepBuilder(solarSystem.Id()).Build())
		starts = append(starts, path)
	}

	return starts
}
