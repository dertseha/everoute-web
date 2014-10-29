package api

type TravelRuleParameter struct {
	Priority uint `json:"priority"`
}

type TransitCountTravelRuleParameter struct {
	TravelRuleParameter
}

type MinSecurityTravelRuleParameter struct {
	TravelRuleParameter
	Limit float64 `json:"limit"`
}

type MaxSecurityTravelRuleParameter struct {
	TravelRuleParameter
	Limit float64 `json:"limit"`
}

type JumpDistanceTravelRuleParameter struct {
	TravelRuleParameter
}

type WarpDistanceTravelRuleParameter struct {
	TravelRuleParameter
}

type TravelRuleset struct {
	TransitCount *TransitCountTravelRuleParameter `json:"transitCount,omitempty"`
	MinSecurity  *MinSecurityTravelRuleParameter  `json:"minSecurity,omitempty"`
	MaxSecurity  *MaxSecurityTravelRuleParameter  `json:"maxSecurity,omitempty"`
	JumpDistance *JumpDistanceTravelRuleParameter `json:"jumpDistance,omitempty"`
	WarpDistance *WarpDistanceTravelRuleParameter `json:"warpDistance,omitempty"`
}
