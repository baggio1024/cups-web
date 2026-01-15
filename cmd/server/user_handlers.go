package main

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"cups-web/internal/auth"
	"cups-web/internal/store"
)

type meResponse struct {
	ID                int64  `json:"id"`
	Username          string `json:"username"`
	Role              string `json:"role"`
	BalanceCents      int64  `json:"balanceCents"`
	PerPageCents      int64  `json:"perPageCents"`
	MonthSpentCents   int64  `json:"monthSpentCents"`
	YearSpentCents    int64  `json:"yearSpentCents"`
	MonthlyLimitCents int64  `json:"monthlyLimitCents"`
	YearlyLimitCents  int64  `json:"yearlyLimitCents"`
}

func MeHandler(w http.ResponseWriter, r *http.Request) {
	sess, err := auth.GetSession(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var resp meResponse
	err = appStore.WithTx(r.Context(), false, func(tx *sql.Tx) error {
		user, err := store.GetUserByID(r.Context(), tx, sess.UserID)
		if err != nil {
			return err
		}
		if err := normalizeUserPeriods(r.Context(), tx, &user, time.Now()); err != nil {
			return err
		}
		perPage, err := store.GetSettingInt(r.Context(), tx, store.SettingPerPageCents, store.DefaultPerPageCents)
		if err != nil {
			return err
		}
		resp = meResponse{
			ID:                user.ID,
			Username:          user.Username,
			Role:              user.Role,
			BalanceCents:      user.BalanceCents,
			PerPageCents:      perPage,
			MonthSpentCents:   user.MonthSpentCents,
			YearSpentCents:    user.YearSpentCents,
			MonthlyLimitCents: user.MonthlyLimitCents,
			YearlyLimitCents:  user.YearlyLimitCents,
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		} else {
			writeJSONError(w, http.StatusInternalServerError, "failed to load profile")
		}
		return
	}
	writeJSON(w, resp)
}
