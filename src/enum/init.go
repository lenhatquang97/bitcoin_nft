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

type WalletActionValue string
type WalletAction struct {
	Balance      WalletActionValue
	Create       WalletActionValue
	Inscribe     WalletActionValue
	Inscriptions WalletActionValue
	Receive      WalletActionValue
	Restore      WalletActionValue
	Sats         WalletActionValue
	Send         WalletActionValue
	Transactions WalletActionValue
	Outputs      WalletActionValue
}

type OutGoingTypeValue string
type OutGoingTypeEnt struct {
	Amount        OutGoingTypeValue
	InscriptionId OutGoingTypeValue
	Satpoint      OutGoingTypeValue
}

var OutGoingType = &OutGoingTypeEnt{
	"amount",
	"inscriptionId",
	"satpoint",
}

func Run(action WalletActionValue, data interface{}) {

}
