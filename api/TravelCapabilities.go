package api

type JumpGateTravelCapability struct {
	AvoidHighSec bool `json:"avoidHighSec"`
}

type JumpDriveTravelCapability struct {
	DistanceLimit float64 `json:"distanceLimit"`
}

type TravelCapabilities struct {
	JumpGate  *JumpGateTravelCapability  `json:"jumpGate"`
	JumpDrive *JumpDriveTravelCapability `json:"jumpDrive"`
}
