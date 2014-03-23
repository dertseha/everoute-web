package api

import (
	"github.com/dertseha/everoute/universe"
)

type FromEntry struct {
	SolarSystems SolarSystemIdList `json:"solarSystems"`
}

type TravelEntry struct {
	SolarSystem universe.Id `json:"solarSystem"`
}

type AvoidEntry struct {
	SolarSystems SolarSystemIdList `json:"solarSystems"`
}

type RouteEntry struct {
	From  FromEntry     `json:"from"`
	Via   []TravelEntry `json:"via,omitempty"`
	To    *TravelEntry  `json:"to,omitempty"`
	Avoid *AvoidEntry   `json:"avoid,omitempty"`
}

type RouteFindRequest struct {
	Route        RouteEntry         `json:"route"`
	Capabilities TravelCapabilities `json:"capabilities"`
	Rules        *TravelRuleset     `json:"rules,omitempty"`
}
