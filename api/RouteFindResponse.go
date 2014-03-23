package api

import (
	"github.com/dertseha/everoute/universe"
)

type PathEntry struct {
	SolarSystem universe.Id `json:"solarSystem"`
}

type RouteFindResponse struct {
	Path []PathEntry `json:"path"`
}
