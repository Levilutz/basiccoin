package utils

import (
	"time"

	"github.com/google/uuid"
)

type ConstantsType struct {
	AllowedFailures          int
	AppVersion               string
	DesiredPeers             int
	InitialConnectRetryDelay time.Duration
	LocalAddr                string
	PollingPeriod            time.Duration
	RuntimeID                string
}

var Constants = ConstantsType{
	AllowedFailures:          3,
	AppVersion:               "0.1.0",
	DesiredPeers:             3,
	InitialConnectRetryDelay: time.Duration(15) * time.Second,
	LocalAddr:                "", // Must initialize
	PollingPeriod:            time.Duration(5) * time.Second,
	RuntimeID:                uuid.New().String(),
}
