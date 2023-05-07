package utils

import "time"

type ConstantsType struct {
	AllowedFailures          int
	AppVersion               string
	InitialConnectRetryDelay time.Duration
	PollingPeriod            int
}

var Constants = ConstantsType{
	AllowedFailures:          3,
	AppVersion:               "0.1.0",
	InitialConnectRetryDelay: time.Duration(15),
	PollingPeriod:            5,
}
