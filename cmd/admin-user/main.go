package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"net/mail"
	"os"
	"strings"
	"time"

	"example.com/786-real-estate/backend/internal/config"
	"example.com/786-real-estate/backend/internal/database"
	"example.com/786-real-estate/backend/internal/httpapi"
	"github.com/jackc/pgx/v5"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	email := flag.String("email", "", "Admin email address")
	name := flag.String("name", "Admin", "Display name")
	previous := flag.String("replace-email", "", "Existing account to rename; retains its ID and role")
	flag.Parse()
	*email = strings.ToLower(strings.TrimSpace(*email))
	address, err := mail.ParseAddress(*email)
	if err != nil || address.Address != *email {
		return fmt.Errorf("provide a valid --email")
	}
	fmt.Fprintln(os.Stderr, "Read password from stdin (never pass it as a command-line argument).")
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && len(line) == 0 {
		return fmt.Errorf("password is required on stdin")
	}
	password := strings.TrimRight(line, "\r\n")
	if len(password) < 10 {
		return fmt.Errorf("password must contain at least 10 characters")
	}
	hash, err := httpapi.HashAdminPassword(password)
	if err != nil {
		return fmt.Errorf("could not hash password")
	}
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	db, err := database.Open(ctx, cfg)
	if err != nil {
		return err
	}
	defer db.Close()
	tx, err := db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("could not start account update")
	}
	defer tx.Rollback(context.Background())
	lookup := *email
	if *previous != "" {
		lookup = strings.ToLower(strings.TrimSpace(*previous))
	}
	var id int64
	err = tx.QueryRow(ctx, `select id from admin_users where email = $1 for update`, lookup).Scan(&id)
	if err == pgx.ErrNoRows && *previous == "" {
		err = tx.QueryRow(ctx, `insert into admin_users (name, email, password_hash, role) values ($1,$2,$3,'owner') returning id`, *name, *email, hash).Scan(&id)
	} else if err == nil {
		_, err = tx.Exec(ctx, `update admin_users set name=$1, email=$2, password_hash=$3 where id=$4`, *name, *email, hash, id)
	}
	if err != nil {
		return fmt.Errorf("account update failed; check the existing email and ensure the new email is unique")
	}
	if _, err = tx.Exec(ctx, `delete from admin_sessions where admin_user_id=$1`, id); err != nil {
		return fmt.Errorf("could not revoke old sessions; apply migrations first")
	}
	if _, err = tx.Exec(ctx, `update admin_password_reset_tokens set used_at=now() where admin_user_id=$1 and used_at is null`, id); err != nil {
		return fmt.Errorf("could not invalidate old reset links")
	}
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("could not commit account update")
	}
	fmt.Println("Admin account updated; old sessions and reset links revoked.")
	return nil
}
