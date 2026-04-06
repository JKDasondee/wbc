// Package profiler assigns behavioral labels to wallet clusters.
package profiler

import (
	"github.com/jaydasondee/wbc/pkg/models"
)

type Profiler struct{}

func New() *Profiler { return &Profiler{} }

func (p *Profiler) Label(clusters []models.Cluster, features map[string]models.Feature) ([]models.Profile, error) {
	var profiles []models.Profile

	for _, c := range clusters {
		lbl := classifyCentroid(c.Centroid)
		c.Label = lbl

		for _, addr := range c.Members {
			f, ok := features[addr]
			if !ok {
				continue
			}
			profiles = append(profiles, models.Profile{
				Addr:       addr,
				Label:      lbl,
				Confidence: confidence(f, c.Centroid),
				Features:   f,
				ClusterID:  c.ID,
			})
		}
	}
	return profiles, nil
}

func classifyCentroid(c []float64) models.Label {
	if len(c) < int(models.NumFeatures) {
		return models.LabelUnknown
	}

	freq := c[models.FTxFreq]
	div := c[models.FProtocolDiv]
	gas := c[models.FAvgGasPremium]
	entropy := c[models.FTimingEntropy]
	copyS := c[models.FCopyScore]
	val := c[models.FValueMean]

	if freq > 50 && div < 5 && gas < 20 {
		return models.LabelBot
	}
	if val > 10 && freq < 5 {
		return models.LabelWhale
	}
	if div > 20 && entropy > 3.0 {
		return models.LabelRetail
	}
	if copyS > 0.7 {
		return models.LabelStrategyCopy
	}
	if freq > 30 && c[models.FFirstInteractLag] < 100 {
		return models.LabelAirdropFarmer
	}
	return models.LabelUnknown
}

func confidence(f models.Feature, centroid []float64) float64 {
	vec := f.Vec()
	if len(vec) != len(centroid) {
		return 0
	}
	var sumSq float64
	for i := range vec {
		d := vec[i] - centroid[i]
		sumSq += d * d
	}
	return 1.0 / (1.0 + sumSq)
}
