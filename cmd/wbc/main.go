package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/jaydasondee/wbc/internal/classifier"
	"github.com/jaydasondee/wbc/internal/config"
	"github.com/jaydasondee/wbc/internal/ingester"
	"github.com/jaydasondee/wbc/internal/profiler"
	"github.com/jaydasondee/wbc/internal/store"
	"github.com/jaydasondee/wbc/internal/transformer"
	"github.com/jaydasondee/wbc/pkg/ethapi"
	"github.com/jaydasondee/wbc/pkg/models"
)

var cfg *config.Config

func main() {
	root := &cobra.Command{
		Use:   "wbc",
		Short: "Wallet Behavior Classifier — DeFi on-chain analytics",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			var err error
			cfg, err = config.Load()
			return err
		},
	}

	root.AddCommand(ingestCmd(), analyzeCmd(), profileCmd())

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func ingestCmd() *cobra.Command {
	var wallet, contract, chain string
	var fromBlk, toBlk uint64

	cmd := &cobra.Command{
		Use:   "ingest",
		Short: "Fetch and store on-chain transactions",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			client := ethapi.NewEtherscan(cfg.Etherscan.APIKey, cfg.Etherscan.BaseURL, cfg.Etherscan.RPS)
			st, err := store.New(cfg.Store.Path)
			if err != nil {
				return err
			}
			defer st.Close()

			ig := ingester.New(client, st)

			if wallet != "" {
				n, err := ig.IngestWallet(ctx, wallet, fromBlk, toBlk)
				if err != nil {
					return err
				}
				fmt.Fprintf(os.Stdout, `{"wallet":"%s","txs_ingested":%d}`+"\n", wallet, n)
			}
			if contract != "" {
				n, err := ig.IngestContract(ctx, contract, fromBlk, toBlk)
				if err != nil {
					return err
				}
				fmt.Fprintf(os.Stdout, `{"contract":"%s","txs_ingested":%d}`+"\n", contract, n)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&wallet, "wallet", "", "wallet address to ingest")
	cmd.Flags().StringVar(&contract, "contract", "", "contract address to ingest")
	cmd.Flags().StringVar(&chain, "chain", "ethereum", "chain name")
	cmd.Flags().Uint64Var(&fromBlk, "from-block", 0, "start block")
	cmd.Flags().Uint64Var(&toBlk, "to-block", 99999999, "end block")
	return cmd
}

func analyzeCmd() *cobra.Command {
	var algo string
	var k, minCluster int

	cmd := &cobra.Command{
		Use:   "analyze",
		Short: "Cluster wallets by behavioral features",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			st, err := store.New(cfg.Store.Path)
			if err != nil {
				return err
			}
			defer st.Close()

			addrs, err := st.GetAllWallets(ctx)
			if err != nil {
				return err
			}

			tf := transformer.New()

			var vecs [][]float64
			addrMap := make(map[int]string)

			for i, addr := range addrs {
				txs, err := st.GetTxsByWallet(ctx, addr)
				if err != nil {
					continue
				}
				f, err := tf.Extract(ctx, addr, txs)
				if err != nil {
					continue
				}
				vecs = append(vecs, f.Vec())
				addrMap[i] = addr
			}

			var clusters []interface{}
			switch algo {
			case "kmeans":
				km := classifier.NewKMeans(k, cfg.Analysis.MaxIter)
				c, err := km.Fit(vecs)
				if err != nil {
					return err
				}
				for _, cl := range c {
					cluster := cl.(classifier.Cluster)
					for i := range cluster.Members {
						cluster.Members[i] = addrMap[cluster.Members[i]]
					}
					clusters = append(clusters, cluster)
				}
			case "hdbscan":
				hdb := classifier.NewHDBSCAN(minCluster, cfg.Analysis.HDBSCANMinPts)
				c, err := hdb.Fit(vecs)
				if err != nil {
					return err
				}
				for _, cl := range c {
					cluster := cl.(classifier.Cluster)
					for i := range cluster.Members {
						cluster.Members[i] = addrMap[cluster.Members[i]]
					}
					clusters = append(clusters, cluster)
				}
			default:
				return fmt.Errorf("unknown algorithm: %s", algo)
			}

			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(clusters)
		},
	}

	cmd.Flags().StringVar(&algo, "algorithm", "kmeans", "clustering algorithm (kmeans|hdbscan)")
	cmd.Flags().IntVar(&k, "k", 5, "number of clusters for k-means")
	cmd.Flags().IntVar(&minCluster, "min-cluster", 10, "minimum cluster size for HDBSCAN")
	return cmd
}

func profileCmd() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "profile",
		Short: "Generate behavioral profiles for classified wallets",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			st, err := store.New(cfg.Store.Path)
			if err != nil {
				return err
			}
			defer st.Close()

			addrs, err := st.GetAllWallets(ctx)
			if err != nil {
				return err
			}

			tf := transformer.New()

			var vecs [][]float64
			addrFeatureMap := make(map[string]models.Feature)

			for _, addr := range addrs {
				txs, err := st.GetTxsByWallet(ctx, addr)
				if err != nil {
					continue
				}
				f, err := tf.Extract(ctx, addr, txs)
				if err != nil {
					continue
				}
				vecs = append(vecs, f.Vec())
				addrFeatureMap[addr] = f
			}

			km := classifier.NewKMeans(cfg.Analysis.DefaultK, cfg.Analysis.MaxIter)
			clusters, err := km.Fit(vecs)
			if err != nil {
				return err
			}

			pf := profiler.New()
			profiles, err := pf.Label(clusters, addrFeatureMap)
			if err != nil {
				return err
			}

			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")

			switch format {
			case "json":
				return enc.Encode(profiles)
			case "table":
				for _, p := range profiles {
					fmt.Fprintf(os.Stdout, "Profile %s | %s | %d members\n", p.Addr, p.Label, len(p.Members))
				}
			default:
				return enc.Encode(profiles)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&format, "output", "json", "output format (json|table|csv)")
	return cmd
}
