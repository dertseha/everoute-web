package api

import (
	"github.com/dertseha/everoute/universe"
)

type PathEntry struct {
	SolarSystem  universe.Id `json:"solarSystem"`
	JumpDistance interface{} `json:"jumpDistance,omitempty"`
	WarpDistance interface{} `json:"warpDistance,omitempty"`
}

type RouteFindResponse struct {
	Path []PathEntry `json:"path"`
}
