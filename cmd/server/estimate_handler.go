package main

import (
	"net/http"
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

	_ = r.FormValue("color")

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

	resp := estimateResp{
		Pages:               pages,
		Estimated:           estimated,
		PerPageCents:        0,
		ColorPageCents:      0,
		CostCents:           0,
		BalanceCents:        0,
		MonthSpentCents:     0,
		YearSpentCents:      0,
		MonthlyLimitCents:   0,
		YearlyLimitCents:    0,
		InsufficientBalance: false,
		WouldExceedMonthly:  false,
		WouldExceedYearly:   false,
	}
	writeJSON(w, resp)
}
