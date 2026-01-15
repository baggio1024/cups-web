package store

import (
	"context"
	"database/sql"
	"time"
)

type User struct {
	ID                int64
	Username          string
	PasswordHash      string
	Role              string
	Protected         bool
	ContactName       string
	Phone             string
	Email             string
	BalanceCents      int64
	DailyTopupCents   int64
	MonthlyTopupCents int64
	YearlyTopupCents  int64
	MonthlyLimitCents int64
	YearlyLimitCents  int64
	MonthSpentCents   int64
	YearSpentCents    int64
	MonthPeriod       string
	YearPeriod        string
	LastDailyTopup    string
	LastMonthlyTopup  string
	LastYearlyTopup   string
	CreatedAt         string
	UpdatedAt         string
}

type CreateUserInput struct {
	Username          string
	PasswordHash      string
	Role              string
	Protected         bool
	ContactName       string
	Phone             string
	Email             string
	BalanceCents      int64
	DailyTopupCents   int64
	MonthlyTopupCents int64
	YearlyTopupCents  int64
	MonthlyLimitCents int64
	YearlyLimitCents  int64
}

type UpdateUserInput struct {
	ID                int64
	Username          string
	PasswordHash      *string
	Role              string
	ContactName       string
	Phone             string
	Email             string
	DailyTopupCents   int64
	MonthlyTopupCents int64
	YearlyTopupCents  int64
	MonthlyLimitCents int64
	YearlyLimitCents  int64
}

func CountUsers(ctx context.Context, tx *sql.Tx) (int, error) {
	var count int
	if err := tx.QueryRowContext(ctx, "SELECT COUNT(1) FROM users").Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func GetUserByUsername(ctx context.Context, tx *sql.Tx, username string) (User, error) {
	row := tx.QueryRowContext(ctx, `SELECT
		id, username, password_hash, role, protected, contact_name, phone, email,
		balance_cents, daily_topup_cents, monthly_topup_cents, yearly_topup_cents,
		monthly_limit_cents, yearly_limit_cents, month_spent_cents, year_spent_cents,
		month_period, year_period, last_daily_topup, last_monthly_topup, last_yearly_topup,
		created_at, updated_at
		FROM users WHERE username = ?`, username)
	return scanUser(row)
}

func GetUserByID(ctx context.Context, tx *sql.Tx, id int64) (User, error) {
	row := tx.QueryRowContext(ctx, `SELECT
		id, username, password_hash, role, protected, contact_name, phone, email,
		balance_cents, daily_topup_cents, monthly_topup_cents, yearly_topup_cents,
		monthly_limit_cents, yearly_limit_cents, month_spent_cents, year_spent_cents,
		month_period, year_period, last_daily_topup, last_monthly_topup, last_yearly_topup,
		created_at, updated_at
		FROM users WHERE id = ?`, id)
	return scanUser(row)
}

func ListUsers(ctx context.Context, tx *sql.Tx) ([]User, error) {
	rows, err := tx.QueryContext(ctx, `SELECT
		id, username, password_hash, role, protected, contact_name, phone, email,
		balance_cents, daily_topup_cents, monthly_topup_cents, yearly_topup_cents,
		monthly_limit_cents, yearly_limit_cents, month_spent_cents, year_spent_cents,
		month_period, year_period, last_daily_topup, last_monthly_topup, last_yearly_topup,
		created_at, updated_at
		FROM users ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []User{}
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

func CreateUser(ctx context.Context, tx *sql.Tx, input CreateUserInput) (User, error) {
	now := nowUTC()
	monthPeriod := time.Now().Format("2006-01")
	yearPeriod := time.Now().Format("2006")
	res, err := tx.ExecContext(ctx, `INSERT INTO users (
		username, password_hash, role, protected, contact_name, phone, email,
		balance_cents, daily_topup_cents, monthly_topup_cents, yearly_topup_cents,
		monthly_limit_cents, yearly_limit_cents,
		month_spent_cents, year_spent_cents, month_period, year_period,
		last_daily_topup, last_monthly_topup, last_yearly_topup,
		created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, 0, ?, ?, '', '', '', ?, ?)`,
		input.Username, input.PasswordHash, input.Role, input.Protected, input.ContactName, input.Phone, input.Email,
		input.BalanceCents, input.DailyTopupCents, input.MonthlyTopupCents, input.YearlyTopupCents,
		input.MonthlyLimitCents, input.YearlyLimitCents,
		monthPeriod, yearPeriod,
		now, now,
	)
	if err != nil {
		return User{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return User{}, err
	}
	return GetUserByID(ctx, tx, id)
}

func UpdateUser(ctx context.Context, tx *sql.Tx, input UpdateUserInput) (User, error) {
	now := nowUTC()
	if input.PasswordHash != nil {
		if _, err := tx.ExecContext(ctx, `UPDATE users SET
			username = ?, password_hash = ?, role = ?, contact_name = ?, phone = ?, email = ?,
			daily_topup_cents = ?, monthly_topup_cents = ?, yearly_topup_cents = ?,
			monthly_limit_cents = ?, yearly_limit_cents = ?, updated_at = ?
			WHERE id = ?`,
			input.Username, *input.PasswordHash, input.Role, input.ContactName, input.Phone, input.Email,
			input.DailyTopupCents, input.MonthlyTopupCents, input.YearlyTopupCents,
			input.MonthlyLimitCents, input.YearlyLimitCents, now, input.ID,
		); err != nil {
			return User{}, err
		}
	} else {
		if _, err := tx.ExecContext(ctx, `UPDATE users SET
			username = ?, role = ?, contact_name = ?, phone = ?, email = ?,
			daily_topup_cents = ?, monthly_topup_cents = ?, yearly_topup_cents = ?,
			monthly_limit_cents = ?, yearly_limit_cents = ?, updated_at = ?
			WHERE id = ?`,
			input.Username, input.Role, input.ContactName, input.Phone, input.Email,
			input.DailyTopupCents, input.MonthlyTopupCents, input.YearlyTopupCents,
			input.MonthlyLimitCents, input.YearlyLimitCents, now, input.ID,
		); err != nil {
			return User{}, err
		}
	}
	return GetUserByID(ctx, tx, input.ID)
}

func DeleteUser(ctx context.Context, tx *sql.Tx, id int64) error {
	res, err := tx.ExecContext(ctx, "DELETE FROM users WHERE id = ?", id)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err == nil && affected == 0 {
		return sql.ErrNoRows
	}
	return err
}

type scanner interface {
	Scan(dest ...interface{}) error
}

func scanUser(s scanner) (User, error) {
	var user User
	err := s.Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.Protected, &user.ContactName, &user.Phone, &user.Email,
		&user.BalanceCents, &user.DailyTopupCents, &user.MonthlyTopupCents, &user.YearlyTopupCents,
		&user.MonthlyLimitCents, &user.YearlyLimitCents, &user.MonthSpentCents, &user.YearSpentCents,
		&user.MonthPeriod, &user.YearPeriod, &user.LastDailyTopup, &user.LastMonthlyTopup, &user.LastYearlyTopup,
		&user.CreatedAt, &user.UpdatedAt,
	)
	return user, err
}
