package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type TopupRecord struct {
	ID                 int64
	UserID             int64
	Username           string
	AmountCents        int64
	BalanceBeforeCents int64
	BalanceAfterCents  int64
	Type               string
	OperatorUserID     sql.NullInt64
	OperatorName       string
	CreatedAt          string
}

type TopupFilter struct {
	Username string
	StartAt  string
	EndAt    string
	Limit    int
}

func InsertTopup(ctx context.Context, tx *sql.Tx, userID int64, amountCents int64, beforeCents int64, afterCents int64, typ string, operatorUserID *int64, operatorName string) (int64, error) {
	var opID sql.NullInt64
	if operatorUserID != nil {
		opID = sql.NullInt64{Int64: *operatorUserID, Valid: true}
	}
	res, err := tx.ExecContext(ctx, `INSERT INTO topups (
		user_id, amount_cents, balance_before_cents, balance_after_cents, type,
		operator_user_id, operator_name, created_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		userID, amountCents, beforeCents, afterCents, typ, opID, operatorName, nowUTC(),
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func ListTopups(ctx context.Context, tx *sql.Tx, filter TopupFilter) ([]TopupRecord, error) {
	args := []interface{}{}
	conds := []string{"1=1"}
	if filter.Username != "" {
		conds = append(conds, "u.username = ?")
		args = append(args, filter.Username)
	}
	if filter.StartAt != "" {
		conds = append(conds, "t.created_at >= ?")
		args = append(args, filter.StartAt)
	}
	if filter.EndAt != "" {
		conds = append(conds, "t.created_at <= ?")
		args = append(args, filter.EndAt)
	}
	query := fmt.Sprintf(`SELECT
		t.id, t.user_id, u.username, t.amount_cents, t.balance_before_cents, t.balance_after_cents,
		t.type, t.operator_user_id, t.operator_name, t.created_at
		FROM topups t
		JOIN users u ON u.id = t.user_id
		WHERE %s
		ORDER BY t.created_at DESC`, strings.Join(conds, " AND "))
	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
	}

	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []TopupRecord
	for rows.Next() {
		var rec TopupRecord
		if err := rows.Scan(
			&rec.ID, &rec.UserID, &rec.Username, &rec.AmountCents, &rec.BalanceBeforeCents, &rec.BalanceAfterCents,
			&rec.Type, &rec.OperatorUserID, &rec.OperatorName, &rec.CreatedAt,
		); err != nil {
			return nil, err
		}
		records = append(records, rec)
	}
	return records, rows.Err()
}
