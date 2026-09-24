// Package sites implements OJT site CRUD with coordinate/radius validation.
package sites

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/skycode/ojt-management/backend/internal/platform/db"
	"github.com/skycode/ojt-management/backend/internal/platform/httpx"
)

type Site struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	Address        string    `json:"address"`
	Latitude       float64   `json:"latitude"`
	Longitude      float64   `json:"longitude"`
	AllowedRadiusM int       `json:"allowed_radius_m"`
	IsActive       bool      `json:"is_active"`
}

type Input struct {
	Name           *string  `json:"name"`
	Address        *string  `json:"address"`
	Latitude       *float64 `json:"latitude"`
	Longitude      *float64 `json:"longitude"`
	AllowedRadiusM *int     `json:"allowed_radius_m"`
	IsActive       *bool    `json:"is_active"`
}

func validate(name, address string, lat, lng float64, radius int) *httpx.APIError {
	fields := map[string]string{}
	if strings.TrimSpace(name) == "" {
		fields["name"] = "Site name is required."
	}
	if strings.TrimSpace(address) == "" {
		fields["address"] = "Address is required."
	}
	if lat < -90 || lat > 90 {
		fields["latitude"] = "Latitude must be between -90 and 90."
	}
	if lng < -180 || lng > 180 {
		fields["longitude"] = "Longitude must be between -180 and 180."
	}
	if radius <= 0 {
		fields["allowed_radius_m"] = "Allowed radius must be a positive number of meters."
	}
	if len(fields) > 0 {
		return httpx.Validation(fields)
	}
	return nil
}

func List(ctx context.Context, pool *db.Pool, p httpx.Page, q string, active *bool) ([]Site, int64, error) {
	where := `WHERE ($1::text IS NULL OR name ILIKE '%' || $1 || '%' OR address ILIKE '%' || $1 || '%')
		AND ($2::boolean IS NULL OR is_active = $2)`
	var qArg any
	if q = strings.TrimSpace(q); q != "" {
		qArg = q
	}

	var total int64
	if err := pool.QueryRow(ctx,
		`SELECT count(*) FROM ojt_sites `+where, qArg, active).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := pool.Query(ctx, `
		SELECT id, name, address, latitude, longitude, allowed_radius_m, is_active
		FROM ojt_sites `+where+`
		ORDER BY name
		LIMIT $3 OFFSET $4`, qArg, active, p.Limit(), p.Offset())
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []Site
	for rows.Next() {
		var s Site
		if err := rows.Scan(&s.ID, &s.Name, &s.Address, &s.Latitude, &s.Longitude,
			&s.AllowedRadiusM, &s.IsActive); err != nil {
			return nil, 0, err
		}
		out = append(out, s)
	}
	return out, total, rows.Err()
}

func Get(ctx context.Context, pool *db.Pool, id uuid.UUID) (*Site, error) {
	var s Site
	err := pool.QueryRow(ctx, `
		SELECT id, name, address, latitude, longitude, allowed_radius_m, is_active
		FROM ojt_sites WHERE id = $1`, id).
		Scan(&s.ID, &s.Name, &s.Address, &s.Latitude, &s.Longitude, &s.AllowedRadiusM, &s.IsActive)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, httpx.ErrNotFound
		}
		return nil, httpx.Internal(err)
	}
	return &s, nil
}

func Create(ctx context.Context, pool *db.Pool, in Input, createdBy uuid.UUID) (*Site, error) {
	if in.Name == nil || in.Address == nil || in.Latitude == nil || in.Longitude == nil || in.AllowedRadiusM == nil {
		return nil, httpx.ValidationMsg("name, address, latitude, longitude, and allowed_radius_m are required.")
	}
	if err := validate(*in.Name, *in.Address, *in.Latitude, *in.Longitude, *in.AllowedRadiusM); err != nil {
		return nil, err
	}
	active := true
	if in.IsActive != nil {
		active = *in.IsActive
	}
	var s Site
	err := pool.QueryRow(ctx, `
		INSERT INTO ojt_sites (name, address, latitude, longitude, allowed_radius_m, is_active, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, name, address, latitude, longitude, allowed_radius_m, is_active`,
		strings.TrimSpace(*in.Name), strings.TrimSpace(*in.Address), *in.Latitude, *in.Longitude,
		*in.AllowedRadiusM, active, createdBy).
		Scan(&s.ID, &s.Name, &s.Address, &s.Latitude, &s.Longitude, &s.AllowedRadiusM, &s.IsActive)
	if err != nil {
		return nil, httpx.Internal(err)
	}
	return &s, nil
}

func Update(ctx context.Context, pool *db.Pool, id uuid.UUID, in Input) (*Site, error) {
	cur, err := Get(ctx, pool, id)
	if err != nil {
		return nil, err
	}
	name, address := cur.Name, cur.Address
	lat, lng, radius, active := cur.Latitude, cur.Longitude, cur.AllowedRadiusM, cur.IsActive
	if in.Name != nil {
		name = *in.Name
	}
	if in.Address != nil {
		address = *in.Address
	}
	if in.Latitude != nil {
		lat = *in.Latitude
	}
	if in.Longitude != nil {
		lng = *in.Longitude
	}
	if in.AllowedRadiusM != nil {
		radius = *in.AllowedRadiusM
	}
	if in.IsActive != nil {
		active = *in.IsActive
	}
	if err := validate(name, address, lat, lng, radius); err != nil {
		return nil, err
	}
	var s Site
	err = pool.QueryRow(ctx, `
		UPDATE ojt_sites SET name=$2, address=$3, latitude=$4, longitude=$5,
			allowed_radius_m=$6, is_active=$7, updated_at=now()
		WHERE id=$1
		RETURNING id, name, address, latitude, longitude, allowed_radius_m, is_active`,
		id, strings.TrimSpace(name), strings.TrimSpace(address), lat, lng, radius, active).
		Scan(&s.ID, &s.Name, &s.Address, &s.Latitude, &s.Longitude, &s.AllowedRadiusM, &s.IsActive)
	if err != nil {
		return nil, httpx.Internal(err)
	}
	return &s, nil
}
