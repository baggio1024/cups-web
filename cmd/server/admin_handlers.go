package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"

	"cups-web/internal/auth"
	"cups-web/internal/store"
)

var (
	errDeleteDefaultAdmin = errors.New("default admin cannot be deleted")
	errProtectedRole      = errors.New("protected admin role cannot change")
	errAdminRename        = errors.New("admin username cannot change")
)

type adminUserPayload struct {
	Username          string `json:"username"`
	Password          string `json:"password"`
	Role              string `json:"role"`
	ContactName       string `json:"contactName"`
	Phone             string `json:"phone"`
	Email             string `json:"email"`
	BalanceCents      int64  `json:"balanceCents"`
	DailyTopupCents   int64  `json:"dailyTopupCents"`
	MonthlyTopupCents int64  `json:"monthlyTopupCents"`
	YearlyTopupCents  int64  `json:"yearlyTopupCents"`
	MonthlyLimitCents int64  `json:"monthlyLimitCents"`
	YearlyLimitCents  int64  `json:"yearlyLimitCents"`
}

type adminUserResponse struct {
	ID                int64  `json:"id"`
	Username          string `json:"username"`
	Role              string `json:"role"`
	Protected         bool   `json:"protected"`
	ContactName       string `json:"contactName"`
	Phone             string `json:"phone"`
	Email             string `json:"email"`
	BalanceCents      int64  `json:"balanceCents"`
	DailyTopupCents   int64  `json:"dailyTopupCents"`
	MonthlyTopupCents int64  `json:"monthlyTopupCents"`
	YearlyTopupCents  int64  `json:"yearlyTopupCents"`
	MonthlyLimitCents int64  `json:"monthlyLimitCents"`
	YearlyLimitCents  int64  `json:"yearlyLimitCents"`
	CreatedAt         string `json:"createdAt"`
	UpdatedAt         string `json:"updatedAt"`
}

type topupPayload struct {
	AmountCents int64 `json:"amountCents"`
}

type settingsPayload struct {
	PerPageCents  *int64 `json:"perPageCents"`
	RetentionDays *int64 `json:"retentionDays"`
}

type topupResponse struct {
	ID                 int64  `json:"id"`
	UserID             int64  `json:"userId"`
	Username           string `json:"username"`
	AmountCents        int64  `json:"amountCents"`
	BalanceBeforeCents int64  `json:"balanceBeforeCents"`
	BalanceAfterCents  int64  `json:"balanceAfterCents"`
	Type               string `json:"type"`
	OperatorUserID     *int64 `json:"operatorUserId"`
	OperatorName       string `json:"operatorName"`
	CreatedAt          string `json:"createdAt"`
}

func adminListUsersHandler(w http.ResponseWriter, r *http.Request) {
	var resp []adminUserResponse
	err := appStore.WithTx(r.Context(), true, func(tx *sql.Tx) error {
		users, err := store.ListUsers(r.Context(), tx)
		if err != nil {
			return err
		}
		resp = mapAdminUsers(users)
		return nil
	})
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to list users")
		return
	}
	writeJSON(w, resp)
}

func adminCreateUserHandler(w http.ResponseWriter, r *http.Request) {
	var payload adminUserPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	payload.Username = strings.TrimSpace(payload.Username)
	if payload.Username == "" || payload.Password == "" {
		writeJSONError(w, http.StatusBadRequest, "username and password required")
		return
	}
	role := normalizeRole(payload.Role)
	if role == "" {
		writeJSONError(w, http.StatusBadRequest, "invalid role")
		return
	}
	if payload.BalanceCents < 0 || payload.DailyTopupCents < 0 || payload.MonthlyTopupCents < 0 || payload.YearlyTopupCents < 0 ||
		payload.MonthlyLimitCents < 0 || payload.YearlyLimitCents < 0 {
		writeJSONError(w, http.StatusBadRequest, "invalid amounts")
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to hash password")
		return
	}

	var created store.User
	err = appStore.WithTx(r.Context(), false, func(tx *sql.Tx) error {
		user, err := store.CreateUser(r.Context(), tx, store.CreateUserInput{
			Username:          payload.Username,
			PasswordHash:      string(hash),
			Role:              role,
			Protected:         false,
			ContactName:       payload.ContactName,
			Phone:             payload.Phone,
			Email:             payload.Email,
			BalanceCents:      payload.BalanceCents,
			DailyTopupCents:   payload.DailyTopupCents,
			MonthlyTopupCents: payload.MonthlyTopupCents,
			YearlyTopupCents:  payload.YearlyTopupCents,
			MonthlyLimitCents: payload.MonthlyLimitCents,
			YearlyLimitCents:  payload.YearlyLimitCents,
		})
		if err != nil {
			return err
		}
		created = user
		return nil
	})
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to create user")
		return
	}
	writeJSON(w, mapAdminUser(created))
}

func adminUpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	var payload adminUserPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	payload.Username = strings.TrimSpace(payload.Username)
	if payload.Username == "" {
		writeJSONError(w, http.StatusBadRequest, "username required")
		return
	}
	role := normalizeRole(payload.Role)
	if role == "" {
		writeJSONError(w, http.StatusBadRequest, "invalid role")
		return
	}
	if payload.DailyTopupCents < 0 || payload.MonthlyTopupCents < 0 || payload.YearlyTopupCents < 0 ||
		payload.MonthlyLimitCents < 0 || payload.YearlyLimitCents < 0 {
		writeJSONError(w, http.StatusBadRequest, "invalid amounts")
		return
	}

	var pwdHash *string
	if strings.TrimSpace(payload.Password) != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to hash password")
			return
		}
		h := string(hash)
		pwdHash = &h
	}

	var updated store.User
	err = appStore.WithTx(r.Context(), false, func(tx *sql.Tx) error {
		current, err := store.GetUserByID(r.Context(), tx, id)
		if err != nil {
			return err
		}
		if current.Username == "admin" && payload.Username != "admin" {
			return errAdminRename
		}
		if current.Username == "admin" && role != store.RoleAdmin {
			return errProtectedRole
		}
		if current.Username == "admin" {
			role = store.RoleAdmin
		}

		user, err := store.UpdateUser(r.Context(), tx, store.UpdateUserInput{
			ID:                id,
			Username:          payload.Username,
			PasswordHash:      pwdHash,
			Role:              role,
			ContactName:       payload.ContactName,
			Phone:             payload.Phone,
			Email:             payload.Email,
			DailyTopupCents:   payload.DailyTopupCents,
			MonthlyTopupCents: payload.MonthlyTopupCents,
			YearlyTopupCents:  payload.YearlyTopupCents,
			MonthlyLimitCents: payload.MonthlyLimitCents,
			YearlyLimitCents:  payload.YearlyLimitCents,
		})
		if err != nil {
			return err
		}
		updated = user
		return nil
	})
	if err != nil {
		if errors.Is(err, errAdminRename) {
			writeJSONError(w, http.StatusBadRequest, errAdminRename.Error())
			return
		}
		if errors.Is(err, errProtectedRole) {
			writeJSONError(w, http.StatusBadRequest, "admin role cannot change")
			return
		}
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, http.StatusNotFound, "user not found")
		} else {
			writeJSONError(w, http.StatusInternalServerError, "failed to update user")
		}
		return
	}
	writeJSON(w, mapAdminUser(updated))
}

func adminDeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	sess, _ := auth.GetSession(r)
	if sess.UserID == id {
		writeJSONError(w, http.StatusBadRequest, "cannot delete current user")
		return
	}
	err = appStore.WithTx(r.Context(), false, func(tx *sql.Tx) error {
		user, err := store.GetUserByID(r.Context(), tx, id)
		if err != nil {
			return err
		}
		if user.Username == "admin" {
			return errDeleteDefaultAdmin
		}
		return store.DeleteUser(r.Context(), tx, id)
	})
	if err != nil {
		if errors.Is(err, errDeleteDefaultAdmin) {
			writeJSONError(w, http.StatusBadRequest, "admin cannot be deleted")
			return
		}
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, http.StatusNotFound, "user not found")
		} else {
			writeJSONError(w, http.StatusInternalServerError, "failed to delete user")
		}
		return
	}
	writeJSON(w, map[string]bool{"ok": true})
}

func adminTopupHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	var payload topupPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	if payload.AmountCents <= 0 {
		writeJSONError(w, http.StatusBadRequest, "amount must be positive")
		return
	}
	sess, _ := auth.GetSession(r)

	var newBalance int64
	err = appStore.WithTx(r.Context(), false, func(tx *sql.Tx) error {
		user, err := store.GetUserByID(r.Context(), tx, id)
		if err != nil {
			return err
		}
		before := user.BalanceCents
		after := before + payload.AmountCents
		if _, err := tx.ExecContext(r.Context(), "UPDATE users SET balance_cents = ?, updated_at = ? WHERE id = ?", after, nowRFC3339(), user.ID); err != nil {
			return err
		}
		opID := sess.UserID
		if _, err := store.InsertTopup(r.Context(), tx, user.ID, payload.AmountCents, before, after, "manual", &opID, sess.Username); err != nil {
			return err
		}
		newBalance = after
		return nil
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, http.StatusNotFound, "user not found")
		} else {
			writeJSONError(w, http.StatusInternalServerError, "failed to top up")
		}
		return
	}
	writeJSON(w, map[string]int64{"balanceCents": newBalance})
}

