package util

// Types of the constants struct
type ConstantsType struct {
	LocalAddr string `json:"localAddr"`
	RuntimeID string `json:"runtimeID"`
	Version   string `json:"version"`
}

// Program-wide constants, should be set on startup
var Constants = ConstantsType{
	LocalAddr: "localhost:21720",
	RuntimeID: AssertUUID(),
	Version:   "0.1.0",
}