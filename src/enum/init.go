package enum

type ChainValue string
type ChainEnt struct {
	Bitcoin ChainValue
	Testnet ChainValue
	Signet  ChainValue
	RegTest ChainValue
}

var Chain = &ChainEnt{
	"bitcoin",
	"testnet3",
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

type StatisticValue int64
type StatisticEnt struct {
	Schema           StatisticValue
	Commits          StatisticValue
	LostSats         StatisticValue
	OutputsTraversed StatisticValue
	SatRanges        StatisticValue
}

var Statistic = StatisticEnt{
	0,
	1,
	2,
	3,
	4,
}
