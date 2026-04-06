// Package transformer extracts behavioral features from raw transactions.
package transformer

import (
	"context"
	"math"
	"sort"

	"github.com/jaydasondee/wbc/pkg/models"
)

type Transformer struct {
	whales map[string][]string
}

func New() *Transformer { return &Transformer{whales: make(map[string][]string)} }

func (t *Transformer) SetWhaleData(d map[string][]string) { t.whales = d }

func (t *Transformer) Extract(_ context.Context, _ string, txs []models.Tx) (models.Feature, error) {
	if len(txs) == 0 {
		return models.Feature{}, nil
	}

	var f models.Feature
	f.TxFreq = txFrequency(txs)
	f.ProtocolDiv = protocolDiversity(txs)
	f.AvgGasPremium = avgGasPremium(txs)
	f.TimingEntropy = timingEntropy(txs)
	f.ValueMean, f.ValueStd, f.ValueSkew = valueDistribution(txs)
	f.SeqPattern = sequencePattern(txs)
	f.FirstInteractLag = firstInteractionLag(txs)
	f.CopyScore = 0
	var seq []string
	for _, tx := range txs {
		if tx.To != "" {
			seq = append(seq, tx.To)
		}
	}
	if len(seq) > 0 && len(t.whales) > 0 {
		for _, ws := range t.whales {
			d := len(seq)
			if len(ws) > d {
				d = len(ws)
			}
			r := float64(lcsLen(seq, ws)) / float64(d)
			if r > f.CopyScore {
				f.CopyScore = r
			}
		}
	}
	return f, nil
}

func (t *Transformer) ExtractBatch(ctx context.Context, wallets map[string][]models.Tx) (map[string]models.Feature, error) {
	out := make(map[string]models.Feature, len(wallets))
	for addr, txs := range wallets {
		f, err := t.Extract(ctx, addr, txs)
		if err != nil {
			return nil, err
		}
		out[addr] = f
	}
	return out, nil
}

func txFrequency(txs []models.Tx) float64 {
	if len(txs) < 2 {
		return float64(len(txs))
	}
	sort.Slice(txs, func(i, j int) bool { return txs[i].Timestamp.Before(txs[j].Timestamp) })
	days := txs[len(txs)-1].Timestamp.Sub(txs[0].Timestamp).Hours() / 24
	if days < 1 {
		days = 1
	}
	return float64(len(txs)) / days
}

func protocolDiversity(txs []models.Tx) float64 {
	seen := make(map[string]struct{})
	for _, tx := range txs {
		if tx.To != "" {
			seen[tx.To] = struct{}{}
		}
	}
	return float64(len(seen))
}

func avgGasPremium(txs []models.Tx) float64 {
	if len(txs) == 0 {
		return 0
	}
	var sum float64
	for _, tx := range txs {
		sum += tx.GasPrice
	}
	return sum / float64(len(txs))
}

func timingEntropy(txs []models.Tx) float64 {
	bins := make([]float64, 24)
	for _, tx := range txs {
		bins[tx.Timestamp.Hour()]++
	}
	n := float64(len(txs))
	var h float64
	for _, c := range bins {
		if c > 0 {
			p := c / n
			h -= p * math.Log2(p)
		}
	}
	return h
}

func valueDistribution(txs []models.Tx) (mean, std, skew float64) {
	n := float64(len(txs))
	if n == 0 {
		return
	}
	var sum float64
	for _, tx := range txs {
		sum += tx.Value
	}
	mean = sum / n

	var v float64
	for _, tx := range txs {
		d := tx.Value - mean
		v += d * d
	}
	std = math.Sqrt(v / n)

	if std == 0 {
		return
	}
	var s float64
	for _, tx := range txs {
		d := (tx.Value - mean) / std
		s += d * d * d
	}
	skew = s / n
	return
}

func sequencePattern(txs []models.Tx) float64 {
	if len(txs) < 4 {
		return 0
	}
	pairs := make(map[string]int)
	for i := 0; i < len(txs)-1; i++ {
		k := txs[i].To + ">" + txs[i+1].To
		pairs[k]++
	}
	maxR := 0
	for _, c := range pairs {
		if c > maxR {
			maxR = c
		}
	}
	return float64(maxR) / float64(len(txs)-1)
}

func firstInteractionLag(txs []models.Tx) float64 {
	if len(txs) == 0 {
		return 0
	}
	sort.Slice(txs, func(i, j int) bool { return txs[i].Timestamp.Before(txs[j].Timestamp) })
	return float64(txs[0].BlockNum)
}

func lcsLen(a, b []string) int {
	n, m := len(a), len(b)
	dp := make([][]int, n+1)
	for i := range dp {
		dp[i] = make([]int, m+1)
	}
	for i := 1; i <= n; i++ {
		for j := 1; j <= m; j++ {
			if a[i-1] == b[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else {
				dp[i][j] = dp[i-1][j]
				if dp[i][j-1] > dp[i][j] {
					dp[i][j] = dp[i][j-1]
				}
			}
		}
	}
	return dp[n][m]
}
