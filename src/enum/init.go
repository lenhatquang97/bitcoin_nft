package enum

type NetWorkValue string
type NetworkEnt struct {
	Bitcoin NetWorkValue
	Testnet NetWorkValue
	Signet  NetWorkValue
	RegTest NetWorkValue
}

var Network = &NetworkEnt{
	"bitcoin",
	"testnet",
	"signet",
	"regtest",
}

type TargetValue string
type TargetEnt struct {
	PostAge TargetValue
	Value   TargetValue
}

var Target = &TargetEnt{
	"PostAge",
	"Value",
}
