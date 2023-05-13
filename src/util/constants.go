package util

// Types of the constants struct
type ConstantsType struct {
	RuntimeID string `json:"runtimeID"`
	Version   string `json:"version"`
}

// Program-wide constants, should be set on startup
var Constants = ConstantsType{
	RuntimeID: AssertUUID(),
	Version:   "0.1.0",
}
