// Package classifier implements clustering algorithms for wallet behavior grouping.
package classifier

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/jaydasondee/wbc/pkg/models"
)

type KMeans struct {
	K         int
	MaxIter   int
	centroids [][]float64
}

func NewKMeans(k, maxIter int) *KMeans {
	return &KMeans{K: k, MaxIter: maxIter}
}

func (km *KMeans) Fit(vecs [][]float64) ([]models.Cluster, error) {
	if len(vecs) == 0 {
		return nil, fmt.Errorf("classifier.KMeans.Fit: empty input")
	}
	if km.K > len(vecs) {
		km.K = len(vecs)
	}

	dim := len(vecs[0])
	km.centroids = initCentroids(vecs, km.K)

	assign := make([]int, len(vecs))
	for iter := 0; iter < km.MaxIter; iter++ {
		changed := false
		for i, v := range vecs {
			nearest := nearestCentroid(v, km.centroids)
			if nearest != assign[i] {
				assign[i] = nearest
				changed = true
			}
		}
		if !changed {
			break
		}
		km.centroids = recalcCentroids(vecs, assign, km.K, dim)
	}

	clusters := make([]models.Cluster, km.K)
	for i := range clusters {
		clusters[i] = models.Cluster{ID: i, Centroid: km.centroids[i]}
	}
	for i, c := range assign {
		clusters[c].Members = append(clusters[c].Members, fmt.Sprintf("%d", i))
	}
	return clusters, nil
}

func (km *KMeans) Predict(v []float64) (int, error) {
	if len(km.centroids) == 0 {
		return 0, fmt.Errorf("classifier.KMeans.Predict: not fitted")
	}
	return nearestCentroid(v, km.centroids), nil
}

func ElbowSSE(vecs [][]float64, maxK int) []float64 {
	sses := make([]float64, maxK)
	for k := 1; k <= maxK; k++ {
		km := NewKMeans(k, 100)
		clusters, err := km.Fit(vecs)
		if err != nil {
			continue
		}
		var sse float64
		for _, c := range clusters {
			for _, mIdx := range c.Members {
				var idx int
				fmt.Sscanf(mIdx, "%d", &idx)
				if idx < len(vecs) {
					sse += euclideanSq(vecs[idx], c.Centroid)
				}
			}
		}
		sses[k-1] = sse
	}
	return sses
}

func initCentroids(vecs [][]float64, k int) [][]float64 {
	perm := rand.Perm(len(vecs))
	centroids := make([][]float64, k)
	for i := 0; i < k; i++ {
		centroids[i] = make([]float64, len(vecs[0]))
		copy(centroids[i], vecs[perm[i]])
	}
	return centroids
}

func nearestCentroid(v []float64, centroids [][]float64) int {
	best := 0
	bestD := math.MaxFloat64
	for i, c := range centroids {
		d := euclideanSq(v, c)
		if d < bestD {
			bestD = d
			best = i
		}
	}
	return best
}

func recalcCentroids(vecs [][]float64, assign []int, k, dim int) [][]float64 {
	sums := make([][]float64, k)
	counts := make([]int, k)
	for i := range sums {
		sums[i] = make([]float64, dim)
	}
	for i, v := range vecs {
		c := assign[i]
		counts[c]++
		for j, val := range v {
			sums[c][j] += val
		}
	}
	for i := range sums {
		if counts[i] == 0 {
			continue
		}
		for j := range sums[i] {
			sums[i][j] /= float64(counts[i])
		}
	}
	return sums
}

func euclideanSq(a, b []float64) float64 {
	var s float64
	for i := range a {
		d := a[i] - b[i]
		s += d * d
	}
	return s
}
