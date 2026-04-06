package classifier

import (
	"fmt"
	"math"
	"sort"

	"github.com/jaydasondee/wbc/pkg/models"
)

type HDBSCAN struct {
	MinCluster int
	MinPts     int
}

func NewHDBSCAN(minCluster, minPts int) *HDBSCAN {
	return &HDBSCAN{MinCluster: minCluster, MinPts: minPts}
}

type edge struct {
	i, j int
	w    float64
}

func (h *HDBSCAN) Fit(vecs [][]float64) ([]models.Cluster, error) {
	n := len(vecs)
	if n == 0 {
		return nil, fmt.Errorf("classifier.HDBSCAN.Fit: empty input")
	}

	coreDists := h.coreDistances(vecs)
	mrd := h.mutualReachability(vecs, coreDists)
	mst := h.primMST(mrd, n)
	sort.Slice(mst, func(i, j int) bool { return mst[i].w < mst[j].w })

	labels := h.extractClusters(mst, n)

	clMap := make(map[int][]string)
	for i, l := range labels {
		clMap[l] = append(clMap[l], fmt.Sprintf("%d", i))
	}

	var clusters []models.Cluster
	id := 0
	for _, members := range clMap {
		if len(members) < h.MinCluster {
			continue
		}
		c := models.Cluster{ID: id, Members: members}
		c.Centroid = centroidOf(vecs, members)
		clusters = append(clusters, c)
		id++
	}
	return clusters, nil
}

func (h *HDBSCAN) Predict(_ []float64) (int, error) {
	return -1, fmt.Errorf("classifier.HDBSCAN.Predict: not supported for HDBSCAN")
}

func (h *HDBSCAN) coreDistances(vecs [][]float64) []float64 {
	n := len(vecs)
	core := make([]float64, n)
	k := h.MinPts
	if k <= 0 {
		k = 5
	}

	for i := 0; i < n; i++ {
		dists := make([]float64, n)
		for j := 0; j < n; j++ {
			dists[j] = euclidean(vecs[i], vecs[j])
		}
		sort.Float64s(dists)
		idx := k
		if idx >= n {
			idx = n - 1
		}
		core[i] = dists[idx]
	}
	return core
}

func (h *HDBSCAN) mutualReachability(vecs [][]float64, core []float64) [][]float64 {
	n := len(vecs)
	mrd := make([][]float64, n)
	for i := range mrd {
		mrd[i] = make([]float64, n)
		for j := range mrd[i] {
			d := euclidean(vecs[i], vecs[j])
			mrd[i][j] = math.Max(core[i], math.Max(core[j], d))
		}
	}
	return mrd
}

func (h *HDBSCAN) primMST(dist [][]float64, n int) []edge {
	inMST := make([]bool, n)
	key := make([]float64, n)
	parent := make([]int, n)
	for i := range key {
		key[i] = math.MaxFloat64
		parent[i] = -1
	}
	key[0] = 0

	var edges []edge
	for count := 0; count < n; count++ {
		u := -1
		for i := 0; i < n; i++ {
			if !inMST[i] && (u == -1 || key[i] < key[u]) {
				u = i
			}
		}
		inMST[u] = true
		if parent[u] >= 0 {
			edges = append(edges, edge{parent[u], u, key[u]})
		}
		for v := 0; v < n; v++ {
			if !inMST[v] && dist[u][v] < key[v] {
				key[v] = dist[u][v]
				parent[v] = u
			}
		}
	}
	return edges
}

func (h *HDBSCAN) extractClusters(mst []edge, n int) []int {
	uf := newUnionFind(n)
	labels := make([]int, n)
	for i := range labels {
		labels[i] = i
	}

	for _, e := range mst {
		ri := uf.find(e.i)
		rj := uf.find(e.j)
		si := uf.size[ri]
		sj := uf.size[rj]
		if si >= h.MinCluster && sj >= h.MinCluster {
			uf.union(e.i, e.j)
		} else {
			uf.union(e.i, e.j)
		}
	}

	for i := range labels {
		labels[i] = uf.find(i)
	}
	return labels
}

type unionFind struct {
	parent []int
	size   []int
}

func newUnionFind(n int) *unionFind {
	p := make([]int, n)
	s := make([]int, n)
	for i := range p {
		p[i] = i
		s[i] = 1
	}
	return &unionFind{p, s}
}

func (uf *unionFind) find(x int) int {
	for uf.parent[x] != x {
		uf.parent[x] = uf.parent[uf.parent[x]]
		x = uf.parent[x]
	}
	return x
}

func (uf *unionFind) union(x, y int) {
	rx, ry := uf.find(x), uf.find(y)
	if rx == ry {
		return
	}
	if uf.size[rx] < uf.size[ry] {
		rx, ry = ry, rx
	}
	uf.parent[ry] = rx
	uf.size[rx] += uf.size[ry]
}

func euclidean(a, b []float64) float64 {
	return math.Sqrt(euclideanSq(a, b))
}

func centroidOf(vecs [][]float64, members []string) []float64 {
	if len(members) == 0 {
		return nil
	}
	dim := len(vecs[0])
	c := make([]float64, dim)
	for _, mStr := range members {
		var idx int
		fmt.Sscanf(mStr, "%d", &idx)
		if idx < len(vecs) {
			for j := range c {
				c[j] += vecs[idx][j]
			}
		}
	}
	for j := range c {
		c[j] /= float64(len(members))
	}
	return c
}
