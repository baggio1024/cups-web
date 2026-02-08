package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"cups-web/internal/auth"
	"cups-web/internal/ipp"
	"cups-web/internal/store"
)

type printResp struct {
	JobID           string `json:"jobId,omitempty"`
	OK              bool   `json:"ok"`
	Pages           int    `json:"pages"`
	CostCents       int64  `json:"costCents"`
	BalanceCents    int64  `json:"balanceCents"`
	MonthSpentCents int64  `json:"monthSpentCents"`
	YearSpentCents  int64  `json:"yearSpentCents"`
	IsDuplex        bool   `json:"isDuplex"`
	IsColor         bool   `json:"isColor"`
}

var (
	errInsufficientBalance = errors.New("insufficient balance")
	errMonthlyLimit        = errors.New("monthly limit exceeded")
	errYearlyLimit         = errors.New("yearly limit exceeded")
)

func printHandler(w http.ResponseWriter, r *http.Request) {
	// Expect multipart form
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

	printer := r.FormValue("printer")
	if printer == "" {
		writeJSONError(w, http.StatusBadRequest, "missing printer field")
		return
	}

	sides := r.FormValue("sides")
	duplexParam := r.FormValue("duplex")
	if sides == "" {
		if duplexParam == "true" {
			sides = "two-sided-long-edge"
		} else {
			sides = "one-sided"
		}
	}
	isDuplex := strings.HasPrefix(sides, "two-sided")
	isColor := r.FormValue("color") == "true"
	
	// 获取打印份数，默认为1
	copies := 1
	if copiesStr := r.FormValue("copies"); copiesStr != "" {
		if c, err := strconv.Atoi(copiesStr); err == nil && c > 0 && c <= 100 {
			copies = c
		}
	}
	
	// 获取页面范围，默认为全部页面
	pageRange := r.FormValue("pageRange")

	storedRel, storedAbs, err := saveUploadedFile(file, fh.Filename, uploadDir)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to save file")
		return
	}

	countCtx, cancel := convertTimeoutContext(r.Context())
	defer cancel()
	printPath := storedAbs
	var printCleanup func()
	printMime := ""
	var pages int
	kind := detectFileKind(storedAbs, fh.Filename)
	switch kind {
	case fileKindPDF:
		var err error
		pages, err = countPDFPages(storedAbs)
		if err != nil {
			_ = os.Remove(storedAbs)
			writeJSONError(w, http.StatusBadRequest, "failed to read pages")
			return
		}
		printMime = "application/pdf"
	case fileKindOffice:
		outPath, cleanup, err := convertOfficeToPDF(countCtx, storedAbs)
		if err != nil {
			_ = os.Remove(storedAbs)
			writeJSONError(w, http.StatusBadRequest, "conversion failed")
			return
		}
		pages, err = countPDFPages(outPath)
		if err != nil {
			cleanup()
			_ = os.Remove(storedAbs)
			writeJSONError(w, http.StatusBadRequest, "failed to read pages")
			return
		}
		_, convertedAbs, err := saveConvertedPDFToUploads(outPath, storedRel, uploadDir)
		if err != nil {
			cleanup()
			_ = os.Remove(storedAbs)
			writeJSONError(w, http.StatusInternalServerError, "failed to save converted file")
			return
		}
		printPath = convertedAbs
		printCleanup = cleanup
		printMime = "application/pdf"
	case fileKindImage:
		outPath, cleanup, err := convertImageToPDF(storedAbs)
		if err != nil {
			_ = os.Remove(storedAbs)
			writeJSONError(w, http.StatusBadRequest, "conversion failed")
			return
		}
		_, convertedAbs, err := saveConvertedPDFToUploads(outPath, storedRel, uploadDir)
		if err != nil {
			cleanup()
			_ = os.Remove(storedAbs)
			writeJSONError(w, http.StatusInternalServerError, "failed to save converted file")
			return
		}
		printPath = convertedAbs
		printCleanup = cleanup
		printMime = "application/pdf"
		pages = 1
	case fileKindText:
		var err error
		pages, err = estimateTextPages(storedAbs)
		if err != nil {
			_ = os.Remove(storedAbs)
			writeJSONError(w, http.StatusBadRequest, "failed to read pages")
			return
		}
		outPath, cleanup, err := convertTextToPDF(storedAbs)
		if err != nil {
			_ = os.Remove(storedAbs)
			writeJSONError(w, http.StatusBadRequest, "conversion failed")
			return
		}
		_, convertedAbs, err := saveConvertedPDFToUploads(outPath, storedRel, uploadDir)
		if err != nil {
			cleanup()
			_ = os.Remove(storedAbs)
			writeJSONError(w, http.StatusInternalServerError, "failed to save converted file")
			return
		}
		printPath = convertedAbs
		printCleanup = cleanup
		printMime = "application/pdf"
	default:
		var err error
		pages, _, err = countPages(countCtx, storedAbs, fh.Filename)
		if err != nil {
			_ = os.Remove(storedAbs)
			writeJSONError(w, http.StatusBadRequest, "failed to read pages")
			return
		}
	}
	if pages < 1 {
		pages = 1
	}
	if printCleanup != nil {
		defer printCleanup()
	}

	sess, _ := auth.GetSession(r)
	var recordID int64
	var balanceAfter int64
	var monthSpent int64
	var yearSpent int64
	var costCents int64

	err = appStore.WithTx(r.Context(), false, func(tx *sql.Tx) error {
		user, err := store.GetUserByID(r.Context(), tx, sess.UserID)
		if err != nil {
			return err
		}
		if err := normalizeUserPeriods(r.Context(), tx, &user, time.Now()); err != nil {
			return err
		}
		costCents = 0
		before := user.BalanceCents
		balanceAfter = before
		monthSpent = user.MonthSpentCents
		yearSpent = user.YearSpentCents

		rec := store.PrintRecord{
			UserID:             user.ID,
			PrinterURI:         printer,
			Filename:           fh.Filename,
			StoredPath:         storedRel,
			Pages:              pages,
			CostCents:          costCents,
			BalanceBeforeCents: before,
			BalanceAfterCents:  balanceAfter,
			MonthTotalCents:    monthSpent,
			YearTotalCents:     yearSpent,
			Status:             "queued",
			IsDuplex:           isDuplex,
			IsColor:            isColor,
			Duplex:             sql.NullString{String: getDuplexDisplayText(sides), Valid: getDuplexDisplayText(sides) != ""},
			Sides:              sql.NullString{String: sides, Valid: sides != ""},
			Copies:             copies,
			PageRange:          sql.NullString{String: pageRange, Valid: pageRange != ""},
			CreatedAt:          time.Now().UTC().Format(time.RFC3339),
		}
		id, err := store.InsertPrintRecord(r.Context(), tx, &rec)
		if err != nil {
			return err
		}
		recordID = id
		return nil
	})
	if err != nil {
		_ = os.Remove(storedAbs)
		writeJSONError(w, http.StatusInternalServerError, "failed to create print record")
		return
	}

	f, err := os.Open(printPath)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to open file")
		return
	}
	defer f.Close()

	mime := printMime
	if mime == "" {
		mime = fh.Header.Get("Content-Type")
	}
	if mime == "" {
		buf := make([]byte, 512)
		if n, _ := f.Read(buf); n > 0 {
			mime = http.DetectContentType(buf[:n])
			if _, err := f.Seek(0, io.SeekStart); err != nil {
				_ = refundPrint(r.Context(), recordID, sess.UserID, costCents)
				writeJSONError(w, http.StatusInternalServerError, "failed to read file")
				return
			}
		}
	}

	job, err := ipp.SendPrintJob(printer, f, mime, sess.Username, fh.Filename, sides, isColor, copies, pageRange)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "print error: "+err.Error())
		return
	}

	_ = appStore.WithTx(r.Context(), false, func(tx *sql.Tx) error {
		return store.UpdatePrintStatus(r.Context(), tx, recordID, "printed", job)
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(printResp{
		JobID:           job,
		OK:              true,
		Pages:           pages,
		CostCents:       costCents,
		BalanceCents:    balanceAfter,
		MonthSpentCents: monthSpent,
		YearSpentCents:  yearSpent,
		IsDuplex:        isDuplex,
		IsColor:         isColor,
	})
}

func getDuplexDisplayText(sides string) string {
	switch sides {
	case "one-sided":
		return "单面打印"
	case "two-sided-long-edge":
		return "双面打印（长边翻转）"
	case "two-sided-short-edge":
		return "双面打印（短边翻转）"
	default:
		if strings.Contains(sides, "one") {
			return "单面打印"
		} else if strings.Contains(sides, "two") {
			if strings.Contains(sides, "short") {
				return "双面打印（短边翻转）"
			} else {
				return "双面打印（长边翻转）"
			}
		}
		return "单面打印" // default
	}
}

