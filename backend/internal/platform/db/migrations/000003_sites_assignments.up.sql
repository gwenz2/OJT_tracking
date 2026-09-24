-- Foundation: ojt_sites + ojt_assignments (data-model.md §3.5, §3.6)
CREATE TABLE ojt_sites (
    id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name             text NOT NULL,
    address          text NOT NULL,
    latitude         double precision NOT NULL CHECK (latitude BETWEEN -90 AND 90),
    longitude        double precision NOT NULL CHECK (longitude BETWEEN -180 AND 180),
    allowed_radius_m integer NOT NULL CHECK (allowed_radius_m > 0),
    is_active        boolean NOT NULL DEFAULT true,
    created_by       uuid REFERENCES users(id),
    created_at       timestamptz NOT NULL DEFAULT now(),
    updated_at       timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE ojt_assignments (
    id                      uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    trainee_id              uuid NOT NULL REFERENCES trainee_profiles(id),
    site_id                 uuid NOT NULL REFERENCES ojt_sites(id),
    start_date              date NOT NULL,
    end_date                date,
    required_minutes        integer NOT NULL CHECK (required_minutes > 0),
    break_rule_type         text NOT NULL DEFAULT 'none'
        CHECK (break_rule_type IN ('none', 'fixed_after_threshold')),
    break_threshold_minutes integer CHECK (break_threshold_minutes IS NULL OR break_threshold_minutes >= 0),
    break_deduction_minutes integer NOT NULL DEFAULT 0 CHECK (break_deduction_minutes >= 0),
    -- every element must be 1..7 (Mon..Sun); containment into the allowed set
    expected_weekdays       smallint[] CHECK (
        expected_weekdays IS NULL OR expected_weekdays <@ '{1,2,3,4,5,6,7}'::smallint[]
    ),
    status      text NOT NULL DEFAULT 'planned'
        CHECK (status IN ('planned', 'active', 'completed', 'suspended', 'cancelled')),
    created_by  uuid REFERENCES users(id),
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    CHECK (end_date IS NULL OR end_date >= start_date),
    CHECK (break_rule_type = 'none' OR break_threshold_minutes IS NOT NULL)
);

-- One active assignment per trainee (MVP decision).
CREATE UNIQUE INDEX ojt_assignments_one_active_per_trainee
    ON ojt_assignments (trainee_id) WHERE status = 'active';

CREATE INDEX ojt_assignments_status_dates_idx ON ojt_assignments (status, start_date, end_date);
CREATE INDEX ojt_assignments_trainee_idx ON ojt_assignments (trainee_id);
CREATE INDEX ojt_assignments_site_idx ON ojt_assignments (site_id);
