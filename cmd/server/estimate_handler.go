package main

import (
	"database/sql"
	"net/http"
	"time"

	"cups-web/internal/auth"
	"cups-web/internal/store"
)

type estimateResp struct {
	Pages               int   `json:"pages"`
	Estimated           bool  `json:"estimated"`
	PerPageCents        int64 `json:"perPageCents"`
	ColorPageCents      int64 `json:"colorPageCents"`
	CostCents           int64 `json:"costCents"`
	BalanceCents        int64 `json:"balanceCents"`
	MonthSpentCents     int64 `json:"monthSpentCents"`
	YearSpentCents      int64 `json:"yearSpentCents"`
	MonthlyLimitCents   int64 `json:"monthlyLimitCents"`
	YearlyLimitCents    int64 `json:"yearlyLimitCents"`
	InsufficientBalance bool  `json:"insufficientBalance"`
	WouldExceedMonthly  bool  `json:"wouldExceedMonthly"`
	WouldExceedYearly   bool  `json:"wouldExceedYearly"`
}

func estimateHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(64 << 20); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid multipart form")
		return
	}
	file, fh, err := r.FormFile("file")
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "missing file field")
		return
	}
	defer file.Close()

	isColor := r.FormValue("color") == "true"

	tmpPath, cleanup, err := saveTempUpload(file, fh.Filename)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to save file")
		return
	}
	defer cleanup()

	countCtx, cancel := convertTimeoutContext(r.Context())
	defer cancel()
	pages, estimated, err := countPages(countCtx, tmpPath, fh.Filename)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "failed to read pages")
		return
	}
	if pages < 1 {
		pages = 1
	}

	sess, _ := auth.GetSession(r)
	var resp estimateResp
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
		colorPage, err := store.GetSettingInt(r.Context(), tx, store.SettingColorPageCents, store.DefaultColorPageCents)
		if err != nil {
			return err
		}
		var cost int64
		if isColor {
			cost = int64(pages) * colorPage
		} else {
			cost = int64(pages) * perPage
		}
		resp = estimateResp{
			Pages:               pages,
			Estimated:           estimated,
			PerPageCents:        perPage,
			ColorPageCents:      colorPage,
			CostCents:           cost,
			BalanceCents:        user.BalanceCents,
			MonthSpentCents:     user.MonthSpentCents,
			YearSpentCents:      user.YearSpentCents,
			MonthlyLimitCents:   user.MonthlyLimitCents,
			YearlyLimitCents:    user.YearlyLimitCents,
			InsufficientBalance: user.BalanceCents < cost,
			WouldExceedMonthly:  user.MonthlyLimitCents > 0 && user.MonthSpentCents+cost > user.MonthlyLimitCents,
			WouldExceedYearly:   user.YearlyLimitCents > 0 && user.YearSpentCents+cost > user.YearlyLimitCents,
		}
		return nil
	})
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to estimate")
		return
	}
	writeJSON(w, resp)
}
