// Package models defines shared types for the wallet behavior classifier.
package models

import "time"

type Tx struct {
	Hash        string
	From        string
	To          string
	Value       float64
	Gas         uint64
	GasPrice    float64
	BlockNum    uint64
	Timestamp   time.Time
	Input       string
	IsError     bool
	ContractAddr string
	Chain       string
}

type Wallet struct {
	Addr    string
	Txs     []Tx
	Profile *Profile
}

type Feature struct {
	TxFreq           float64
	ProtocolDiv      float64
	AvgGasPremium    float64
	TimingEntropy    float64
	ValueMean        float64
	ValueStd         float64
	ValueSkew        float64
	SeqPattern       float64
	FirstInteractLag float64
	CopyScore        float64
}

const (
	FTxFreq = iota
	FProtocolDiv
	FAvgGasPremium
	FTimingEntropy
	FValueMean
	FValueStd
	FValueSkew
	FSeqPattern
	FFirstInteractLag
	FCopyScore
	NumFeatures
)

func (f Feature) Vec() []float64 {
	return []float64{
		f.TxFreq, f.ProtocolDiv, f.AvgGasPremium, f.TimingEntropy,
		f.ValueMean, f.ValueStd, f.ValueSkew,
		f.SeqPattern, f.FirstInteractLag, f.CopyScore,
	}
}

type Label string

const (
	LabelBot           Label = "bot"
	LabelWhale         Label = "whale"
	LabelRetail        Label = "retail_explorer"
	LabelStrategyCopy  Label = "strategy_copier"
	LabelAirdropFarmer Label = "airdrop_farmer"
	LabelUnknown       Label = "unknown"
)

type Profile struct {
	Addr       string
	Label      Label
	Confidence float64
	Features   Feature
	ClusterID  int
}

type Cluster struct {
	ID       int
	Centroid []float64
	Members  []string
	Label    Label
}

type IngestOpts struct {
	Wallet   string
	Contract string
	Protocol string
	Chain    string
	FromBlk  uint64
	ToBlk    uint64
}

type AnalyzeOpts struct {
	Algorithm  string
	K          int
	MinCluster int
}

type ProfileOpts struct {
	Format string
}
