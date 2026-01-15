package main

import (
	"context"
	"database/sql"
	"time"

	"cups-web/internal/store"
)

func normalizeUserPeriods(ctx context.Context, tx *sql.Tx, user *store.User, now time.Time) error {
	monthPeriod := now.Format("2006-01")
	yearPeriod := now.Format("2006")
	updated := false

	if user.MonthPeriod != monthPeriod {
		user.MonthPeriod = monthPeriod
		user.MonthSpentCents = 0
		updated = true
	}
	if user.YearPeriod != yearPeriod {
		user.YearPeriod = yearPeriod
		user.YearSpentCents = 0
		updated = true
	}
	if !updated {
		return nil
	}
	_, err := tx.ExecContext(ctx, `UPDATE users SET
		month_period = ?, year_period = ?, month_spent_cents = ?, year_spent_cents = ?, updated_at = ?
		WHERE id = ?`,
		user.MonthPeriod, user.YearPeriod, user.MonthSpentCents, user.YearSpentCents, time.Now().UTC().Format(time.RFC3339), user.ID,
	)
	return err
}
