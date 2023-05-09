package utils

import "time"

type ConstantsType struct {
	AllowedFailures          int
	AppVersion               string
	DesiredPeers             int
	InitialConnectRetryDelay time.Duration
	LocalAddr                string
	PollingPeriod            time.Duration
}

var Constants = ConstantsType{
	AllowedFailures:          3,
	AppVersion:               "0.1.0",
	DesiredPeers:             4,
	InitialConnectRetryDelay: time.Duration(15) * time.Second,
	LocalAddr:                "", // Must initialize
	PollingPeriod:            time.Duration(5) * time.Second,
}
