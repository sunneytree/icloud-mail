package store

import (
	"context"
	"database/sql"
	"fmt"
)

// migrateAccountForwardingPolicy keeps legacy accounts on the default-off policy.
func (s *Store) migrateAccountForwardingPolicy(ctx context.Context, tx *sql.Tx) error {
	_, err := s.txExecContext(ctx, tx, `CREATE TABLE IF NOT EXISTS account_forwarding_policy (
		account_id BIGINT PRIMARY KEY REFERENCES accounts(id) ON DELETE CASCADE,
		verify_forwarding_target BOOLEAN NOT NULL DEFAULT FALSE
	)`)
	if err != nil {
		return fmt.Errorf("migrate account forwarding policy: %w", err)
	}
	return nil
}

func (s *Store) SetAccountForwardingPolicy(ctx context.Context, accountID int64, verify bool) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin forwarding policy update: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	var id int64
	if err := s.txQueryRowContext(ctx, tx, `SELECT id FROM accounts WHERE id = ?`, accountID).Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			return ErrNotFound
		}
		return fmt.Errorf("find account for forwarding policy: %w", err)
	}
	if _, err := s.txExecContext(ctx, tx, `INSERT INTO account_forwarding_policy(account_id, verify_forwarding_target)
		VALUES(?, ?) ON CONFLICT(account_id) DO UPDATE SET verify_forwarding_target = excluded.verify_forwarding_target`, accountID, verify); err != nil {
		return fmt.Errorf("set account forwarding policy: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit forwarding policy update: %w", err)
	}
	return nil
}
