package prot

type Params struct {
	RuntimeID string `json:"runtimeId"` // An id to uniquely identify this node.
}

func StdParams(runtimeId string) Params {
	return Params{
		RuntimeID: runtimeId,
	}
}
