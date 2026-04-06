// Package store provides SQLite-backed transaction storage.
package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"

	"github.com/jaydasondee/wbc/pkg/models"
)

type SQLite struct {
	db *sql.DB
}

func New(path string) (*SQLite, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("store.New: %w", err)
	}
	s := &SQLite{db: db}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("store.New: %w", err)
	}
	return s, nil
}

func (s *SQLite) migrate() error {
	q := `CREATE TABLE IF NOT EXISTS txs (
		hash TEXT PRIMARY KEY,
		from_addr TEXT NOT NULL,
		to_addr TEXT,
		value REAL,
		gas INTEGER,
		gas_price REAL,
		block_num INTEGER,
		timestamp INTEGER,
		input TEXT,
		is_error INTEGER,
		contract_addr TEXT,
		chain TEXT
	);
	CREATE INDEX IF NOT EXISTS idx_txs_from ON txs(from_addr);
	CREATE INDEX IF NOT EXISTS idx_txs_to ON txs(to_addr);
	CREATE INDEX IF NOT EXISTS idx_txs_contract ON txs(contract_addr);`
	_, err := s.db.Exec(q)
	return err
}

func (s *SQLite) SaveTxs(ctx context.Context, txs []models.Tx) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("store.SaveTxs: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx,
		`INSERT OR IGNORE INTO txs (hash,from_addr,to_addr,value,gas,gas_price,block_num,timestamp,input,is_error,contract_addr,chain)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`)
	if err != nil {
		return fmt.Errorf("store.SaveTxs: %w", err)
	}
	defer stmt.Close()

	for _, t := range txs {
		isErr := 0
		if t.IsError {
			isErr = 1
		}
		_, err := stmt.ExecContext(ctx, t.Hash, t.From, t.To, t.Value, t.Gas, t.GasPrice,
			t.BlockNum, t.Timestamp.Unix(), t.Input, isErr, t.ContractAddr, t.Chain)
		if err != nil {
			return fmt.Errorf("store.SaveTxs: %w", err)
		}
	}
	return tx.Commit()
}

func (s *SQLite) GetTxsByWallet(ctx context.Context, addr string) ([]models.Tx, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT hash,from_addr,to_addr,value,gas,gas_price,block_num,timestamp,input,is_error,contract_addr,chain
		FROM txs WHERE from_addr=? OR to_addr=? ORDER BY block_num`, addr, addr)
	if err != nil {
		return nil, fmt.Errorf("store.GetTxsByWallet: %w", err)
	}
	defer rows.Close()
	return scanTxs(rows)
}

func (s *SQLite) GetTxsByContract(ctx context.Context, addr string) ([]models.Tx, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT hash,from_addr,to_addr,value,gas,gas_price,block_num,timestamp,input,is_error,contract_addr,chain
		FROM txs WHERE to_addr=? OR contract_addr=? ORDER BY block_num`, addr, addr)
	if err != nil {
		return nil, fmt.Errorf("store.GetTxsByContract: %w", err)
	}
	defer rows.Close()
	return scanTxs(rows)
}

func (s *SQLite) GetAllWallets(ctx context.Context) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT DISTINCT from_addr FROM txs`)
	if err != nil {
		return nil, fmt.Errorf("store.GetAllWallets: %w", err)
	}
	defer rows.Close()

	var addrs []string
	for rows.Next() {
		var a string
		if err := rows.Scan(&a); err != nil {
			return nil, fmt.Errorf("store.GetAllWallets: %w", err)
		}
		addrs = append(addrs, a)
	}
	return addrs, rows.Err()
}

func (s *SQLite) HasWallet(ctx context.Context, addr string) (bool, error) {
	var n int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM txs WHERE from_addr=? LIMIT 1`, addr).Scan(&n)
	if err != nil {
		return false, fmt.Errorf("store.HasWallet: %w", err)
	}
	return n > 0, nil
}

func (s *SQLite) Close() error {
	return s.db.Close()
}

func scanTxs(rows *sql.Rows) ([]models.Tx, error) {
	var txs []models.Tx
	for rows.Next() {
		var t models.Tx
		var ts int64
		var isErr int
		err := rows.Scan(&t.Hash, &t.From, &t.To, &t.Value, &t.Gas, &t.GasPrice,
			&t.BlockNum, &ts, &t.Input, &isErr, &t.ContractAddr, &t.Chain)
		if err != nil {
			return nil, fmt.Errorf("store.scanTxs: %w", err)
		}
		t.Timestamp = timeFromUnix(ts)
		t.IsError = isErr == 1
		txs = append(txs, t)
	}
	return txs, rows.Err()
}

func timeFromUnix(ts int64) time.Time {
	return time.Unix(ts, 0)
}
