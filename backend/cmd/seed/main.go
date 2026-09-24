package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/skycode/ojt-management/backend/internal/auth"
)

func main() {
	ctx := context.Background()
	pool, _ := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	hash, _ := auth.HashPassword("Trainee123!")
	var id string
	err := pool.QueryRow(ctx, `
		INSERT INTO users (email, display_name, password_hash, role, account_status)
		VALUES ('trainee1@ojt.local','Trainee One',$1,'trainee','active')
		ON CONFLICT (lower(email)) DO UPDATE SET password_hash=$1
		RETURNING id`, hash).Scan(&id)
	fmt.Println("trainee user:", id, err)
	var tid string
	err = pool.QueryRow(ctx, `
		INSERT INTO trainee_profiles (user_id, student_number, program, year_level)
		VALUES ($1,'2024-0001','BSIT','4')
		ON CONFLICT (user_id) DO UPDATE SET student_number='2024-0001'
		RETURNING id`, id).Scan(&tid)
	fmt.Println("trainee profile:", tid, err)
	var sid string
	err = pool.QueryRow(ctx, `SELECT id FROM ojt_sites LIMIT 1`).Scan(&sid)
	if sid == "" {
		err = pool.QueryRow(ctx, `
			INSERT INTO ojt_sites (name,address,latitude,longitude,allowed_radius_m)
			VALUES ('ABC Technologies','Isulan',6.629,124.605,150) RETURNING id`).Scan(&sid)
	}
	fmt.Println("site:", sid, err)
	var aid string
	err = pool.QueryRow(ctx, `SELECT id FROM ojt_assignments WHERE trainee_id=$1 LIMIT 1`, tid).Scan(&aid)
	if aid == "" {
		err = pool.QueryRow(ctx, `
			INSERT INTO ojt_assignments (trainee_id,site_id,start_date,required_minutes,expected_weekdays,status)
			VALUES ($1,$2,CURRENT_DATE - 7,480*60,'{1,2,3,4,5,6,7}','active') RETURNING id`, tid, sid).Scan(&aid)
	}
	fmt.Println("assignment:", aid, err)
}
