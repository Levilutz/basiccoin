package utils

import "time"

type ConstantsType struct {
	AllowedFailures          int
	AppVersion               string
	InitialConnectRetryDelay time.Duration
	PollingPeriod            time.Duration
}

var Constants = ConstantsType{
	AllowedFailures:          3,
	AppVersion:               "0.1.0",
	InitialConnectRetryDelay: time.Duration(15) * time.Second,
	PollingPeriod:            time.Duration(5) * time.Second,
}
