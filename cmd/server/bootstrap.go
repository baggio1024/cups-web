package main

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"cups-web/internal/store"
	"golang.org/x/crypto/bcrypt"
)

func ensureDefaultAdmin(ctx context.Context) error {
	return appStore.WithTx(ctx, false, func(tx *sql.Tx) error {
		user, err := store.GetUserByUsername(ctx, tx, "admin")
		if err == nil {
			if user.Role != store.RoleAdmin || !user.Protected {
				if _, err := tx.ExecContext(ctx, "UPDATE users SET role = ?, protected = 1 WHERE id = ?", store.RoleAdmin, user.ID); err != nil {
					return err
				}
			}
			return nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}

		hash, err := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		if _, err := store.CreateUser(ctx, tx, store.CreateUserInput{
			Username:     "admin",
			PasswordHash: string(hash),
			Role:         store.RoleAdmin,
			Protected:    true,
		}); err != nil {
			return err
		}
		log.Printf("default admin created: admin")
		return nil
	})
}
