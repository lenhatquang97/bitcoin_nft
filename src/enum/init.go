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
