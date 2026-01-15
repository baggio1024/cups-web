package main

import (
	"context"
	"database/sql"
	"time"

	"cups-web/internal/store"
)

func refundPrint(ctx context.Context, recordID int64, userID int64, costCents int64) error {
	return appStore.WithTx(ctx, false, func(tx *sql.Tx) error {
		user, err := store.GetUserByID(ctx, tx, userID)
		if err != nil {
			return err
		}
		balance := user.BalanceCents + costCents
		monthSpent := user.MonthSpentCents
		yearSpent := user.YearSpentCents
		if monthSpent >= costCents {
			monthSpent -= costCents
		} else {
			monthSpent = 0
		}
		if yearSpent >= costCents {
			yearSpent -= costCents
		} else {
			yearSpent = 0
		}
		if _, err := tx.ExecContext(ctx, `UPDATE users SET
			balance_cents = ?, month_spent_cents = ?, year_spent_cents = ?, updated_at = ?
			WHERE id = ?`, balance, monthSpent, yearSpent, time.Now().UTC().Format(time.RFC3339), user.ID,
		); err != nil {
			return err
		}
		return store.UpdatePrintStatus(ctx, tx, recordID, "failed", "")
	})
}
