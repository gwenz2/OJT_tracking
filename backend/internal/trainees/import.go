// Two-step trainee CSV import: validate produces a server-side snapshot
// referenced by a short-lived import token; commit inserts that snapshot
// transactionally — the client cannot alter rows between steps.
package trainees

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/skycode/ojt-management/backend/internal/auth"
	"github.com/skycode/ojt-management/backend/internal/platform/db"
	"github.com/skycode/ojt-management/backend/internal/platform/httpx"
)

const importTokenTTL = 15 * time.Minute

// ImportRow is one normalized CSV row.
type ImportRow struct {
	Email         string `json:"email"`
	StudentNumber string `json:"student_number"`
	DisplayName   string `json:"display_name"`
	Program       string `json:"program"`
	YearLevel     string `json:"year_level"`
	ContactNumber string `json:"contact_number"`
	Password      string `json:"-"` // optional explicit password; generated when empty
}

type ImportRowResult struct {
	Row        int               `json:"row"`
	Status     string            `json:"status"` // valid | invalid
	Normalized *ImportRow        `json:"normalized,omitempty"`
	Errors     map[string]string `json:"errors,omitempty"`
}

type ImportPreview struct {
	ValidRows   int               `json:"valid_rows"`
	InvalidRows int               `json:"invalid_rows"`
	Rows        []ImportRowResult `json:"rows"`
	ImportToken string            `json:"import_token"`
}

type importBatch struct {
	rows        []ImportRow
	creatorID   uuid.UUID
	creatorRole string
	expiresAt   time.Time
}

// ImportStore holds validated snapshots in memory. Single-instance MVP —
// process restart invalidates pending tokens, which is safe (re-validate).
type ImportStore struct {
	mu      sync.Mutex
	batches map[string]*importBatch
}

func NewImportStore() *ImportStore {
	return &ImportStore{batches: map[string]*importBatch{}}
}

func (s *ImportStore) put(b *importBatch) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.batches) > 1000 {
		now := time.Now()
		for k, v := range s.batches {
			if now.After(v.expiresAt) {
				delete(s.batches, k)
			}
		}
	}
	token, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	s.batches[token.String()] = b
	return token.String(), nil
}

// take consumes a batch — a token commits at most once.
func (s *ImportStore) take(token string) (*importBatch, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, ok := s.batches[token]
	if !ok || time.Now().After(b.expiresAt) {
		delete(s.batches, token)
		return nil, false
	}
	delete(s.batches, token)
	return b, true
}

var requiredHeaders = []string{"email", "student_number", "display_name"}
var allHeaders = map[string]bool{
	"email": true, "student_number": true, "display_name": true,
	"program": true, "year_level": true, "contact_number": true, "password": true,
}

// ParseAndValidate reads a CSV upload and returns the preview plus the
// validated rows. Duplicate/invalid rows are reported, never silently dropped.
func ParseAndValidate(ctx context.Context, pool *db.Pool, r io.Reader) ([]ImportRow, []ImportRowResult, error) {
	cr := csv.NewReader(r)
	cr.TrimLeadingSpace = true

	header, err := cr.Read()
	if err != nil {
		return nil, nil, httpx.ValidationMsg("The CSV file is empty or unreadable.")
	}
	cols := map[string]int{}
	for i, h := range header {
		h = strings.ToLower(strings.TrimSpace(h))
		if allHeaders[h] {
			cols[h] = i
		}
	}
	for _, req := range requiredHeaders {
		if _, ok := cols[req]; !ok {
			return nil, nil, httpx.ValidationMsg(fmt.Sprintf("Missing required column %q. Expected headers: email, student_number, display_name, program, year_level, contact_number.", req))
		}
	}

	get := func(rec []string, name string) string {
		if i, ok := cols[name]; ok && i < len(rec) {
			return strings.TrimSpace(rec[i])
		}
		return ""
	}

	var rows []ImportRow
	var results []ImportRowResult
	seenEmail := map[string]int{}
	seenStudent := map[string]int{}
	line := 1
	for {
		rec, err := cr.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		line++
		if err != nil {
			results = append(results, ImportRowResult{Row: line, Status: "invalid",
				Errors: map[string]string{"row": "Malformed CSV row."}})
			continue
		}
		row := ImportRow{
			Email:         strings.ToLower(get(rec, "email")),
			StudentNumber: get(rec, "student_number"),
			DisplayName:   get(rec, "display_name"),
			Program:       get(rec, "program"),
			YearLevel:     get(rec, "year_level"),
			ContactNumber: get(rec, "contact_number"),
			Password:      get(rec, "password"),
		}
		fields := map[string]string{}
		if row.Email == "" || !strings.Contains(row.Email, "@") {
			fields["email"] = "A valid email is required."
		} else if prev, dup := seenEmail[row.Email]; dup {
			fields["email"] = fmt.Sprintf("Duplicate of row %d.", prev)
		} else {
			var exists bool
			if err := pool.QueryRow(ctx,
				`SELECT EXISTS(SELECT 1 FROM users WHERE lower(email) = lower($1))`, row.Email).Scan(&exists); err == nil && exists {
				fields["email"] = "Email already exists."
			}
			seenEmail[row.Email] = line
		}
		if row.StudentNumber == "" {
			fields["student_number"] = "Student number is required."
		} else if prev, dup := seenStudent[row.StudentNumber]; dup {
			fields["student_number"] = fmt.Sprintf("Duplicate of row %d.", prev)
		} else {
			var exists bool
			if err := pool.QueryRow(ctx,
				`SELECT EXISTS(SELECT 1 FROM trainee_profiles WHERE student_number = $1)`, row.StudentNumber).Scan(&exists); err == nil && exists {
				fields["student_number"] = "Student number already exists."
			}
			seenStudent[row.StudentNumber] = line
		}
		if row.DisplayName == "" {
			fields["display_name"] = "Display name is required."
		}
		if row.Password != "" && len(row.Password) < 8 {
			fields["password"] = "Password must be at least 8 characters."
		}

		if len(fields) > 0 {
			results = append(results, ImportRowResult{Row: line, Status: "invalid", Errors: fields})
		} else {
			rows = append(rows, row)
			r := row
			results = append(results, ImportRowResult{Row: line, Status: "valid", Normalized: &r})
		}
	}
	return rows, results, nil
}

