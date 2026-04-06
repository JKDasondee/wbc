// Package config handles configuration loading from file, env, and flags.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	Etherscan EtherscanConfig `mapstructure:"etherscan"`
	Alchemy   AlchemyConfig   `mapstructure:"alchemy"`
	Store     StoreConfig     `mapstructure:"store"`
	Analysis  AnalysisConfig  `mapstructure:"analysis"`
}

type EtherscanConfig struct {
	APIKey  string `mapstructure:"api_key"`
	BaseURL string `mapstructure:"base_url"`
	RPS     int    `mapstructure:"rps"`
}

type AlchemyConfig struct {
	APIKey  string `mapstructure:"api_key"`
	BaseURL string `mapstructure:"base_url"`
}

type StoreConfig struct {
	Path string `mapstructure:"path"`
}

type AnalysisConfig struct {
	DefaultK      int `mapstructure:"default_k"`
	MaxIter       int `mapstructure:"max_iter"`
	MinCluster    int `mapstructure:"min_cluster"`
	HDBSCANMinPts int `mapstructure:"hdbscan_min_pts"`
}

func Load() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("config.Load: %w", err)
	}

	cfgDir := filepath.Join(home, ".wbc")
	os.MkdirAll(cfgDir, 0755)

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(cfgDir)
	viper.AddConfigPath(".")

	viper.SetDefault("etherscan.base_url", "https://api.etherscan.io/api")
	viper.SetDefault("etherscan.rps", 5)
	viper.SetDefault("store.path", filepath.Join(cfgDir, "wbc.db"))
	viper.SetDefault("analysis.default_k", 5)
	viper.SetDefault("analysis.max_iter", 100)
	viper.SetDefault("analysis.min_cluster", 10)
	viper.SetDefault("analysis.hdbscan_min_pts", 5)

	viper.SetEnvPrefix("WBC")
	viper.AutomaticEnv()

	viper.ReadInConfig()

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("config.Load: %w", err)
	}
	return &cfg, nil
}