func adminTopupsHandler(w http.ResponseWriter, r *http.Request) {
	startAt, endAt, err := parseDateRange(r)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid date range")
		return
	}
	username := r.URL.Query().Get("username")

	var records []store.TopupRecord
	err = appStore.WithTx(r.Context(), true, func(tx *sql.Tx) error {
		list, err := store.ListTopups(r.Context(), tx, store.TopupFilter{
			Username: username,
			StartAt:  startAt,
			EndAt:    endAt,
		})
		if err != nil {
			return err
		}
		records = list
		return nil
	})
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to load topups")
		return
	}
	writeJSON(w, mapTopups(records))
}

func adminGetSettingsHandler(w http.ResponseWriter, r *http.Request) {
	var perPage int64
	var retention int64
	err := appStore.WithTx(r.Context(), true, func(tx *sql.Tx) error {
		val, err := store.GetSettingInt(r.Context(), tx, store.SettingPerPageCents, store.DefaultPerPageCents)
		if err != nil {
			return err
		}
		perPage = val
		val, err = store.GetSettingInt(r.Context(), tx, store.SettingRetentionDays, 0)
		if err != nil {
			return err
		}
		retention = val
		return nil
	})
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to load settings")
		return
	}
	writeJSON(w, map[string]int64{"perPageCents": perPage, "retentionDays": retention})
}

func adminUpdateSettingsHandler(w http.ResponseWriter, r *http.Request) {
	var payload settingsPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	err := appStore.WithTx(r.Context(), false, func(tx *sql.Tx) error {
		if payload.PerPageCents != nil {
			if *payload.PerPageCents < 0 {
				return errors.New("invalid perPageCents")
			}
			if err := store.SetSettingInt(r.Context(), tx, store.SettingPerPageCents, *payload.PerPageCents); err != nil {
				return err
			}
		}
		if payload.RetentionDays != nil {
			if *payload.RetentionDays < 0 {
				return errors.New("invalid retentionDays")
			}
			if err := store.SetSettingInt(r.Context(), tx, store.SettingRetentionDays, *payload.RetentionDays); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, map[string]bool{"ok": true})
}

func normalizeRole(role string) string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "":
		return store.RoleUser
	case store.RoleUser:
		return store.RoleUser
	case store.RoleAdmin:
		return store.RoleAdmin
	default:
		return ""
	}
}

func parseIDParam(r *http.Request) (int64, error) {
	idStr := mux.Vars(r)["id"]
	return strconv.ParseInt(idStr, 10, 64)
}

func mapAdminUsers(users []store.User) []adminUserResponse {
	resp := make([]adminUserResponse, 0, len(users))
	for _, user := range users {
		resp = append(resp, mapAdminUser(user))
	}
	return resp
}

func mapAdminUser(user store.User) adminUserResponse {
	return adminUserResponse{
		ID:                user.ID,
		Username:          user.Username,
		Role:              user.Role,
		Protected:         user.Username == "admin",
		ContactName:       user.ContactName,
		Phone:             user.Phone,
		Email:             user.Email,
		BalanceCents:      user.BalanceCents,
		DailyTopupCents:   user.DailyTopupCents,
		MonthlyTopupCents: user.MonthlyTopupCents,
		YearlyTopupCents:  user.YearlyTopupCents,
		MonthlyLimitCents: user.MonthlyLimitCents,
		YearlyLimitCents:  user.YearlyLimitCents,
		CreatedAt:         user.CreatedAt,
		UpdatedAt:         user.UpdatedAt,
	}
}

func mapTopups(records []store.TopupRecord) []topupResponse {
	resp := make([]topupResponse, 0, len(records))
	for _, rec := range records {
		var opID *int64
		if rec.OperatorUserID.Valid {
			id := rec.OperatorUserID.Int64
			opID = &id
		}
		resp = append(resp, topupResponse{
			ID:                 rec.ID,
			UserID:             rec.UserID,
			Username:           rec.Username,
			AmountCents:        rec.AmountCents,
			BalanceBeforeCents: rec.BalanceBeforeCents,
			BalanceAfterCents:  rec.BalanceAfterCents,
			Type:               rec.Type,
			OperatorUserID:     opID,
			OperatorName:       rec.OperatorName,
			CreatedAt:          rec.CreatedAt,
		})
	}
	return resp
}

func nowRFC3339() string {
	return time.Now().UTC().Format(time.RFC3339)
}