// Stage stores the validated rows and returns the import token.
func Stage(store *ImportStore, rows []ImportRow, creator *auth.User) (string, error) {
	return store.put(&importBatch{
		rows:        rows,
		creatorID:   creator.ID,
		creatorRole: creator.Role,
		expiresAt:   time.Now().Add(importTokenTTL),
	})
}

type CommitResult struct {
	Created     int `json:"created"`
	Credentials []struct {
		Email             string `json:"email"`
		TemporaryPassword string `json:"temporary_password"`
	} `json:"credentials"`
}

// Commit inserts the staged snapshot in one transaction. Any failure rolls
// the whole batch back — no partial silent import.
func Commit(ctx context.Context, pool *db.Pool, store *ImportStore, token string) (*CommitResult, error) {
	b, ok := store.take(token)
	if !ok {
		return nil, httpx.Validation(map[string]string{"import_token": "Import token is invalid or expired. Re-upload the CSV to validate again."})
	}
	if len(b.rows) == 0 {
		return nil, httpx.ValidationMsg("The import contains no valid rows.")
	}

	res := &CommitResult{}
	err := pool.InTx(ctx, func(tx pgx.Tx) error {
		for _, row := range b.rows {
			password := row.Password
			if password == "" {
				var err error
				password, err = TempPassword()
				if err != nil {
					return err
				}
			}
			hash, err := auth.HashPassword(password)
			if err != nil {
				return err
			}
			var userID uuid.UUID
			if err := tx.QueryRow(ctx, `
				INSERT INTO users (email, password_hash, role, display_name)
				VALUES ($1, $2, 'trainee', $3) RETURNING id`,
				row.Email, hash, row.DisplayName).Scan(&userID); err != nil {
				return err
			}
			var traineeID uuid.UUID
			if err := tx.QueryRow(ctx, `
				INSERT INTO trainee_profiles (user_id, student_number, program, year_level, contact_number)
				VALUES ($1, $2, $3, $4, $5) RETURNING id`,
				userID, row.StudentNumber, nullIfEmpty(row.Program),
				nullIfEmpty(row.YearLevel), nullIfEmpty(row.ContactNumber)).Scan(&traineeID); err != nil {
				return err
			}
			if b.creatorRole == "coordinator" {
				if _, err := tx.Exec(ctx, `
					INSERT INTO coordinator_scopes (coordinator_user_id, trainee_id) VALUES ($1, $2)
					ON CONFLICT DO NOTHING`, b.creatorID, traineeID); err != nil {
					return err
				}
			}
			res.Created++
			if row.Password == "" {
				res.Credentials = append(res.Credentials, struct {
					Email             string `json:"email"`
					TemporaryPassword string `json:"temporary_password"`
				}{Email: row.Email, TemporaryPassword: password})
			}
		}
		return nil
	})
	if err != nil {
		return nil, httpx.Conflict("IMPORT_FAILED", "Import failed — no rows were committed. Check for duplicates and try again.")
	}
	return res, nil
}
