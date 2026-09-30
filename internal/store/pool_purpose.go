package store

import (
	"context"
	"database/sql"
	"icloud-api/internal/domain"
	"strings"
)

func normalizePoolPurpose(value string) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		value = "common"
	}
	if len(value) > 64 {
		return "", ErrPoolInput
	}
	for _, r := range value {
		if !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '-' || r == '_') {
			return "", ErrPoolInput
		}
	}
	return value, nil
}

// Column presence is the transactionally committed, one-time migration marker.
// Legacy used allocations belong to Adobe; new omitted purposes belong to common.
func (s *Store) migratePoolPurpose(ctx context.Context, tx *sql.Tx) error {
	found := false
	if s.dialect == dialectPostgres {
		if err := s.txQueryRowContext(ctx, tx, `SELECT EXISTS(SELECT 1 FROM information_schema.columns WHERE table_schema=current_schema() AND table_name='pool_leases' AND column_name='purpose')`).Scan(&found); err != nil {
			return err
		}
	} else {
		rows, err := s.txQueryContext(ctx, tx, `PRAGMA table_info(pool_leases)`)
		if err != nil {
			return err
		}
		for rows.Next() {
			var cid, nn, pk int
			var name, typ string
			var def sql.NullString
			if err := rows.Scan(&cid, &name, &typ, &nn, &def, &pk); err != nil {
				rows.Close()
				return err
			}
			if name == "purpose" {
				found = true
			}
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return err
		}
	}
	if _, err := s.txExecContext(ctx, tx, `CREATE TABLE IF NOT EXISTS pool_legacy_purpose_marks(alias_id BIGINT NOT NULL REFERENCES aliases(id) ON DELETE CASCADE,purpose TEXT NOT NULL,PRIMARY KEY(alias_id,purpose))`); err != nil {
		return err
	}
	if !found {
		for _, q := range []string{
			`ALTER TABLE pool_leases ADD COLUMN purpose TEXT NOT NULL DEFAULT 'common'`,
			`UPDATE pool_leases SET purpose='adobe' WHERE state IN ('used','retiring','retired')`,
			`INSERT INTO pool_legacy_purpose_marks(alias_id,purpose) SELECT m.alias_id,'adobe' FROM pool_members m WHERE m.state='used' AND NOT EXISTS(SELECT 1 FROM pool_leases l WHERE l.alias_id=m.alias_id AND l.state IN ('used','leased','retiring')) ON CONFLICT(alias_id,purpose) DO NOTHING`,
			`INSERT INTO pool_legacy_purpose_marks(alias_id,purpose) SELECT m.alias_id,'adobe' FROM pool_suspended_members m WHERE m.state='used' AND NOT EXISTS(SELECT 1 FROM pool_leases l WHERE l.alias_id=m.alias_id AND l.state IN ('used','leased','retiring')) ON CONFLICT(alias_id,purpose) DO NOTHING`,
		} {
			if _, err := s.txExecContext(ctx, tx, q); err != nil {
				return err
			}
		}
	}
	for _, q := range []string{`DROP INDEX IF EXISTS pool_one_live_lease`, `CREATE UNIQUE INDEX IF NOT EXISTS pool_one_live_purpose_lease ON pool_leases(alias_id,purpose) WHERE state IN ('leased','used','retiring')`} {
		if _, err := s.txExecContext(ctx, tx, q); err != nil {
			return err
		}
	}
	return nil
}

// Credentials are alias-wide: never rotate them while another purpose is live.
func (s *Store) rotatePoolAliasUnlessSharedTx(ctx context.Context, tx *sql.Tx, alias domain.Alias, excludeID string) error {
	var live int
	if err := s.txQueryRowContext(ctx, tx, `SELECT COUNT(*) FROM (SELECT alias_id FROM pool_leases WHERE id<>? AND state IN ('leased','used','retiring') UNION ALL SELECT alias_id FROM pool_legacy_purpose_marks) live WHERE alias_id=?`, excludeID, alias.ID).Scan(&live); err != nil {
		return err
	}
	if live > 0 {
		return nil
	}
	return s.rotatePoolAliasTx(ctx, tx, alias)
}

// The legacy member state remains a global summary for scheduler/retirement.
// Claim eligibility and UI usage states are evaluated separately per purpose.
func (s *Store) refreshPoolMemberSummaryTx(ctx context.Context, tx *sql.Tx, aliasID int64) error {
	for _, table := range []string{"pool_members", "pool_suspended_members"} {
		_, err := s.txExecContext(ctx, tx, `UPDATE `+table+` SET state=CASE WHEN state='paused' THEN 'paused'
  WHEN EXISTS(SELECT 1 FROM pool_leases WHERE alias_id=? AND state IN ('used','retiring')) OR EXISTS(SELECT 1 FROM pool_legacy_purpose_marks WHERE alias_id=?) THEN 'used'
  WHEN EXISTS(SELECT 1 FROM pool_leases WHERE alias_id=? AND state='leased') THEN 'leased' ELSE 'available' END,updated_at=? WHERE alias_id=?`, aliasID, aliasID, aliasID, timestamp(s.now()), aliasID)
		if err != nil {
			return err
		}
	}
	return nil
}
