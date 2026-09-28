--
-- PostgreSQL database dump
--

\restrict qSCzibu9axTf34LROs8qHOeQrnYMAdQzaxVIrWEZa9kdWXc5bVDOMc7vkPDlhfe

-- Dumped from database version 18.6
-- Dumped by pg_dump version 18.6

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

ALTER TABLE IF EXISTS ONLY public.user_sessions DROP CONSTRAINT IF EXISTS user_sessions_user_id_fkey;
ALTER TABLE IF EXISTS ONLY public.trainee_profiles DROP CONSTRAINT IF EXISTS trainee_profiles_user_id_fkey;
ALTER TABLE IF EXISTS ONLY public.ojt_sites DROP CONSTRAINT IF EXISTS ojt_sites_created_by_fkey;
ALTER TABLE IF EXISTS ONLY public.ojt_assignments DROP CONSTRAINT IF EXISTS ojt_assignments_trainee_id_fkey;
ALTER TABLE IF EXISTS ONLY public.ojt_assignments DROP CONSTRAINT IF EXISTS ojt_assignments_site_id_fkey;
ALTER TABLE IF EXISTS ONLY public.ojt_assignments DROP CONSTRAINT IF EXISTS ojt_assignments_created_by_fkey;
ALTER TABLE IF EXISTS ONLY public.notifications DROP CONSTRAINT IF EXISTS notifications_recipient_user_id_fkey;
ALTER TABLE IF EXISTS ONLY public.journal_reviews DROP CONSTRAINT IF EXISTS journal_reviews_reviewer_user_id_fkey;
ALTER TABLE IF EXISTS ONLY public.journal_reviews DROP CONSTRAINT IF EXISTS journal_reviews_journal_id_fkey;
ALTER TABLE IF EXISTS ONLY public.journal_evidence DROP CONSTRAINT IF EXISTS journal_evidence_journal_id_fkey;
ALTER TABLE IF EXISTS ONLY public.institution_settings DROP CONSTRAINT IF EXISTS institution_settings_updated_by_fkey;
ALTER TABLE IF EXISTS ONLY public.daily_journals DROP CONSTRAINT IF EXISTS daily_journals_trainee_id_fkey;
ALTER TABLE IF EXISTS ONLY public.daily_journals DROP CONSTRAINT IF EXISTS daily_journals_attendance_id_fkey;
ALTER TABLE IF EXISTS ONLY public.correction_requests DROP CONSTRAINT IF EXISTS correction_requests_trainee_id_fkey;
ALTER TABLE IF EXISTS ONLY public.correction_requests DROP CONSTRAINT IF EXISTS correction_requests_decided_by_fkey;
ALTER TABLE IF EXISTS ONLY public.correction_requests DROP CONSTRAINT IF EXISTS correction_requests_attendance_id_fkey;
ALTER TABLE IF EXISTS ONLY public.coordinator_scopes DROP CONSTRAINT IF EXISTS coordinator_scopes_trainee_id_fkey;
ALTER TABLE IF EXISTS ONLY public.coordinator_scopes DROP CONSTRAINT IF EXISTS coordinator_scopes_coordinator_user_id_fkey;
ALTER TABLE IF EXISTS ONLY public.audit_events DROP CONSTRAINT IF EXISTS audit_events_actor_user_id_fkey;
ALTER TABLE IF EXISTS ONLY public.attendance_sessions DROP CONSTRAINT IF EXISTS attendance_sessions_trainee_id_fkey;
ALTER TABLE IF EXISTS ONLY public.attendance_sessions DROP CONSTRAINT IF EXISTS attendance_sessions_assignment_id_fkey;
ALTER TABLE IF EXISTS ONLY public.attendance_evidence DROP CONSTRAINT IF EXISTS attendance_evidence_trainee_id_fkey;
ALTER TABLE IF EXISTS ONLY public.attendance_evidence DROP CONSTRAINT IF EXISTS attendance_evidence_attendance_id_fkey;
ALTER TABLE IF EXISTS ONLY public.attendance_adjustments DROP CONSTRAINT IF EXISTS attendance_adjustments_correction_request_id_fkey;
ALTER TABLE IF EXISTS ONLY public.attendance_adjustments DROP CONSTRAINT IF EXISTS attendance_adjustments_attendance_id_fkey;
ALTER TABLE IF EXISTS ONLY public.attendance_adjustments DROP CONSTRAINT IF EXISTS attendance_adjustments_approved_by_fkey;
DROP INDEX IF EXISTS public.users_email_lower_key;
DROP INDEX IF EXISTS public.user_sessions_user_expires_idx;
DROP INDEX IF EXISTS public.user_sessions_token_hash_key;
DROP INDEX IF EXISTS public.trainee_profiles_student_number_key;
DROP INDEX IF EXISTS public.ojt_assignments_trainee_idx;
DROP INDEX IF EXISTS public.ojt_assignments_status_dates_idx;
DROP INDEX IF EXISTS public.ojt_assignments_site_idx;
DROP INDEX IF EXISTS public.ojt_assignments_one_active_per_trainee;
DROP INDEX IF EXISTS public.notifications_recipient_read_idx;
DROP INDEX IF EXISTS public.notifications_recipient_event_key;
DROP INDEX IF EXISTS public.journal_reviews_journal_idx;
DROP INDEX IF EXISTS public.journal_evidence_journal_idx;
DROP INDEX IF EXISTS public.institution_settings_singleton;
DROP INDEX IF EXISTS public.daily_journals_trainee_idx;
DROP INDEX IF EXISTS public.daily_journals_status_idx;
DROP INDEX IF EXISTS public.correction_requests_trainee_idx;
DROP INDEX IF EXISTS public.correction_requests_status_idx;
DROP INDEX IF EXISTS public.correction_requests_one_pending;
DROP INDEX IF EXISTS public.coordinator_scopes_trainee_idx;
DROP INDEX IF EXISTS public.audit_events_resource_idx;
DROP INDEX IF EXISTS public.audit_events_actor_idx;
DROP INDEX IF EXISTS public.attendance_sessions_trainee_date_key;
DROP INDEX IF EXISTS public.attendance_sessions_trainee_date_idx;
DROP INDEX IF EXISTS public.attendance_sessions_date_status_idx;
DROP INDEX IF EXISTS public.attendance_sessions_assignment_date_idx;
DROP INDEX IF EXISTS public.attendance_evidence_client_action_key;
DROP INDEX IF EXISTS public.attendance_evidence_attendance_idx;
DROP INDEX IF EXISTS public.attendance_evidence_action_key;
DROP INDEX IF EXISTS public.attendance_adjustments_attendance_idx;
ALTER TABLE IF EXISTS ONLY public.users DROP CONSTRAINT IF EXISTS users_pkey;
ALTER TABLE IF EXISTS ONLY public.user_sessions DROP CONSTRAINT IF EXISTS user_sessions_pkey;
ALTER TABLE IF EXISTS ONLY public.trainee_profiles DROP CONSTRAINT IF EXISTS trainee_profiles_user_id_key;
ALTER TABLE IF EXISTS ONLY public.trainee_profiles DROP CONSTRAINT IF EXISTS trainee_profiles_pkey;
ALTER TABLE IF EXISTS ONLY public.schema_migrations DROP CONSTRAINT IF EXISTS schema_migrations_pkey;
ALTER TABLE IF EXISTS ONLY public.ojt_sites DROP CONSTRAINT IF EXISTS ojt_sites_pkey;
ALTER TABLE IF EXISTS ONLY public.ojt_assignments DROP CONSTRAINT IF EXISTS ojt_assignments_pkey;
ALTER TABLE IF EXISTS ONLY public.notifications DROP CONSTRAINT IF EXISTS notifications_pkey;
ALTER TABLE IF EXISTS ONLY public.journal_reviews DROP CONSTRAINT IF EXISTS journal_reviews_pkey;
ALTER TABLE IF EXISTS ONLY public.journal_evidence DROP CONSTRAINT IF EXISTS journal_evidence_pkey;
ALTER TABLE IF EXISTS ONLY public.institution_settings DROP CONSTRAINT IF EXISTS institution_settings_pkey;
ALTER TABLE IF EXISTS ONLY public.daily_journals DROP CONSTRAINT IF EXISTS daily_journals_pkey;
ALTER TABLE IF EXISTS ONLY public.daily_journals DROP CONSTRAINT IF EXISTS daily_journals_attendance_id_key;
ALTER TABLE IF EXISTS ONLY public.correction_requests DROP CONSTRAINT IF EXISTS correction_requests_pkey;
ALTER TABLE IF EXISTS ONLY public.coordinator_scopes DROP CONSTRAINT IF EXISTS coordinator_scopes_pkey;
ALTER TABLE IF EXISTS ONLY public.audit_events DROP CONSTRAINT IF EXISTS audit_events_pkey;
ALTER TABLE IF EXISTS ONLY public.attendance_sessions DROP CONSTRAINT IF EXISTS attendance_sessions_pkey;
ALTER TABLE IF EXISTS ONLY public.attendance_evidence DROP CONSTRAINT IF EXISTS attendance_evidence_pkey;
ALTER TABLE IF EXISTS ONLY public.attendance_adjustments DROP CONSTRAINT IF EXISTS attendance_adjustments_pkey;
ALTER TABLE IF EXISTS ONLY public.attendance_adjustments DROP CONSTRAINT IF EXISTS attendance_adjustments_correction_request_id_key;
DROP TABLE IF EXISTS public.users;
DROP TABLE IF EXISTS public.user_sessions;
DROP TABLE IF EXISTS public.trainee_profiles;
DROP TABLE IF EXISTS public.schema_migrations;
DROP TABLE IF EXISTS public.ojt_sites;
DROP TABLE IF EXISTS public.ojt_assignments;
DROP TABLE IF EXISTS public.notifications;
DROP TABLE IF EXISTS public.journal_reviews;
DROP TABLE IF EXISTS public.journal_evidence;
DROP TABLE IF EXISTS public.institution_settings;
DROP TABLE IF EXISTS public.daily_journals;
DROP TABLE IF EXISTS public.correction_requests;
DROP TABLE IF EXISTS public.coordinator_scopes;
DROP TABLE IF EXISTS public.audit_events;
DROP TABLE IF EXISTS public.attendance_sessions;
DROP TABLE IF EXISTS public.attendance_evidence;
DROP TABLE IF EXISTS public.attendance_adjustments;
DROP EXTENSION IF EXISTS pgcrypto;
--
-- Name: pgcrypto; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS pgcrypto WITH SCHEMA public;


--
-- Name: EXTENSION pgcrypto; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION pgcrypto IS 'cryptographic functions';


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: attendance_adjustments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.attendance_adjustments (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    attendance_id uuid NOT NULL,
    correction_request_id uuid NOT NULL,
    effective_time_in_at timestamp with time zone NOT NULL,
    effective_time_out_at timestamp with time zone,
    credited_minutes integer,
    approved_by uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT attendance_adjustments_check CHECK (((effective_time_out_at IS NULL) OR (effective_time_out_at > effective_time_in_at))),
    CONSTRAINT attendance_adjustments_credited_minutes_check CHECK (((credited_minutes IS NULL) OR (credited_minutes >= 0)))
);


--
-- Name: attendance_evidence; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.attendance_evidence (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    attendance_id uuid NOT NULL,
    trainee_id uuid NOT NULL,
    action text NOT NULL,
    client_action_id uuid NOT NULL,
    server_captured_at timestamp with time zone NOT NULL,
    device_captured_at timestamp with time zone,
    device_timezone_offset_min integer,
    latitude double precision,
    longitude double precision,
    gps_accuracy_m double precision,
    distance_from_site_m double precision,
    radius_used_m integer,
    location_status text NOT NULL,
    location_exception_reason text,
    original_object_key text NOT NULL,
    watermarked_object_key text NOT NULL,
    original_sha256 text NOT NULL,
    watermarked_sha256 text NOT NULL,
    mime_type text NOT NULL,
    size_bytes bigint NOT NULL,
    width_px integer NOT NULL,
    height_px integer NOT NULL,
    watermark_text text,
    watermark_version integer DEFAULT 1 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT attendance_evidence_action_check CHECK ((action = ANY (ARRAY['time_in'::text, 'time_out'::text]))),
    CONSTRAINT attendance_evidence_check CHECK (((location_status <> 'unavailable'::text) OR (location_exception_reason IS NOT NULL))),
    CONSTRAINT attendance_evidence_distance_from_site_m_check CHECK (((distance_from_site_m IS NULL) OR (distance_from_site_m >= (0)::double precision))),
    CONSTRAINT attendance_evidence_gps_accuracy_m_check CHECK (((gps_accuracy_m IS NULL) OR (gps_accuracy_m >= (0)::double precision))),
    CONSTRAINT attendance_evidence_height_px_check CHECK ((height_px > 0)),
    CONSTRAINT attendance_evidence_location_status_check CHECK ((location_status = ANY (ARRAY['verified'::text, 'outside_radius'::text, 'low_accuracy'::text, 'unavailable'::text]))),
    CONSTRAINT attendance_evidence_size_bytes_check CHECK ((size_bytes > 0)),
    CONSTRAINT attendance_evidence_width_px_check CHECK ((width_px > 0))
);


--
-- Name: attendance_sessions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.attendance_sessions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    trainee_id uuid NOT NULL,
    assignment_id uuid NOT NULL,
    attendance_date date NOT NULL,
    original_time_in_at timestamp with time zone NOT NULL,
    original_time_out_at timestamp with time zone,
    effective_time_in_at timestamp with time zone NOT NULL,
    effective_time_out_at timestamp with time zone,
    credited_minutes integer,
    status text DEFAULT 'open'::text NOT NULL,
    flag_codes text[] DEFAULT '{}'::text[] NOT NULL,
    version integer DEFAULT 1 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    closed_at timestamp with time zone,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT attendance_sessions_check CHECK (((original_time_out_at IS NULL) OR (original_time_out_at > original_time_in_at))),
    CONSTRAINT attendance_sessions_check1 CHECK (((effective_time_out_at IS NULL) OR (effective_time_out_at > effective_time_in_at))),
    CONSTRAINT attendance_sessions_credited_minutes_check CHECK (((credited_minutes IS NULL) OR (credited_minutes >= 0))),
    CONSTRAINT attendance_sessions_status_check CHECK ((status = ANY (ARRAY['open'::text, 'valid'::text, 'flagged'::text, 'corrected'::text])))
);


--
-- Name: audit_events; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.audit_events (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    actor_user_id uuid,
    event_type text NOT NULL,
    resource_type text NOT NULL,
    resource_id uuid,
    request_id text,
    metadata jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: coordinator_scopes; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.coordinator_scopes (
    coordinator_user_id uuid NOT NULL,
    trainee_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: correction_requests; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.correction_requests (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    attendance_id uuid NOT NULL,
    trainee_id uuid NOT NULL,
    type text NOT NULL,
    reason text NOT NULL,
    proposed_time_in_at timestamp with time zone,
    proposed_time_out_at timestamp with time zone,
    status text DEFAULT 'pending'::text NOT NULL,
    requested_at timestamp with time zone DEFAULT now() NOT NULL,
    decided_by uuid,
    decision_comment text,
    decided_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT correction_requests_status_check CHECK ((status = ANY (ARRAY['pending'::text, 'approved'::text, 'rejected'::text, 'cancelled'::text]))),
    CONSTRAINT correction_requests_type_check CHECK ((type = ANY (ARRAY['missed_time_out'::text, 'incorrect_time'::text, 'field_assignment'::text, 'gps_issue'::text, 'other'::text])))
);


--
-- Name: daily_journals; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.daily_journals (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    attendance_id uuid NOT NULL,
    trainee_id uuid NOT NULL,
    narrative text,
    status text DEFAULT 'draft'::text NOT NULL,
    submitted_at timestamp with time zone,
    reviewed_at timestamp with time zone,
    revision_count integer DEFAULT 0 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT daily_journals_status_check CHECK ((status = ANY (ARRAY['draft'::text, 'submitted'::text, 'reviewed'::text, 'needs_revision'::text])))
);


--
-- Name: institution_settings; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.institution_settings (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    timezone text DEFAULT 'Asia/Manila'::text NOT NULL,
    low_accuracy_threshold_m integer DEFAULT 100 NOT NULL,
    near_completion_percent numeric(5,2) DEFAULT 90 NOT NULL,
    missing_journal_cutoff_hours integer DEFAULT 24 NOT NULL,
    unusual_session_minutes integer DEFAULT 960 NOT NULL,
    retention_days integer,
    updated_by uuid,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT institution_settings_low_accuracy_threshold_m_check CHECK ((low_accuracy_threshold_m > 0)),
    CONSTRAINT institution_settings_missing_journal_cutoff_hours_check CHECK ((missing_journal_cutoff_hours >= 0)),
    CONSTRAINT institution_settings_near_completion_percent_check CHECK (((near_completion_percent >= (0)::numeric) AND (near_completion_percent <= (100)::numeric))),
    CONSTRAINT institution_settings_retention_days_check CHECK (((retention_days IS NULL) OR (retention_days > 0))),
    CONSTRAINT institution_settings_unusual_session_minutes_check CHECK ((unusual_session_minutes > 0))
);


--
-- Name: journal_evidence; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.journal_evidence (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    journal_id uuid NOT NULL,
    object_key text NOT NULL,
    sha256 text NOT NULL,
    mime_type text NOT NULL,
    size_bytes bigint NOT NULL,
    width_px integer NOT NULL,
    height_px integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT journal_evidence_height_px_check CHECK ((height_px > 0)),
    CONSTRAINT journal_evidence_size_bytes_check CHECK ((size_bytes > 0)),
    CONSTRAINT journal_evidence_width_px_check CHECK ((width_px > 0))
);


--
-- Name: journal_reviews; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.journal_reviews (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    journal_id uuid NOT NULL,
    reviewer_user_id uuid NOT NULL,
    decision text NOT NULL,
    comment text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT journal_reviews_check CHECK (((decision <> 'needs_revision'::text) OR (comment IS NOT NULL))),
    CONSTRAINT journal_reviews_decision_check CHECK ((decision = ANY (ARRAY['reviewed'::text, 'needs_revision'::text])))
);


--
-- Name: notifications; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.notifications (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    recipient_user_id uuid NOT NULL,
    event_key text NOT NULL,
    type text NOT NULL,
    title text NOT NULL,
    body text NOT NULL,
    resource_type text,
    resource_id uuid,
    read_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: ojt_assignments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.ojt_assignments (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    trainee_id uuid NOT NULL,
    site_id uuid NOT NULL,
    start_date date NOT NULL,
    end_date date,
    required_minutes integer NOT NULL,
    break_rule_type text DEFAULT 'none'::text NOT NULL,
    break_threshold_minutes integer,
    break_deduction_minutes integer DEFAULT 0 NOT NULL,
    expected_weekdays smallint[],
    status text DEFAULT 'planned'::text NOT NULL,
    created_by uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT ojt_assignments_break_deduction_minutes_check CHECK ((break_deduction_minutes >= 0)),
    CONSTRAINT ojt_assignments_break_rule_type_check CHECK ((break_rule_type = ANY (ARRAY['none'::text, 'fixed_after_threshold'::text]))),
    CONSTRAINT ojt_assignments_break_threshold_minutes_check CHECK (((break_threshold_minutes IS NULL) OR (break_threshold_minutes >= 0))),
    CONSTRAINT ojt_assignments_check CHECK (((end_date IS NULL) OR (end_date >= start_date))),
    CONSTRAINT ojt_assignments_check1 CHECK (((break_rule_type = 'none'::text) OR (break_threshold_minutes IS NOT NULL))),
    CONSTRAINT ojt_assignments_expected_weekdays_check CHECK (((expected_weekdays IS NULL) OR (expected_weekdays <@ '{1,2,3,4,5,6,7}'::smallint[]))),
    CONSTRAINT ojt_assignments_required_minutes_check CHECK ((required_minutes > 0)),
    CONSTRAINT ojt_assignments_status_check CHECK ((status = ANY (ARRAY['planned'::text, 'active'::text, 'completed'::text, 'suspended'::text, 'cancelled'::text])))
);


--
-- Name: ojt_sites; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.ojt_sites (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name text NOT NULL,
    address text NOT NULL,
    latitude double precision NOT NULL,
    longitude double precision NOT NULL,
    allowed_radius_m integer NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    created_by uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT ojt_sites_allowed_radius_m_check CHECK ((allowed_radius_m > 0)),
    CONSTRAINT ojt_sites_latitude_check CHECK (((latitude >= ('-90'::integer)::double precision) AND (latitude <= (90)::double precision))),
    CONSTRAINT ojt_sites_longitude_check CHECK (((longitude >= ('-180'::integer)::double precision) AND (longitude <= (180)::double precision)))
);


--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.schema_migrations (
    version text NOT NULL,
    applied_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: trainee_profiles; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.trainee_profiles (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    student_number text NOT NULL,
    program text,
    year_level text,
    contact_number text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: user_sessions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_sessions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    token_hash text NOT NULL,
    csrf_secret_hash text,
    user_agent_hash text,
    ip_prefix inet,
    expires_at timestamp with time zone NOT NULL,
    revoked_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    last_seen_at timestamp with time zone
);


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    email text NOT NULL,
    password_hash text NOT NULL,
    role text NOT NULL,
    account_status text DEFAULT 'active'::text NOT NULL,
    display_name text NOT NULL,
    last_login_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT users_account_status_check CHECK ((account_status = ANY (ARRAY['active'::text, 'inactive'::text, 'locked'::text]))),
    CONSTRAINT users_role_check CHECK ((role = ANY (ARRAY['trainee'::text, 'coordinator'::text, 'admin'::text])))
);


--
-- Data for Name: attendance_adjustments; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.attendance_adjustments (id, attendance_id, correction_request_id, effective_time_in_at, effective_time_out_at, credited_minutes, approved_by, created_at) FROM stdin;
ca4e2fe8-a8c5-4b28-a4e2-b425bc66590d	f490a0f2-e2b3-4933-9282-5bf10e48779f	f83913b5-ba0c-46a4-a007-07c719287233	2026-09-25 01:39:39.420707+00	\N	\N	2f131664-4d97-40af-965b-e0078b1fa9b3	2026-09-25 01:41:06.814038+00
\.


--
-- Data for Name: attendance_evidence; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.attendance_evidence (id, attendance_id, trainee_id, action, client_action_id, server_captured_at, device_captured_at, device_timezone_offset_min, latitude, longitude, gps_accuracy_m, distance_from_site_m, radius_used_m, location_status, location_exception_reason, original_object_key, watermarked_object_key, original_sha256, watermarked_sha256, mime_type, size_bytes, width_px, height_px, watermark_text, watermark_version, created_at) FROM stdin;
0a4991da-4a88-4e37-b606-fcec707e11af	f490a0f2-e2b3-4933-9282-5bf10e48779f	0a534bd5-c192-43bd-a224-c6a7158e2ca2	time_in	fe755d9c-0b54-41b0-8b55-bd632304a3c9	2026-09-25 01:39:39.420707+00	2026-09-25 01:39:36.876+00	480	6.6265692710876465	124.59123229980469	200	1544.500832246913	150	low_accuracy	\N	attendance/original/8c9df09a-8de7-4ca7-a53a-cc7536e9f642.jpg	attendance/watermarked/ccd5be3b-8a6b-4155-bbd0-99f68dfebe43.jpg	d31e0ba6cacf4f74db5a1c5d41f2134653a0f7cc9073c62094a7c96a1ff88921	497e38fc52235510a3296a8e2a9dc34066b2be8a4aeaa064c13471be28efbc20	image/jpeg	318527	1280	720	Test 12 (TEST-001) — Time In / 2026-09-25 09:39:39 Asia/Manila @ ABC Technologies · GPS: low_accuracy	1	2026-09-25 01:39:39.73618+00
2524f5a7-859e-4cdb-9154-029a34653791	0eb70f52-7a87-459f-96de-ac917ac1efc3	c0377216-d3d8-4d68-82c0-1594f17ead44	time_in	32b6d68f-27e1-473f-9329-9dc9516b88eb	2026-09-25 02:48:08.708073+00	2026-09-25 02:48:06.537+00	480	6.619433	124.596306	78	1433.1093109411252	150	outside_radius	\N	attendance/original/b51d575c-559b-4ace-88a9-3ab7c3410af3.jpg	attendance/watermarked/4415b557-f3ba-4204-a4cc-e33be30d3c6b.jpg	9cd5dbbf73d7dca6226c159e567ccc50d06d4a7e7a6333ad5aa34afec6b95202	944d3e170c956c18dfb1b552e0caf371481853e9e825e427f7c5ac63f9fc421a	image/jpeg	322231	1280	720	Trainee One (2024-0001) — Time In / 2026-09-25 10:48:08 Asia/Manila @ ABC Technologies · GPS: outside_radius	1	2026-09-25 02:48:08.910219+00
e2136b07-fdaa-470a-ba3f-094c227f87e6	58d883d5-d311-4af5-b19d-e351b5c67fca	9e1cfcf3-7812-4d97-8f91-6a925ee4178c	time_in	ed48861b-c6ff-4a76-882d-d1248075eace	2026-09-25 03:11:52.856709+00	2026-09-25 03:11:50.931+00	480	6.619433	124.596306	78	1433.1093109411252	150	outside_radius	\N	attendance/original/529b2c62-4deb-4703-9806-47856eac5a8b.jpg	attendance/watermarked/e798b33d-ff96-4011-85ef-85918064d330.jpg	9bdc2c16c496c1f02fe3e84ad3c437af8f4cac1c0a9552f3bd5794ec9558b51d	083bb51701cc7cebea1c31e6d917fef0e9033728c1a3c2bad39d329260c6606f	image/jpeg	342221	1280	720	Test 13 (TEST-002) — Time In / 2026-09-25 11:11:52 Asia/Manila @ ABC Technologies · GPS: outside_radius	1	2026-09-25 03:11:53.094332+00
aedecc88-6527-4514-b62d-6c02fd4fdfe8	58d883d5-d311-4af5-b19d-e351b5c67fca	9e1cfcf3-7812-4d97-8f91-6a925ee4178c	time_out	d433a5cf-7641-4ebe-a087-3285fdec9b14	2026-09-25 04:03:03.770393+00	2026-09-25 04:02:58.931+00	480	6.619473833333333	124.59635666666667	100	1425.9894639412323	150	outside_radius	\N	attendance/original/72d2ffa4-31ff-4b73-a4d5-2e3da3894f14.jpg	attendance/watermarked/c8f31eeb-f371-4d6b-a5b9-3d4ea18bfeac.jpg	d5291b07c3ea9c22cc657893d1168917fb9ef16781b7b34091d8fd114141436d	5e6f4a5b6a8ec76495b3d9f92b6c31b58b338b583cf536a65017f85a9e1d93e5	image/jpeg	332104	1280	720	Test 13 (TEST-002) — Time Out / 2026-09-25 12:03:03 Asia/Manila @ ABC Technologies · GPS: outside_radius	1	2026-09-25 04:03:04.032293+00
1f097410-43d8-4344-ab1d-24ca2055c3e9	0eb70f52-7a87-459f-96de-ac917ac1efc3	c0377216-d3d8-4d68-82c0-1594f17ead44	time_out	0be32705-d926-4c76-ab6f-e0565d58d027	2026-09-25 06:16:34.394855+00	2026-09-25 06:16:32.098+00	480	6.619433	124.596306	78	100.02312405458964	200	verified	\N	attendance/original/ecda0890-0107-46ca-ab90-f0b7ed85a829.jpg	attendance/watermarked/9e081049-a3de-44f6-8dbd-0f00237843b9.jpg	42a332570d10e9fdb03dc3674137d5350239cc7534f30c5a8f72ce3d89b70c6b	478dfcc22e9c560778242053dc724a284eadd4d226bb860ceae75641049f25a6	image/jpeg	315754	1280	720	Trainee One (2024-0001) — Time Out / 2026-09-25 14:16:34 Asia/Manila @ SkyCode · GPS: verified	1	2026-09-25 06:16:35.810306+00
\.


--
-- Data for Name: attendance_sessions; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.attendance_sessions (id, trainee_id, assignment_id, attendance_date, original_time_in_at, original_time_out_at, effective_time_in_at, effective_time_out_at, credited_minutes, status, flag_codes, version, created_at, closed_at, updated_at) FROM stdin;
f490a0f2-e2b3-4933-9282-5bf10e48779f	0a534bd5-c192-43bd-a224-c6a7158e2ca2	f5dd6760-c402-410b-876c-752c2a52c2a7	2026-09-25	2026-09-25 01:39:39.420707+00	\N	2026-09-25 01:39:39.420707+00	\N	\N	corrected	{low_accuracy}	2	2026-09-25 01:39:39.73618+00	\N	2026-09-25 01:41:06.814038+00
58d883d5-d311-4af5-b19d-e351b5c67fca	9e1cfcf3-7812-4d97-8f91-6a925ee4178c	f79bfa08-0180-444e-bc15-f063f7343d3e	2026-09-25	2026-09-25 03:11:52.856709+00	2026-09-25 04:03:03.770393+00	2026-09-25 03:11:52.856709+00	2026-09-25 04:03:03.770393+00	51	flagged	{outside_radius}	2	2026-09-25 03:11:53.094332+00	2026-09-25 04:03:03.770393+00	2026-09-25 04:03:04.032293+00
0eb70f52-7a87-459f-96de-ac917ac1efc3	c0377216-d3d8-4d68-82c0-1594f17ead44	1b585b1f-befb-491f-9891-10b43d24b53b	2026-09-25	2026-09-25 02:48:08.708073+00	2026-09-25 06:16:34.394855+00	2026-09-25 02:48:08.708073+00	2026-09-25 06:16:34.394855+00	208	flagged	{outside_radius}	2	2026-09-25 02:48:08.910219+00	2026-09-25 06:16:34.394855+00	2026-09-25 06:16:35.810306+00
\.


--
-- Data for Name: audit_events; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.audit_events (id, actor_user_id, event_type, resource_type, resource_id, request_id, metadata, created_at) FROM stdin;
fae2cbf0-ca60-4860-a4af-89b7c2a6d6b0	\N	auth.login_failed	user	\N	req_68c55f42288b34fcd1d39078	{}	2026-09-25 01:24:15.742819+00
f6ddf4a0-e752-419d-a2ec-2aab0950ffd2	\N	auth.login_failed	user	\N	req_b931ac23745a20687fa3e533	{}	2026-09-25 01:24:26.131613+00
5e61862a-a84e-4156-8b7f-14065b5ee158	\N	auth.login_failed	user	\N	req_1dbc36cd409cd7d546b391f2	{}	2026-09-25 01:24:36.485565+00
f136123e-8387-4ab0-b875-0f41a302ed17	\N	auth.login_failed	user	\N	req_d3beb65efbb9ccf276d81bc8	{}	2026-09-25 01:24:37.343358+00
76e01a3c-eac1-4ce8-a455-7ef69c2375e8	\N	auth.login_failed	user	\N	req_11cc88351bb2d629b6c8ac9e	{}	2026-09-25 01:24:37.794808+00
7ac5e0a4-7baa-461d-8624-8b97472c4d6f	\N	auth.login_failed	user	\N	req_3eeb64e37713e9ca4cdcc9f5	{}	2026-09-25 01:24:38.005114+00
e27bc7bb-a59b-4cd5-9cb6-66a72d8ac191	\N	auth.login_failed	user	\N	req_225fe63c37460b22015c15ee	{}	2026-09-25 01:24:38.17144+00
2966d075-eb45-46fb-957d-7c85186310a4	\N	auth.login_failed	user	\N	req_eb1f9073b08aac180b02ce9b	{}	2026-09-25 01:24:38.385801+00
1b71c396-df24-46ce-96a1-df90878d26d3	2f131664-4d97-40af-965b-e0078b1fa9b3	auth.login	user	2f131664-4d97-40af-965b-e0078b1fa9b3	req_ab97e98349ba87efff0859db	{}	2026-09-25 01:26:50.912577+00
d874dfcc-da06-4a93-87de-fa64fba2bad0	be2bc35a-2e2a-4465-aa72-8759621fade6	auth.login	user	be2bc35a-2e2a-4465-aa72-8759621fade6	req_319ce210026acd73cfc85d73	{}	2026-09-25 01:26:52.765333+00
2732bb2d-f35f-4d24-b7ba-1c9ea194d689	\N	auth.login_failed	user	\N	req_575e09f13de6c8f38ae6ae65	{}	2026-09-25 01:26:52.911269+00
259185be-aa99-4247-980f-75bc60e8e1b6	\N	auth.login_failed	user	\N	req_f44c3c9de054d8cb327bb8c4	{}	2026-09-25 01:26:53.189704+00
433ba4e3-1da3-4de3-b6e2-354f8c38d9cb	2f131664-4d97-40af-965b-e0078b1fa9b3	auth.login	user	2f131664-4d97-40af-965b-e0078b1fa9b3	req_74d727bf8c2138070ee52cf5	{}	2026-09-25 01:30:08.580454+00
39c99db9-4aef-41f7-a6a6-8ab4aee48a24	2f131664-4d97-40af-965b-e0078b1fa9b3	trainee.imported	trainee_profile	\N	req_33e5d6702e9a463b31df27f3	{"created": 2}	2026-09-25 01:30:11.474385+00
4d3007b0-aaca-48ee-b0f9-4cc17ce9d929	71d44ce9-a767-45f0-8d6d-650d598f9d4b	auth.login	user	71d44ce9-a767-45f0-8d6d-650d598f9d4b	req_457f561dd2747db265867ffc	{}	2026-09-25 01:30:39.32335+00
5d9b099b-67bf-4acf-869e-7e429eec2f2a	786f62a0-39d6-40f9-a3d9-ba750c34d85e	auth.login	user	786f62a0-39d6-40f9-a3d9-ba750c34d85e	req_638fc60b97fd66a9de03ccac	{}	2026-09-25 01:30:39.588663+00
b01686aa-477a-49fc-ac7f-2165e19017c6	71d44ce9-a767-45f0-8d6d-650d598f9d4b	auth.login	user	71d44ce9-a767-45f0-8d6d-650d598f9d4b	req_d3d119d154d78cc4a3edb12e	{}	2026-09-25 01:31:31.063028+00
267f5888-97ea-487a-8ea2-602ca1a6e361	71d44ce9-a767-45f0-8d6d-650d598f9d4b	auth.logout	user	71d44ce9-a767-45f0-8d6d-650d598f9d4b	req_d748147bb7d02a0b0567953c	{}	2026-09-25 01:31:49.14828+00
b2c8eee9-343e-42f7-9683-074e0bdbaa7a	786f62a0-39d6-40f9-a3d9-ba750c34d85e	auth.login	user	786f62a0-39d6-40f9-a3d9-ba750c34d85e	req_bfc6b69bfd61faf92dba88d0	{}	2026-09-25 01:32:06.380174+00
a2e79888-d04b-4756-8f88-e2803ff0225d	786f62a0-39d6-40f9-a3d9-ba750c34d85e	auth.logout	user	786f62a0-39d6-40f9-a3d9-ba750c34d85e	req_863d267a3fae15fe379cba13	{}	2026-09-25 01:32:54.442864+00
b954aea2-9107-4693-8be8-49175a7ba069	\N	auth.login_failed	user	\N	req_10faa1878206c03f9534ca72	{}	2026-09-25 01:33:04.87065+00
31aca545-1acf-485e-bfa4-983229a97207	2f131664-4d97-40af-965b-e0078b1fa9b3	auth.login	user	2f131664-4d97-40af-965b-e0078b1fa9b3	req_27d387562d07bf5663df32ae	{}	2026-09-25 01:33:21.114484+00
0221f324-f91c-427b-a9d3-185710d9a371	2f131664-4d97-40af-965b-e0078b1fa9b3	assignment.created	ojt_assignment	f5dd6760-c402-410b-876c-752c2a52c2a7	req_37729e3e30c4cd3ecf9e1f74	{"site_id": "01b2c820-2dc9-4033-b352-8b028323cf73", "trainee_id": "0a534bd5-c192-43bd-a224-c6a7158e2ca2"}	2026-09-25 01:35:25.145178+00
7c7da27b-8968-4d73-ae79-d5a674d25fd0	2f131664-4d97-40af-965b-e0078b1fa9b3	assignment.created	ojt_assignment	f79bfa08-0180-444e-bc15-f063f7343d3e	req_f8f3e9748e8dde070f00106b	{"site_id": "01b2c820-2dc9-4033-b352-8b028323cf73", "trainee_id": "9e1cfcf3-7812-4d97-8f91-6a925ee4178c"}	2026-09-25 01:35:41.552837+00
c4ac7999-23b7-4326-969f-17a744656a42	2f131664-4d97-40af-965b-e0078b1fa9b3	auth.logout	user	2f131664-4d97-40af-965b-e0078b1fa9b3	req_0457d0a2a52fe9168a6412ab	{}	2026-09-25 01:39:01.896761+00
f7bb997f-b3e7-403f-9506-29da37c2a546	71d44ce9-a767-45f0-8d6d-650d598f9d4b	auth.login	user	71d44ce9-a767-45f0-8d6d-650d598f9d4b	req_3314f331f3f384864730abd2	{}	2026-09-25 01:39:13.258246+00
5d7ba34f-50b0-46e9-9634-cea5de0811e7	71d44ce9-a767-45f0-8d6d-650d598f9d4b	attendance.time_in	attendance_session	f490a0f2-e2b3-4933-9282-5bf10e48779f		{"location_status": "low_accuracy"}	2026-09-25 01:39:39.73618+00
53462970-2dda-416e-9844-d07cd740b9ee	71d44ce9-a767-45f0-8d6d-650d598f9d4b	correction.requested	correction_request	f83913b5-ba0c-46a4-a007-07c719287233		{"type": "gps_issue"}	2026-09-25 01:40:03.947883+00
679f7c0d-f000-42c9-8876-a4e46aa4ca01	71d44ce9-a767-45f0-8d6d-650d598f9d4b	auth.logout	user	71d44ce9-a767-45f0-8d6d-650d598f9d4b	req_f003ecba7878dda944977a1e	{}	2026-09-25 01:40:16.522898+00
0164bd72-9c6d-4091-ac48-58c0c2964b53	2f131664-4d97-40af-965b-e0078b1fa9b3	auth.login	user	2f131664-4d97-40af-965b-e0078b1fa9b3	req_af226e150f5eefacb88691e2	{}	2026-09-25 01:40:49.932716+00
bb0052ab-a193-44c6-bb6a-3c143cb5ac40	2f131664-4d97-40af-965b-e0078b1fa9b3	correction.approved	correction_request	f83913b5-ba0c-46a4-a007-07c719287233		{}	2026-09-25 01:41:06.814038+00
68097fd0-bd9c-413b-a086-c7dbf84ccfce	2f131664-4d97-40af-965b-e0078b1fa9b3	coordinator.created	user	be8659af-d267-4217-8bdf-6573794960f8	req_ab413dc3c1dbce113b01d312	{}	2026-09-25 01:41:32.96243+00
479c69ff-80e9-4ce4-8600-5ea22894444f	2f131664-4d97-40af-965b-e0078b1fa9b3	coordinator.scope_updated	user	be8659af-d267-4217-8bdf-6573794960f8	req_f6907ea4aaab317d91001f21	{"trainee_count": 3}	2026-09-25 01:42:00.129244+00
c8c3b17f-26dc-404c-80d4-f21ee761cff4	2f131664-4d97-40af-965b-e0078b1fa9b3	auth.logout	user	2f131664-4d97-40af-965b-e0078b1fa9b3	req_cec0cdbf161eeedd8fa9bd3a	{}	2026-09-25 01:42:23.969252+00
5897dad2-d383-48ec-ade5-a62853d04f0d	be8659af-d267-4217-8bdf-6573794960f8	auth.login	user	be8659af-d267-4217-8bdf-6573794960f8	req_274f52a82de0cd0ad8c2c8a7	{}	2026-09-25 01:42:36.258979+00
9b49c65f-4cd6-423b-9570-cd55de59085d	be8659af-d267-4217-8bdf-6573794960f8	auth.logout	user	be8659af-d267-4217-8bdf-6573794960f8	req_661037a774b5ee42506f5cc4	{}	2026-09-25 01:43:18.907372+00
00d73225-1625-4e63-8d53-760c051609c7	\N	auth.login_failed	user	\N	req_2e73c3e02bf05306cb24b735	{}	2026-09-25 01:43:34.662787+00
2702ae58-c8cd-44eb-99ac-db2e672a7ed0	2f131664-4d97-40af-965b-e0078b1fa9b3	auth.login	user	2f131664-4d97-40af-965b-e0078b1fa9b3	req_dbe0dea4a9a791327eea0359	{}	2026-09-25 01:43:48.308131+00
af108416-010f-4fe9-bee4-be054f56f26e	\N	auth.login_failed	user	\N	req_eb40423d518af16ad1dd192f	{}	2026-09-25 02:13:35.678477+00
641f550d-defd-4280-a09c-78d89fdff380	2f131664-4d97-40af-965b-e0078b1fa9b3	auth.login	user	2f131664-4d97-40af-965b-e0078b1fa9b3	req_5f95b045b2714c4a21122d1a	{}	2026-09-25 02:13:56.476017+00
fb288fec-e9f3-4dc5-bc78-06a0054bfe61	2f131664-4d97-40af-965b-e0078b1fa9b3	auth.logout	user	2f131664-4d97-40af-965b-e0078b1fa9b3	req_011de04c07eb48f50f48bfdb	{}	2026-09-25 02:14:04.753106+00
e8cd7762-f1cd-42c8-be3f-81ee314abe1a	be8659af-d267-4217-8bdf-6573794960f8	auth.login	user	be8659af-d267-4217-8bdf-6573794960f8	req_c30f66d037f87ce3e5651e26	{}	2026-09-25 02:23:01.298126+00
9717ae18-24f3-4f01-abc7-285480d241ae	2f131664-4d97-40af-965b-e0078b1fa9b3	auth.logout	user	2f131664-4d97-40af-965b-e0078b1fa9b3	req_db87cb136ad0e7dc29d35cc5	{}	2026-09-25 02:47:29.509583+00
869c3a4d-d967-4b22-bc96-f62128d5a8ed	be2bc35a-2e2a-4465-aa72-8759621fade6	auth.login	user	be2bc35a-2e2a-4465-aa72-8759621fade6	req_c60302b1d5cca9ad6d9a9daf	{}	2026-09-25 02:47:37.866544+00
c62440d9-852a-479f-9174-e1baa7a47a54	be2bc35a-2e2a-4465-aa72-8759621fade6	attendance.time_in	attendance_session	0eb70f52-7a87-459f-96de-ac917ac1efc3		{"location_status": "outside_radius"}	2026-09-25 02:48:08.910219+00
91461117-93a8-4d14-b07b-94d91f54fb03	be2bc35a-2e2a-4465-aa72-8759621fade6	auth.logout	user	be2bc35a-2e2a-4465-aa72-8759621fade6	req_7c4dbc074e074b46a860a48d	{}	2026-09-25 02:59:14.111518+00
b6bd0f2f-0fe2-4c9b-b0ba-22a7fe501528	71d44ce9-a767-45f0-8d6d-650d598f9d4b	auth.login	user	71d44ce9-a767-45f0-8d6d-650d598f9d4b	req_aa10391856ad694d64cd7ac3	{}	2026-09-25 03:10:59.489272+00
6d8a17bd-8437-4577-9d61-f23b83221733	71d44ce9-a767-45f0-8d6d-650d598f9d4b	auth.logout	user	71d44ce9-a767-45f0-8d6d-650d598f9d4b	req_8522583f2f2f68a99e35d07e	{}	2026-09-25 03:11:25.769947+00
e44f5a80-aa98-4b65-9df7-65fac1fb86d7	786f62a0-39d6-40f9-a3d9-ba750c34d85e	auth.login	user	786f62a0-39d6-40f9-a3d9-ba750c34d85e	req_a84c8c92e16a3afceaf6c559	{}	2026-09-25 03:11:35.72309+00
dd45a3a8-77f3-49dd-be72-e49799ce829d	786f62a0-39d6-40f9-a3d9-ba750c34d85e	attendance.time_in	attendance_session	58d883d5-d311-4af5-b19d-e351b5c67fca		{"location_status": "outside_radius"}	2026-09-25 03:11:53.094332+00
6e9bb7c8-d68c-43a1-8f8b-e2bfb97f3b1d	786f62a0-39d6-40f9-a3d9-ba750c34d85e	auth.logout	user	786f62a0-39d6-40f9-a3d9-ba750c34d85e	req_338250e1058efe66110efdbb	{}	2026-09-25 03:15:20.122648+00
e13cdd44-a252-46c9-b525-2f4fd963ba22	2f131664-4d97-40af-965b-e0078b1fa9b3	auth.login	user	2f131664-4d97-40af-965b-e0078b1fa9b3	req_17150e42054549e9d6d1640f	{}	2026-09-25 03:15:47.4628+00
58b0cb58-baca-4315-9acc-8f21f95bc25f	2f131664-4d97-40af-965b-e0078b1fa9b3	coordinator.scope_updated	user	be8659af-d267-4217-8bdf-6573794960f8	req_9c88c77a2067c3db21e6391b	{"trainee_count": 3}	2026-09-25 03:16:13.287602+00
2c6d3c36-aba3-4a5f-934a-b3e3f3b47a66	2f131664-4d97-40af-965b-e0078b1fa9b3	coordinator.updated	user	be8659af-d267-4217-8bdf-6573794960f8	req_61aa06091abf6f875cbbe9f0	{"account_status": "inactive"}	2026-09-25 03:29:11.530805+00
9de78e14-b643-49f2-8d51-2bb18e5c833b	2f131664-4d97-40af-965b-e0078b1fa9b3	coordinator.created	user	621742d6-7e67-4b9f-a3de-c63f12d4dfa5	req_8e81c4ffafec41a06a448034	{}	2026-09-25 03:29:33.581196+00
0d23abae-f0eb-4f99-b7f4-0fdc1696e9c2	2f131664-4d97-40af-965b-e0078b1fa9b3	auth.logout	user	2f131664-4d97-40af-965b-e0078b1fa9b3	req_0ab057e9e7b4129490bdae79	{}	2026-09-25 03:29:40.979825+00
83f44fde-16b0-4c85-932c-f52a5d8a7a44	621742d6-7e67-4b9f-a3de-c63f12d4dfa5	auth.login	user	621742d6-7e67-4b9f-a3de-c63f12d4dfa5	req_f38716ec12e1edbd27bdd737	{}	2026-09-25 03:29:47.686183+00
f8896d2e-b353-4b7b-88c9-2b428e7187a4	621742d6-7e67-4b9f-a3de-c63f12d4dfa5	auth.logout	user	621742d6-7e67-4b9f-a3de-c63f12d4dfa5	req_c57d89211f6fd932daf1279d	{}	2026-09-25 03:29:55.560656+00
9801a917-1ec7-4a92-a683-fc2030fc35bb	2f131664-4d97-40af-965b-e0078b1fa9b3	auth.login	user	2f131664-4d97-40af-965b-e0078b1fa9b3	req_497d465c0f1610fdbae73ff2	{}	2026-09-25 03:30:03.541419+00
8c06ecc6-4272-4f67-8941-3b02d4f954d4	2f131664-4d97-40af-965b-e0078b1fa9b3	coordinator.scope_updated	user	621742d6-7e67-4b9f-a3de-c63f12d4dfa5	req_725ca598574a851171fb251e	{"trainee_count": 3}	2026-09-25 03:30:12.375592+00
4b9b295c-fb48-4b65-af1e-2ed12deb14c2	2f131664-4d97-40af-965b-e0078b1fa9b3	auth.logout	user	2f131664-4d97-40af-965b-e0078b1fa9b3	req_89754913dfe085518ce4c3e1	{}	2026-09-25 03:30:14.293317+00
2888565b-534b-43dc-8a30-998c612a786f	621742d6-7e67-4b9f-a3de-c63f12d4dfa5	auth.login	user	621742d6-7e67-4b9f-a3de-c63f12d4dfa5	req_d0325cb6f03c2f7ee37513a3	{}	2026-09-25 03:30:29.695506+00
253aef4c-3ec6-4697-a184-092da2577e94	621742d6-7e67-4b9f-a3de-c63f12d4dfa5	auth.logout	user	621742d6-7e67-4b9f-a3de-c63f12d4dfa5	req_ca31b9fc21fbd0e21dbc8ae4	{}	2026-09-25 03:33:01.540651+00
46936eb7-bcfb-4822-b35f-b0c95c38f1c6	2f131664-4d97-40af-965b-e0078b1fa9b3	auth.login	user	2f131664-4d97-40af-965b-e0078b1fa9b3	req_2c2c5b3a8bd6fec24dfcf89e	{}	2026-09-25 03:33:15.770543+00
0b406fc2-f39b-49fc-8c29-58f97f54991e	2f131664-4d97-40af-965b-e0078b1fa9b3	user.password_reset	user	be2bc35a-2e2a-4465-aa72-8759621fade6	req_66250f00fab5ef3c3f9d3f1a	{}	2026-09-25 03:33:39.943052+00
3abd756c-4e8e-4d30-b390-72eb78dfef51	2f131664-4d97-40af-965b-e0078b1fa9b3	auth.logout	user	2f131664-4d97-40af-965b-e0078b1fa9b3	req_b03f8ba70c01153ea0bfa349	{}	2026-09-25 03:33:55.001672+00
1bbecf9d-2a12-4879-bae4-063b6ecc2344	be2bc35a-2e2a-4465-aa72-8759621fade6	auth.login	user	be2bc35a-2e2a-4465-aa72-8759621fade6	req_8f0760c9849de94557c6c693	{}	2026-09-25 03:34:01.017414+00
473ce4b3-4b14-4651-9452-bd6aca4d2c1e	be2bc35a-2e2a-4465-aa72-8759621fade6	auth.logout	user	be2bc35a-2e2a-4465-aa72-8759621fade6	req_a1a01ba700ac54f828ee8beb	{}	2026-09-25 03:35:40.038685+00
6727b1e4-f7ce-4e6f-be9c-0d3590c05b45	\N	auth.login_failed	user	\N	req_f3520c1cb93c74bea19b6254	{}	2026-09-25 03:42:32.702541+00
5113e0da-f140-49ea-8efa-36ff1bdb028f	\N	auth.login_failed	user	\N	req_6fff1b748b1f76ad11267faf	{}	2026-09-25 03:42:59.248721+00
c62435f6-4921-4723-bf66-cc1be2ac5529	621742d6-7e67-4b9f-a3de-c63f12d4dfa5	auth.login	user	621742d6-7e67-4b9f-a3de-c63f12d4dfa5	req_153288d199d517492fa9f5f7	{}	2026-09-25 03:43:20.619483+00
d8ea990a-86f0-42ef-acea-e3a30d032b4d	621742d6-7e67-4b9f-a3de-c63f12d4dfa5	auth.logout	user	621742d6-7e67-4b9f-a3de-c63f12d4dfa5	req_dcc1a490842d89e1d134f128	{}	2026-09-25 03:45:39.47544+00
7c1f0ab3-e0ef-4956-aa16-8ac220d25766	2f131664-4d97-40af-965b-e0078b1fa9b3	auth.login	user	2f131664-4d97-40af-965b-e0078b1fa9b3	req_91da2f2d3b21b3e4999a2e91	{}	2026-09-25 03:45:58.823734+00
81488070-8936-4175-953b-d2b343aa97b7	2f131664-4d97-40af-965b-e0078b1fa9b3	auth.password_changed	user	2f131664-4d97-40af-965b-e0078b1fa9b3	req_9affd90fca30bbf56026036d	{}	2026-09-25 03:46:45.782723+00
3130fca9-073c-4b79-b267-bde5ec822116	2f131664-4d97-40af-965b-e0078b1fa9b3	auth.logout	user	2f131664-4d97-40af-965b-e0078b1fa9b3	req_bd0e935bb57e3f6b9978d213	{}	2026-09-25 03:47:18.231458+00
48a6e094-674b-4f01-81c5-be8eca071768	2f131664-4d97-40af-965b-e0078b1fa9b3	auth.login	user	2f131664-4d97-40af-965b-e0078b1fa9b3	req_d15f4eaf9cf6423e0dca79ca	{}	2026-09-25 03:47:32.073633+00
95aca462-5609-454a-95a6-242a7e8b17fb	2f131664-4d97-40af-965b-e0078b1fa9b3	auth.logout	user	2f131664-4d97-40af-965b-e0078b1fa9b3	req_d2b992c4fb87f8e0d5816ac9	{}	2026-09-25 03:47:39.988048+00
25aa04c7-8e55-41a2-80b1-5b2ca6cbb9d1	71d44ce9-a767-45f0-8d6d-650d598f9d4b	auth.login	user	71d44ce9-a767-45f0-8d6d-650d598f9d4b	req_ad0997aa4174b74043c6ab99	{}	2026-09-25 03:48:00.141516+00
e4045e0d-c476-4b5c-a371-604655d8378d	71d44ce9-a767-45f0-8d6d-650d598f9d4b	auth.logout	user	71d44ce9-a767-45f0-8d6d-650d598f9d4b	req_6a94d1f9174ba779027f1490	{}	2026-09-25 03:56:31.992657+00
8f0e7e4f-5129-452f-877e-6197ad0fed36	\N	auth.login_failed	user	\N	req_590859d2381387366b43d3e0	{}	2026-09-25 03:56:49.127288+00
b7cb56b1-ac47-4720-8a77-82ccfffef749	\N	auth.login_failed	user	\N	req_4c439cf9ce3ba7558b7bbce1	{}	2026-09-25 03:57:02.350366+00
7eb11f1f-d641-41df-ae54-b8da01dccc22	2f131664-4d97-40af-965b-e0078b1fa9b3	auth.login	user	2f131664-4d97-40af-965b-e0078b1fa9b3	req_be6e842c501bde39a3e3404c	{}	2026-09-25 03:57:23.097285+00
1f56c66a-be6d-4f64-aff7-3ee9d078978d	2f131664-4d97-40af-965b-e0078b1fa9b3	auth.logout	user	2f131664-4d97-40af-965b-e0078b1fa9b3	req_87669d8d8a4d618fc882e434	{}	2026-09-25 03:57:55.177227+00
25eb7f98-8ca0-45c4-ad52-ed1b270830b0	71d44ce9-a767-45f0-8d6d-650d598f9d4b	auth.login	user	71d44ce9-a767-45f0-8d6d-650d598f9d4b	req_6db32f42d9b6737d8db4561f	{}	2026-09-25 03:58:03.888323+00
f3b52acc-3bc0-47ef-8bc2-29b27042797c	71d44ce9-a767-45f0-8d6d-650d598f9d4b	auth.logout	user	71d44ce9-a767-45f0-8d6d-650d598f9d4b	req_442293af0735cfb940cb0f58	{}	2026-09-25 03:58:07.508236+00
b1bedf7d-4796-4cc9-a1c1-c3244ebf4608	2f131664-4d97-40af-965b-e0078b1fa9b3	auth.login	user	2f131664-4d97-40af-965b-e0078b1fa9b3	req_3f4198cc59210320ca4204d6	{}	2026-09-25 04:00:51.06408+00
eb6ceaf4-f4a3-43ca-9d44-2f0c65ebeeea	2f131664-4d97-40af-965b-e0078b1fa9b3	auth.logout	user	2f131664-4d97-40af-965b-e0078b1fa9b3	req_1846efcf53330a9aa4dd328f	{}	2026-09-25 04:01:27.386866+00
18d97f22-1615-4408-b9ab-e59e011e1b84	71d44ce9-a767-45f0-8d6d-650d598f9d4b	auth.login	user	71d44ce9-a767-45f0-8d6d-650d598f9d4b	req_627fc2430cb52a6036873ef3	{}	2026-09-25 04:01:37.190825+00
9732977b-9f0a-4efa-a3ec-b52bf429d0da	71d44ce9-a767-45f0-8d6d-650d598f9d4b	auth.logout	user	71d44ce9-a767-45f0-8d6d-650d598f9d4b	req_d4433a7f9344121262071dc9	{}	2026-09-25 04:01:55.118321+00
755a3972-40cb-4ab2-8e9e-823336517161	786f62a0-39d6-40f9-a3d9-ba750c34d85e	auth.login	user	786f62a0-39d6-40f9-a3d9-ba750c34d85e	req_fbdb38722801048debf4edbb	{}	2026-09-25 04:02:06.966926+00
ae00e4ee-ebd3-4238-aec1-529986b2c2d1	786f62a0-39d6-40f9-a3d9-ba750c34d85e	attendance.time_out	attendance_session	58d883d5-d311-4af5-b19d-e351b5c67fca		{"flags": ["outside_radius"], "location_status": "outside_radius", "credited_minutes": 51}	2026-09-25 04:03:04.032293+00
a54fc4a7-bb3c-4b7e-b10f-4807dcb1edd9	786f62a0-39d6-40f9-a3d9-ba750c34d85e	journal.submitted	daily_journal	2b2ddf1f-a0e1-48ef-b054-b0fcf9b307a8		{}	2026-09-25 04:04:21.328009+00
38b463d9-b65e-4fdf-8f65-4b2bf1422826	786f62a0-39d6-40f9-a3d9-ba750c34d85e	auth.logout	user	786f62a0-39d6-40f9-a3d9-ba750c34d85e	req_a1be1a655fc8aa947726da70	{}	2026-09-25 04:04:36.63985+00
893f7b9f-d4b1-40d7-85a9-e7bf3721bf01	2f131664-4d97-40af-965b-e0078b1fa9b3	auth.login	user	2f131664-4d97-40af-965b-e0078b1fa9b3	req_cb32065001db71eb3ebf6029	{}	2026-09-25 04:04:49.396801+00
b3f6ad9a-04a2-4095-9a44-29cdb1318587	2f131664-4d97-40af-965b-e0078b1fa9b3	journal.reviewed	daily_journal	2b2ddf1f-a0e1-48ef-b054-b0fcf9b307a8		{}	2026-09-25 04:05:17.576529+00
b930a1e8-f052-4f8e-b971-bb29e5b02692	2f131664-4d97-40af-965b-e0078b1fa9b3	auth.logout	user	2f131664-4d97-40af-965b-e0078b1fa9b3	req_91a7c9ea59df1467100a7cf4	{}	2026-09-25 04:06:56.342779+00
8878d89e-f68d-4aa0-a67c-5b0ca773c93e	621742d6-7e67-4b9f-a3de-c63f12d4dfa5	auth.login	user	621742d6-7e67-4b9f-a3de-c63f12d4dfa5	req_871dd63825c7f6c9e5617504	{}	2026-09-25 04:07:15.182185+00
b0d48d74-1311-4256-9dab-f645db15a62a	621742d6-7e67-4b9f-a3de-c63f12d4dfa5	auth.logout	user	621742d6-7e67-4b9f-a3de-c63f12d4dfa5	req_19515f181930b098ae0efe4a	{}	2026-09-25 05:23:51.300415+00
293cfdfe-6093-4f9e-9674-14c430203e97	786f62a0-39d6-40f9-a3d9-ba750c34d85e	auth.login	user	786f62a0-39d6-40f9-a3d9-ba750c34d85e	req_ff424d02a0f781174967cdad	{}	2026-09-25 05:24:07.077892+00
dfa9198b-259b-4323-9d5c-225282657dda	786f62a0-39d6-40f9-a3d9-ba750c34d85e	auth.logout	user	786f62a0-39d6-40f9-a3d9-ba750c34d85e	req_0fe03993df3735cc51b4a183	{}	2026-09-25 05:35:17.423592+00
466c7f7c-fa3d-477f-a9f8-1893d497c3d1	2f131664-4d97-40af-965b-e0078b1fa9b3	auth.login	user	2f131664-4d97-40af-965b-e0078b1fa9b3	req_e0307ec5f73a5c3170a35862	{}	2026-09-25 05:35:33.674779+00
790e07f9-e139-4bf2-9746-5e0f5287e183	2f131664-4d97-40af-965b-e0078b1fa9b3	coordinator.password_assigned	user	be8659af-d267-4217-8bdf-6573794960f8	req_e78f98a01db53853ddb367f1	{}	2026-09-25 05:35:50.15258+00
95cdd621-556a-4c3f-89d8-32ae645f79a5	2f131664-4d97-40af-965b-e0078b1fa9b3	auth.logout	user	2f131664-4d97-40af-965b-e0078b1fa9b3	req_31ca827b049aa55ce5da1192	{}	2026-09-25 05:44:39.604871+00
9bfc4fbe-b7e1-4cb0-8120-61fd01bc48b1	621742d6-7e67-4b9f-a3de-c63f12d4dfa5	auth.login	user	621742d6-7e67-4b9f-a3de-c63f12d4dfa5	req_d3064b780ca216ea10ef0965	{}	2026-09-25 05:44:49.258063+00
94832a94-d53d-4145-91b7-935ea93b5ced	621742d6-7e67-4b9f-a3de-c63f12d4dfa5	auth.logout	user	621742d6-7e67-4b9f-a3de-c63f12d4dfa5	req_e8bca3246cf020de9ecb2719	{}	2026-09-25 06:08:04.926071+00
002a3f9f-d0ff-4b4b-ad44-127414e653ba	2f131664-4d97-40af-965b-e0078b1fa9b3	auth.login	user	2f131664-4d97-40af-965b-e0078b1fa9b3	req_73c02a24bf4f1630c3ac1357	{}	2026-09-25 06:08:15.604568+00
147f3e26-4328-481a-9fee-dd5d17f2d0a2	2f131664-4d97-40af-965b-e0078b1fa9b3	auth.logout	user	2f131664-4d97-40af-965b-e0078b1fa9b3	req_a2b3f1cfc03586ff03346c62	{}	2026-09-25 06:13:30.142546+00
43a7dd69-3f82-472a-a415-e5a05799bd0e	be2bc35a-2e2a-4465-aa72-8759621fade6	auth.login	user	be2bc35a-2e2a-4465-aa72-8759621fade6	req_41d60cfe98a35e0e24e51ca4	{}	2026-09-25 06:13:35.799905+00
71e24171-397b-4312-a43d-1a0ce6753f45	be2bc35a-2e2a-4465-aa72-8759621fade6	auth.logout	user	be2bc35a-2e2a-4465-aa72-8759621fade6	req_3f4fcc32d52e29172cae80f5	{}	2026-09-25 06:14:24.117319+00
32e44ef6-fbb7-46e9-b7d1-49a9ac3c6250	621742d6-7e67-4b9f-a3de-c63f12d4dfa5	auth.login	user	621742d6-7e67-4b9f-a3de-c63f12d4dfa5	req_0eead422550e2476a23b5388	{}	2026-09-25 06:14:35.919004+00
9a7c8be9-8123-4198-9a1b-bd9487de7384	621742d6-7e67-4b9f-a3de-c63f12d4dfa5	site.updated	ojt_site	01b2c820-2dc9-4033-b352-8b028323cf73	req_38ea02a177dd71999153cea5	{}	2026-09-25 06:15:39.954302+00
dba19f69-e5f1-4632-a99b-ecc8fe077d07	621742d6-7e67-4b9f-a3de-c63f12d4dfa5	site.updated	ojt_site	01b2c820-2dc9-4033-b352-8b028323cf73	req_7f31255bb19f4f032455c3c9	{}	2026-09-25 06:15:53.008518+00
e3cce5b6-22e9-4c0d-ba91-673b36dad700	621742d6-7e67-4b9f-a3de-c63f12d4dfa5	auth.logout	user	621742d6-7e67-4b9f-a3de-c63f12d4dfa5	req_d66cd6a989c1bb020489610e	{}	2026-09-25 06:15:56.196977+00
26bed37c-831c-41df-b903-f3871ceda3f6	be2bc35a-2e2a-4465-aa72-8759621fade6	auth.login	user	be2bc35a-2e2a-4465-aa72-8759621fade6	req_92f2616338649800a1470622	{}	2026-09-25 06:16:10.132285+00
c4429887-719e-4931-bae7-dce63beb833f	be2bc35a-2e2a-4465-aa72-8759621fade6	attendance.time_out	attendance_session	0eb70f52-7a87-459f-96de-ac917ac1efc3		{"flags": ["outside_radius"], "location_status": "verified", "credited_minutes": 208}	2026-09-25 06:16:35.810306+00
a38a8544-fa41-454f-a2b8-8dbb0b48fb9b	be2bc35a-2e2a-4465-aa72-8759621fade6	auth.logout	user	be2bc35a-2e2a-4465-aa72-8759621fade6	req_3478b37dd9b9767bf04578e8	{}	2026-09-25 06:17:18.22517+00
7e07af31-611d-4460-93de-ee4c3b4ec192	621742d6-7e67-4b9f-a3de-c63f12d4dfa5	auth.login	user	621742d6-7e67-4b9f-a3de-c63f12d4dfa5	req_e09e127253a3051d88bd7453	{}	2026-09-25 06:17:31.194344+00
\.


--
-- Data for Name: coordinator_scopes; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.coordinator_scopes (coordinator_user_id, trainee_id, created_at) FROM stdin;
be8659af-d267-4217-8bdf-6573794960f8	0a534bd5-c192-43bd-a224-c6a7158e2ca2	2026-09-25 03:16:13.278818+00
be8659af-d267-4217-8bdf-6573794960f8	c0377216-d3d8-4d68-82c0-1594f17ead44	2026-09-25 03:16:13.278818+00
be8659af-d267-4217-8bdf-6573794960f8	9e1cfcf3-7812-4d97-8f91-6a925ee4178c	2026-09-25 03:16:13.278818+00
621742d6-7e67-4b9f-a3de-c63f12d4dfa5	0a534bd5-c192-43bd-a224-c6a7158e2ca2	2026-09-25 03:30:12.361439+00
621742d6-7e67-4b9f-a3de-c63f12d4dfa5	9e1cfcf3-7812-4d97-8f91-6a925ee4178c	2026-09-25 03:30:12.361439+00
621742d6-7e67-4b9f-a3de-c63f12d4dfa5	c0377216-d3d8-4d68-82c0-1594f17ead44	2026-09-25 03:30:12.361439+00
\.


--
-- Data for Name: correction_requests; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.correction_requests (id, attendance_id, trainee_id, type, reason, proposed_time_in_at, proposed_time_out_at, status, requested_at, decided_by, decision_comment, decided_at, created_at, updated_at) FROM stdin;
f83913b5-ba0c-46a4-a007-07c719287233	f490a0f2-e2b3-4933-9282-5bf10e48779f	0a534bd5-c192-43bd-a224-c6a7158e2ca2	gps_issue	my laptop has low accuracy	\N	\N	approved	2026-09-25 01:40:03.947883+00	2f131664-4d97-40af-965b-e0078b1fa9b3	u good	2026-09-25 01:41:06.814038+00	2026-09-25 01:40:03.947883+00	2026-09-25 01:41:06.814038+00
\.


--
-- Data for Name: daily_journals; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.daily_journals (id, attendance_id, trainee_id, narrative, status, submitted_at, reviewed_at, revision_count, created_at, updated_at) FROM stdin;
2b2ddf1f-a0e1-48ef-b054-b0fcf9b307a8	58d883d5-d311-4af5-b19d-e351b5c67fca	9e1cfcf3-7812-4d97-8f91-6a925ee4178c	New pass right now\nNew pass right now	reviewed	2026-09-25 04:04:21.328009+00	2026-09-25 04:05:17.576529+00	0	2026-09-25 04:03:04.032293+00	2026-09-25 04:05:17.576529+00
cd5c804a-e34f-40e3-8361-6501ac0af7ea	0eb70f52-7a87-459f-96de-ac917ac1efc3	c0377216-d3d8-4d68-82c0-1594f17ead44	\N	draft	\N	\N	0	2026-09-25 06:16:35.810306+00	2026-09-25 06:16:35.810306+00
\.


--
-- Data for Name: institution_settings; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.institution_settings (id, timezone, low_accuracy_threshold_m, near_completion_percent, missing_journal_cutoff_hours, unusual_session_minutes, retention_days, updated_by, updated_at) FROM stdin;
f9589f7a-593d-4506-b6f4-24870558bb55	Asia/Manila	100	90.00	24	960	\N	\N	2026-09-25 01:20:41.843377+00
\.


--
-- Data for Name: journal_evidence; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.journal_evidence (id, journal_id, object_key, sha256, mime_type, size_bytes, width_px, height_px, created_at) FROM stdin;
89de1edd-7585-4e5f-a66c-cfc19755b0ca	2b2ddf1f-a0e1-48ef-b054-b0fcf9b307a8	journals/88e60b00-646c-4235-b98e-958863e0b1ed.webp	031146d9ed50fedc374d11febd622d01df98f0500eaf30a074dead4a2214e6ad	image/webp	43616	960	960	2026-09-25 04:04:11.991155+00
\.


--
-- Data for Name: journal_reviews; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.journal_reviews (id, journal_id, reviewer_user_id, decision, comment, created_at) FROM stdin;
e25fca01-a2e5-4580-a138-77b62d1df647	2b2ddf1f-a0e1-48ef-b054-b0fcf9b307a8	2f131664-4d97-40af-965b-e0078b1fa9b3	reviewed	Goods	2026-09-25 04:05:17.576529+00
\.


--
-- Data for Name: notifications; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.notifications (id, recipient_user_id, event_key, type, title, body, resource_type, resource_id, read_at, created_at) FROM stdin;
f43eac2b-2eda-46ea-a343-91884d732f75	71d44ce9-a767-45f0-8d6d-650d598f9d4b	correction:f83913b5-ba0c-46a4-a007-07c719287233:approved	correction_decided	Correction request approved	Your correction request was approved.	correction_request	f83913b5-ba0c-46a4-a007-07c719287233	2026-09-25 03:48:42.394368+00	2026-09-25 01:41:06.814038+00
fd670dc9-a794-4de8-9774-77b5b80afe28	786f62a0-39d6-40f9-a3d9-ba750c34d85e	journal:2b2ddf1f-a0e1-48ef-b054-b0fcf9b307a8:review:1	journal_reviewed	Journal reviewed	Your daily journal was reviewed.	daily_journal	2b2ddf1f-a0e1-48ef-b054-b0fcf9b307a8	2026-09-25 05:24:38.854318+00	2026-09-25 04:05:17.576529+00
\.


--
-- Data for Name: ojt_assignments; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.ojt_assignments (id, trainee_id, site_id, start_date, end_date, required_minutes, break_rule_type, break_threshold_minutes, break_deduction_minutes, expected_weekdays, status, created_by, created_at, updated_at) FROM stdin;
1b585b1f-befb-491f-9891-10b43d24b53b	c0377216-d3d8-4d68-82c0-1594f17ead44	01b2c820-2dc9-4033-b352-8b028323cf73	2026-09-18	\N	28800	none	\N	0	{1,2,3,4,5,6,7}	active	\N	2026-09-25 01:20:50.889886+00	2026-09-25 01:20:50.889886+00
f5dd6760-c402-410b-876c-752c2a52c2a7	0a534bd5-c192-43bd-a224-c6a7158e2ca2	01b2c820-2dc9-4033-b352-8b028323cf73	2026-09-24	\N	1440	none	\N	0	{1,2,3,4,5}	active	2f131664-4d97-40af-965b-e0078b1fa9b3	2026-09-25 01:35:25.096957+00	2026-09-25 01:35:25.096957+00
f79bfa08-0180-444e-bc15-f063f7343d3e	9e1cfcf3-7812-4d97-8f91-6a925ee4178c	01b2c820-2dc9-4033-b352-8b028323cf73	2026-09-25	\N	720	none	\N	0	{1,2,3,4,5}	active	2f131664-4d97-40af-965b-e0078b1fa9b3	2026-09-25 01:35:41.548708+00	2026-09-25 01:35:41.548708+00
\.


--
-- Data for Name: ojt_sites; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.ojt_sites (id, name, address, latitude, longitude, allowed_radius_m, is_active, created_by, created_at, updated_at) FROM stdin;
01b2c820-2dc9-4033-b352-8b028323cf73	SkyCode	Isulan	6.619653778174684	124.59718386745926	200	t	\N	2026-09-25 01:20:50.877163+00	2026-09-25 06:15:53.003805+00
\.


--
-- Data for Name: schema_migrations; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.schema_migrations (version, applied_at) FROM stdin;
000001_users_sessions	2026-09-25 01:20:41.576607+00
000002_trainees_scopes	2026-09-25 01:20:41.747932+00
000003_sites_assignments	2026-09-25 01:20:41.782997+00
000004_settings_audit	2026-09-25 01:20:41.843377+00
000005_attendance	2026-09-25 01:20:41.982955+00
000006_journals	2026-09-25 01:20:42.072037+00
000007_corrections	2026-09-25 01:20:42.23081+00
000008_notifications	2026-09-25 01:20:42.370841+00
\.


--
-- Data for Name: trainee_profiles; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.trainee_profiles (id, user_id, student_number, program, year_level, contact_number, created_at, updated_at) FROM stdin;
c0377216-d3d8-4d68-82c0-1594f17ead44	be2bc35a-2e2a-4465-aa72-8759621fade6	2024-0001	BSIT	4	\N	2026-09-25 01:20:50.855763+00	2026-09-25 01:20:50.855763+00
0a534bd5-c192-43bd-a224-c6a7158e2ca2	71d44ce9-a767-45f0-8d6d-650d598f9d4b	TEST-001	BSIT	4	\N	2026-09-25 01:30:09.291079+00	2026-09-25 01:30:09.291079+00
9e1cfcf3-7812-4d97-8f91-6a925ee4178c	786f62a0-39d6-40f9-a3d9-ba750c34d85e	TEST-002	BSIT	4	\N	2026-09-25 01:30:09.291079+00	2026-09-25 01:30:09.291079+00
\.


--
-- Data for Name: user_sessions; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.user_sessions (id, user_id, token_hash, csrf_secret_hash, user_agent_hash, ip_prefix, expires_at, revoked_at, created_at, last_seen_at) FROM stdin;
9e4af8f7-56b3-4cd0-99cf-6e9be3d96242	2f131664-4d97-40af-965b-e0078b1fa9b3	02b1922bd24ad1e3cecb69ce9030fa7067cd60db69641277dfeaaf4cfae6c4a3	df1a07d8277cec2f990088fcd45b14bd0f3b9595d18670f06271a54ccd77223b	8504cac8d960aaf18d13986d1e7597c82cc6e0fc8258f20928dec5242ca43acb	172.18.0.0	2026-09-25 13:26:50.781324+00	\N	2026-09-25 01:26:50.816134+00	\N
4ef80c04-76ae-423d-b721-0b8e89200b63	2f131664-4d97-40af-965b-e0078b1fa9b3	39fc659d5663d13ed381cc89749b9b06f0ea12188efdcfcc81c170be426ee17f	0cde5f1c917f0be4604af5740434f18a5ebb20a8a9feba5b846ed7bb7d1b03ba	ac628f4cfda1bcac1aa445ee899b5f1385b4ffff94a36e6c4923a865bb69cfc6	172.18.0.0	2026-09-25 13:40:49.835343+00	2026-09-25 01:42:23.915474+00	2026-09-25 01:40:49.860286+00	2026-09-25 01:42:23.841711+00
92f7d5a7-6580-4d4c-a6ad-d2331cd9474d	2f131664-4d97-40af-965b-e0078b1fa9b3	4fc98e6c078208dab7220a8eb961344a60c7261573141e50be5a2876563210c2	5801f6f6bd24d24e4b15d99dc2f5856b4fe7607c54c82da5cf301a540259ce98	e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855	172.18.0.0	2026-09-25 13:30:08.404216+00	\N	2026-09-25 01:30:08.458349+00	2026-09-25 01:30:09.27893+00
1b511e8d-5683-4bed-87c1-988d2285b25c	71d44ce9-a767-45f0-8d6d-650d598f9d4b	0e11071cc67e3b449cf1301dd03585d3824bea0c290e0cfe3b0269ccb0ce57be	b22cdaf5ceb673bc71adeaa45f1ffacf47c6b631cfd4b24ffe21fddcac2d8790	8504cac8d960aaf18d13986d1e7597c82cc6e0fc8258f20928dec5242ca43acb	172.18.0.0	2026-09-25 13:30:39.277678+00	\N	2026-09-25 01:30:39.278583+00	\N
a6902904-dbd0-496f-bb04-1a3cd63a64e2	786f62a0-39d6-40f9-a3d9-ba750c34d85e	ff0d54f7f65d5205465150d6dd4c19a63f52f704fc08a560847851b1a56c31d8	454949e075756c6da4e4010d6d7221132d55de6628ea9b604f18a21f020936e0	8504cac8d960aaf18d13986d1e7597c82cc6e0fc8258f20928dec5242ca43acb	172.18.0.0	2026-09-25 13:30:39.582598+00	\N	2026-09-25 01:30:39.582917+00	\N
90780e04-9efb-4380-a344-3ee1fc20117c	786f62a0-39d6-40f9-a3d9-ba750c34d85e	77c92975505de74e15278dd4851de5fb4a6157b1462d9e3299981ca1d20cdab9	ecf998cd538d5b5e357524abe79424f489200677e81bc822b1485c9ca747b14a	ac628f4cfda1bcac1aa445ee899b5f1385b4ffff94a36e6c4923a865bb69cfc6	172.18.0.0	2026-09-25 15:11:35.717143+00	2026-09-25 03:15:20.110673+00	2026-09-25 03:11:35.71751+00	2026-09-25 03:15:20.10229+00
b38ad97f-fa6d-435b-9fd3-81d1d3ff773d	71d44ce9-a767-45f0-8d6d-650d598f9d4b	047b865187543bc47b93e6732a5dbec7bad2875520ce3cbbfdb27e993f9eb0b3	e80710fd3f6b3d10bb4da438ceee8bf1beb71862b1048488a694822ab78cf7b9	340e357be1c12b4f13ce33da0fc04c8f2e32284c9cfc40921dd984189c26b278	172.18.0.0	2026-09-25 13:39:13.237664+00	2026-09-25 01:40:16.520421+00	2026-09-25 01:39:13.242058+00	2026-09-25 01:40:16.51591+00
66c710d6-01ec-4b80-94dd-aeefbff5b803	be8659af-d267-4217-8bdf-6573794960f8	4c69fff7f48d2df326501963451bba3a48e6e4ae91fcbcf638b264bbd81b128e	50d3abb0f343dbe6360de7a479abfc3b541e9392e5d09bc0c1e807eabbb13d70	ac628f4cfda1bcac1aa445ee899b5f1385b4ffff94a36e6c4923a865bb69cfc6	172.18.0.0	2026-09-25 13:42:36.241709+00	2026-09-25 01:43:18.90078+00	2026-09-25 01:42:36.244492+00	2026-09-25 01:43:18.895767+00
1afbed2c-de64-4167-be1f-df548b6ee1f9	71d44ce9-a767-45f0-8d6d-650d598f9d4b	ef4e91c1f3a0d93801bed6f55f80af012afd80a555726014a2c4abb5bc9854f4	fd31ec97c9050c785a2ff5307f65d20dfcc012eb855315d8b6f697f3ca1e4acd	340e357be1c12b4f13ce33da0fc04c8f2e32284c9cfc40921dd984189c26b278	172.18.0.0	2026-09-25 13:31:30.996263+00	2026-09-25 01:31:49.145436+00	2026-09-25 01:31:31.004123+00	2026-09-25 01:31:49.139093+00
257f9cde-584f-4624-8967-308f276d3cad	2f131664-4d97-40af-965b-e0078b1fa9b3	ca983e548240d64303e4403cd5625d5dabd868cf3d738f8073aa0333d3af641d	9550a3295c1256de477ee5114c58997052e4e1a3f3d351c885aba7f1736bc77d	340e357be1c12b4f13ce33da0fc04c8f2e32284c9cfc40921dd984189c26b278	172.18.0.0	2026-09-25 13:33:21.08465+00	2026-09-25 01:39:01.878663+00	2026-09-25 01:33:21.086924+00	2026-09-25 01:39:01.845928+00
cc7db09a-baba-4665-bef0-08e2645a843f	2f131664-4d97-40af-965b-e0078b1fa9b3	c9813b6736d16c51c7f8c5b2c7a474fbc7f4954c9aa0ea4a0364f927a9b8c4d1	6e5a5644cc073b1b92bbcc8961e8d145be53827ff62dbeb3cd5c78e8529c0f6a	f2fb2718a135020651e79a9cb30b63437b5f08271e54d4f44ba79956278c81c2	172.18.0.0	2026-09-25 14:13:56.395589+00	2026-09-25 02:14:04.75067+00	2026-09-25 02:13:56.403875+00	2026-09-25 02:14:04.746802+00
6a429fea-07f1-489e-b131-53b70d3c3c2f	786f62a0-39d6-40f9-a3d9-ba750c34d85e	b3bae3206f1f5635f56761ff56597d44cfbf30e44724f09742a4522b4a7ed3cf	f01a1ab3f1c33eda654445f56772164401252f76d40a78ecd4616d5002ef152b	340e357be1c12b4f13ce33da0fc04c8f2e32284c9cfc40921dd984189c26b278	172.18.0.0	2026-09-25 13:32:06.373193+00	2026-09-25 01:32:54.438454+00	2026-09-25 01:32:06.374558+00	2026-09-25 01:32:54.433804+00
242d6741-f261-4001-b25e-4a5bcb04a1c8	be2bc35a-2e2a-4465-aa72-8759621fade6	9ce82c3bbe79ee202363d8e0f665eb1280e16ce74d32b675a02cdbf477e27b1a	4c4f539ea023dbe9af2178a6a0181c1455638439a6dffeeaadb5506d9df6797d	ac628f4cfda1bcac1aa445ee899b5f1385b4ffff94a36e6c4923a865bb69cfc6	172.18.0.0	2026-09-25 14:47:37.854271+00	2026-09-25 02:59:14.093254+00	2026-09-25 02:47:37.856198+00	2026-09-25 02:59:14.08538+00
a8cbb71c-95a0-4e00-98dd-16c7deca8afe	2f131664-4d97-40af-965b-e0078b1fa9b3	8e024a517a10002b85adfedbbc0e823fe934dba0458ad14b2bc84d2cd1e49da3	7f8593aad79b641314497d7fc58bdf537df21a6e910cd4e6857cff98607e3171	ac628f4cfda1bcac1aa445ee899b5f1385b4ffff94a36e6c4923a865bb69cfc6	172.18.0.0	2026-09-25 13:43:48.27535+00	2026-09-25 02:47:29.505935+00	2026-09-25 01:43:48.279813+00	2026-09-25 02:47:29.503056+00
5625d141-8d9a-4363-bc28-ae68d4bd2663	2f131664-4d97-40af-965b-e0078b1fa9b3	3b3c45dfc41d814b091437b28330bc431454037a314c55492ac3cdc83e0f6548	e7675ad1e1dac51c1cdbd3ba6b64c483bb675a2f641f324896c7d1f3642f62b9	ac628f4cfda1bcac1aa445ee899b5f1385b4ffff94a36e6c4923a865bb69cfc6	172.18.0.0	2026-09-25 15:30:03.534122+00	2026-09-25 03:30:14.288985+00	2026-09-25 03:30:03.534429+00	2026-09-25 03:30:14.287076+00
de0e7bc2-3bb0-4c77-b498-dd4a82044fa3	71d44ce9-a767-45f0-8d6d-650d598f9d4b	3e59a01643baea83a6e0b57e8bc4c70ba0f26a3049024ffce10fe6263f6250a7	232c0fc5f4736b8cc17dace19c8d39acb62e6309f50770c20353e53d5d4269b7	ac628f4cfda1bcac1aa445ee899b5f1385b4ffff94a36e6c4923a865bb69cfc6	172.18.0.0	2026-09-25 15:10:59.449967+00	2026-09-25 03:11:25.768678+00	2026-09-25 03:10:59.454747+00	2026-09-25 03:11:25.764047+00
5919e6d7-ba47-4fa5-b7d2-5fcb1d8159f4	2f131664-4d97-40af-965b-e0078b1fa9b3	278af2f45667703632844efc3e72b8e08994c27652163ef6783939bdbb4c266d	6116b48391c10b44fa300621fdae5aac5d3cd3a0f35853c8590a4f4f6bfbe502	ac628f4cfda1bcac1aa445ee899b5f1385b4ffff94a36e6c4923a865bb69cfc6	172.18.0.0	2026-09-25 15:15:47.434143+00	2026-09-25 03:29:40.97558+00	2026-09-25 03:15:47.445177+00	2026-09-25 03:29:40.963021+00
a9b92737-65f9-4601-8933-0e710e8d9a89	be8659af-d267-4217-8bdf-6573794960f8	68d5ed5e49b690251b7bd6da95a116228cf3f3b08fc27e8d2548f000526c8ba8	0a05f7c28d1cc05b2b328442f25d3290df93fda07b120370c6a585b7d234310b	c4ea766a755ddee08f488270809d57453d14de55e76ff46b7a9b02b099f0d440	172.18.0.0	2026-09-25 14:23:01.254248+00	2026-09-25 03:29:11.495781+00	2026-09-25 02:23:01.259136+00	2026-09-25 02:24:15.514987+00
679cac5a-7b7f-4b6d-9a99-d4ed56b434c7	621742d6-7e67-4b9f-a3de-c63f12d4dfa5	8b9e89ea7bd9aa3144be7071bc8d5fa088f4e8a5c23e5201c64ad61bc024b8e7	086cb0320f0fdd01bfac734ab2bf9246c17caf06043953531208f9e93b443346	ac628f4cfda1bcac1aa445ee899b5f1385b4ffff94a36e6c4923a865bb69cfc6	172.18.0.0	2026-09-25 15:29:47.616036+00	2026-09-25 03:29:55.557438+00	2026-09-25 03:29:47.621286+00	2026-09-25 03:29:55.553743+00
fad2ae72-4e66-424a-ac75-0e7c00f42506	be2bc35a-2e2a-4465-aa72-8759621fade6	fd100f348b3a0b1ff3a3d495265fc76fea15c5ef653bf26913d62e6de0e7fb26	d807bb67ba546633fd091739ff541fa7410ce26c034887669d4a089a42f53846	8504cac8d960aaf18d13986d1e7597c82cc6e0fc8258f20928dec5242ca43acb	172.18.0.0	2026-09-25 13:26:52.720066+00	2026-09-25 03:33:39.937801+00	2026-09-25 01:26:52.721304+00	\N
43c2e475-93a7-42e4-844c-77bac43a3773	be2bc35a-2e2a-4465-aa72-8759621fade6	7e4a36441250d4f4723e465eb01228fe2fec92e99da9a41cfacfac6799ff9a36	aa407bb3929ec6846b1fb37c3d3a3dcd365a3a30e572dea636270c6386204047	ac628f4cfda1bcac1aa445ee899b5f1385b4ffff94a36e6c4923a865bb69cfc6	172.18.0.0	2026-09-25 15:34:00.990944+00	2026-09-25 03:35:39.99067+00	2026-09-25 03:34:00.991775+00	2026-09-25 03:35:39.938645+00
8757b205-9c52-4b77-bb0a-9e924bad928a	2f131664-4d97-40af-965b-e0078b1fa9b3	950ab6a1129655877826716f56dacb1823952aae61a93208f88bd9746dc12588	5f287c9f6d03ec244f3d8ece736b4fe9b5af5eeeadd9d27675c35301c47429c5	ac628f4cfda1bcac1aa445ee899b5f1385b4ffff94a36e6c4923a865bb69cfc6	172.18.0.0	2026-09-25 15:33:15.749007+00	2026-09-25 03:33:54.998729+00	2026-09-25 03:33:15.754834+00	2026-09-25 03:33:54.995345+00
fca00ef9-e33c-4142-92e1-e31f71ee8b43	621742d6-7e67-4b9f-a3de-c63f12d4dfa5	354653fa0d49200aaf0a5be39a7ce1f38712efa23a2d0e110bb07271cd5a18eb	737117f7730fba4de54b3ab24e33bb19f9df0fe6cf3356b055d15beb3316b0fc	ac628f4cfda1bcac1aa445ee899b5f1385b4ffff94a36e6c4923a865bb69cfc6	172.18.0.0	2026-09-25 15:30:29.676766+00	2026-09-25 03:33:01.50742+00	2026-09-25 03:30:29.677627+00	2026-09-25 03:33:01.476708+00
58e05066-af36-4622-be72-f055e5689f72	621742d6-7e67-4b9f-a3de-c63f12d4dfa5	b3d9f0e5c977011b006c751ce2343ce191485de86b9151d471eb245422c9ce43	7e4e22835a1f3bbe9e7ac392102395f09ccf2fd6221a1eb2cb7eb2c29a736cc4	ac628f4cfda1bcac1aa445ee899b5f1385b4ffff94a36e6c4923a865bb69cfc6	172.18.0.0	2026-09-25 15:43:20.562584+00	2026-09-25 03:45:39.441819+00	2026-09-25 03:43:20.569819+00	2026-09-25 03:45:39.335385+00
c2bea7b8-4fb3-4e40-886d-a721ae5a5393	2f131664-4d97-40af-965b-e0078b1fa9b3	220ae81d0a4bf0242b2c4e9c1da3742ea6ce49198d93ac5c61d26bc208076048	7b9e7eba973c0cb26aeea760ec778b04febb9e905090532ca237fb0b135570f0	ac628f4cfda1bcac1aa445ee899b5f1385b4ffff94a36e6c4923a865bb69cfc6	172.18.0.0	2026-09-25 15:45:58.81152+00	2026-09-25 03:47:18.178704+00	2026-09-25 03:45:58.813422+00	2026-09-25 03:47:18.170186+00
552711d5-7d22-4a42-9f28-5f97012b7730	2f131664-4d97-40af-965b-e0078b1fa9b3	02f5e359e4c40a1d8eb5beb69b55a60b1b7991ae2df66e170344d3a83bf7603d	83c2f53edef4288a66d2a23bac206fc4dfee826d7361d7ae66194fe2e4cb8ccb	ac628f4cfda1bcac1aa445ee899b5f1385b4ffff94a36e6c4923a865bb69cfc6	172.18.0.0	2026-09-25 15:47:32.060258+00	2026-09-25 03:47:39.986027+00	2026-09-25 03:47:32.062008+00	2026-09-25 03:47:39.974475+00
bab1f4ff-6bff-4ce4-9e30-c1540e73c01c	2f131664-4d97-40af-965b-e0078b1fa9b3	8b0efa12d197a304c8ddde66dfc5fd1ae46988107a194916d28e814ed5171fc5	c857c0159881fba3e33b1adb1b364d16342fc60af9e1197ec7c69aa3382b793b	ac628f4cfda1bcac1aa445ee899b5f1385b4ffff94a36e6c4923a865bb69cfc6	172.18.0.0	2026-09-25 16:04:49.390355+00	2026-09-25 04:06:56.340277+00	2026-09-25 04:04:49.39099+00	2026-09-25 04:06:56.33833+00
06a6b4c8-d0db-4163-a224-7c6a4d2a7fb8	71d44ce9-a767-45f0-8d6d-650d598f9d4b	789c401b8724876d098e4a8a6641171d34693542d3b9407f0669dced4075b1c7	1f75ff7797ffde8f5dd9e6e71bd540b5e5688b4f3393116250f1845405ca709e	ac628f4cfda1bcac1aa445ee899b5f1385b4ffff94a36e6c4923a865bb69cfc6	172.18.0.0	2026-09-25 15:58:03.875789+00	2026-09-25 03:58:07.504422+00	2026-09-25 03:58:03.877526+00	2026-09-25 03:58:07.501188+00
4fada2ba-c265-4532-a48e-15ae4ff06cef	621742d6-7e67-4b9f-a3de-c63f12d4dfa5	bc9005944e57b9c3266dda22bc49c23abcd6f0eb449b77f46e094dd2bcb27498	f3ef652ccc88b94ac19d5abec62d87b37e74e30524994ee5837b3576f324eee6	ac628f4cfda1bcac1aa445ee899b5f1385b4ffff94a36e6c4923a865bb69cfc6	172.18.0.0	2026-09-25 17:44:49.213924+00	2026-09-25 06:08:04.92317+00	2026-09-25 05:44:49.222777+00	2026-09-25 06:08:04.918946+00
8b43bc13-77a4-4b4a-9664-0003197445c4	be2bc35a-2e2a-4465-aa72-8759621fade6	52df6bec8bae043f7cb9415e7effe4718959f5724b07e5e5889f250044d42e85	34d7d642769e73142823c91cebe66fcb088204ab4de11bb539798c08ded45a05	ac628f4cfda1bcac1aa445ee899b5f1385b4ffff94a36e6c4923a865bb69cfc6	172.18.0.0	2026-09-25 18:16:10.096431+00	2026-09-25 06:17:18.211028+00	2026-09-25 06:16:10.10065+00	2026-09-25 06:17:18.124433+00
1a2ae922-99b6-4b7f-9731-3e8978e8d2a1	2f131664-4d97-40af-965b-e0078b1fa9b3	cbbe7a6ef8393fd959317ec366bdaeab91aa58d2e33bc765c06574e43c9b3524	161cf57f6073790f0ac0cc538b845eee13a40a4d8d99a50a4a918e7eae448ce6	ac628f4cfda1bcac1aa445ee899b5f1385b4ffff94a36e6c4923a865bb69cfc6	172.18.0.0	2026-09-25 17:35:33.64619+00	2026-09-25 05:44:39.582644+00	2026-09-25 05:35:33.650011+00	2026-09-25 05:44:39.577164+00
cc401936-026f-469a-b624-b6cd3fe56acd	786f62a0-39d6-40f9-a3d9-ba750c34d85e	baeb5e7cc0674bbc89c833876a1292986b88a0d5ee9d088b287354027c792035	3ce3f2318178e2c2d7fd91c9ae941519212627072ed80c29f28e8fc0a2264050	ac628f4cfda1bcac1aa445ee899b5f1385b4ffff94a36e6c4923a865bb69cfc6	172.18.0.0	2026-09-25 16:02:06.95948+00	2026-09-25 04:04:36.621459+00	2026-09-25 04:02:06.959746+00	2026-09-25 04:04:36.598191+00
cc3c48b9-e782-4cc4-92b7-8edf75e6b398	2f131664-4d97-40af-965b-e0078b1fa9b3	e18d5c8938ac9c6793b270e99e1bea974213fe8744706a747b0d280205bbde1a	19a59781829ffaa00bd87f9cdb49c821ccd50eaa14d3504fb96c2e177b47b9b9	ac628f4cfda1bcac1aa445ee899b5f1385b4ffff94a36e6c4923a865bb69cfc6	172.18.0.0	2026-09-25 16:00:50.999843+00	2026-09-25 04:01:27.384841+00	2026-09-25 04:00:51.011261+00	2026-09-25 04:01:27.382614+00
eeb6b09e-3371-4f0e-b874-c6a36e8aaa34	71d44ce9-a767-45f0-8d6d-650d598f9d4b	adac7fb188aab216af8e5eb86b82e2b904d9a3c5018058897262ab0b32a34495	fb2701cf7b3223bb4cfdcae36521c68dff1bcb0066c3c11a7a7ccd29be9536fa	ac628f4cfda1bcac1aa445ee899b5f1385b4ffff94a36e6c4923a865bb69cfc6	172.18.0.0	2026-09-25 15:48:00.128428+00	2026-09-25 03:56:31.983659+00	2026-09-25 03:48:00.129998+00	2026-09-25 03:56:31.976786+00
65f26334-679a-4e45-bf2a-73c65e104e04	621742d6-7e67-4b9f-a3de-c63f12d4dfa5	2c7bb287bb528aab103a37daed97ff04cf68de625b113f2d971266080d1a58d0	5b8dbb107a86f0739b2799bc1f9ec07c855664ba7601b5a7852a7c877078c747	ac628f4cfda1bcac1aa445ee899b5f1385b4ffff94a36e6c4923a865bb69cfc6	172.18.0.0	2026-09-25 18:17:31.103082+00	\N	2026-09-25 06:17:31.106049+00	2026-09-25 06:19:19.673746+00
9de65d82-7470-4c69-ad7b-f76882fb3248	2f131664-4d97-40af-965b-e0078b1fa9b3	de8f4db46eab65a34532fee72c583b9f7908889f4235edff664571524e08d57a	3564895ea89deabb5909e40c39f84ea2f476a241316e071f4c9d2eb383ed7ecd	ac628f4cfda1bcac1aa445ee899b5f1385b4ffff94a36e6c4923a865bb69cfc6	172.18.0.0	2026-09-25 15:57:23.030004+00	2026-09-25 03:57:55.174177+00	2026-09-25 03:57:23.035897+00	2026-09-25 03:57:55.171497+00
d7851615-4043-4a18-8832-ea3e8abaaef6	621742d6-7e67-4b9f-a3de-c63f12d4dfa5	67844c019de8906a38fab20eda2aecd59e663757500896f9d1538eb44daf1af1	ce48b7036f64a3ad216c8d38aac042881bc9d2d676e527a4b44579082c63f80b	ac628f4cfda1bcac1aa445ee899b5f1385b4ffff94a36e6c4923a865bb69cfc6	172.18.0.0	2026-09-25 16:07:15.174607+00	2026-09-25 05:23:51.254553+00	2026-09-25 04:07:15.17507+00	2026-09-25 05:23:51.215554+00
c1b4d127-22ea-4e42-9e50-5b943e2414f6	71d44ce9-a767-45f0-8d6d-650d598f9d4b	ddc5d0b1d8f1b2673fdbc0524b436bacead02f7c90b9076c57486c56f25bfcfa	d6e22f7a16bfb520bb993c6a8e0356182f821fc9cd5b179a60b19952957e83bd	ac628f4cfda1bcac1aa445ee899b5f1385b4ffff94a36e6c4923a865bb69cfc6	172.18.0.0	2026-09-25 16:01:37.13748+00	2026-09-25 04:01:55.116504+00	2026-09-25 04:01:37.138401+00	2026-09-25 04:01:55.114157+00
b8544868-7508-4deb-ab90-307fccc91199	621742d6-7e67-4b9f-a3de-c63f12d4dfa5	234bd94da7f420a3dd21c261607b53cf7e4585276014ff5179d4c1af70ed1f47	ce55dba32aab201419ca98724b0ff93fe8696af2c5161edec91d8ba5d2d95123	ac628f4cfda1bcac1aa445ee899b5f1385b4ffff94a36e6c4923a865bb69cfc6	172.18.0.0	2026-09-25 18:14:35.883853+00	2026-09-25 06:15:56.192489+00	2026-09-25 06:14:35.885555+00	2026-09-25 06:15:56.189894+00
b7f78541-8ca4-4547-bcdc-9ef7ff7931a1	786f62a0-39d6-40f9-a3d9-ba750c34d85e	81abe4d8c677455fb1211696b9ee99d98aa6ac1b39ead7bbdde2841928506967	1f5c75d8a31c48f7340276f6940ca943b37d301d1c44be9af4bc80d7bdefcb4d	ac628f4cfda1bcac1aa445ee899b5f1385b4ffff94a36e6c4923a865bb69cfc6	172.18.0.0	2026-09-25 17:24:07.012468+00	2026-09-25 05:35:17.41818+00	2026-09-25 05:24:07.017324+00	2026-09-25 05:35:17.414493+00
ed8acc90-76bc-4a41-8472-d17f59a69746	be2bc35a-2e2a-4465-aa72-8759621fade6	bbadd8d979ac5488b95a64905cf91074d5e0eb5aedf21d3a351d54ed890a3783	c2be5c5bba581d2f3a0d444873ebc78bbbaae395b54a7e0144dea4c19de83665	ac628f4cfda1bcac1aa445ee899b5f1385b4ffff94a36e6c4923a865bb69cfc6	172.18.0.0	2026-09-25 18:13:35.743628+00	2026-09-25 06:14:24.114724+00	2026-09-25 06:13:35.747857+00	2026-09-25 06:14:24.109185+00
4157cc8a-431c-4404-900d-0e0f8045ea86	2f131664-4d97-40af-965b-e0078b1fa9b3	0cc5ac8db27bc24fc5576b631050580a3ed3804cae925db9db61c0b24ac15996	46b6f83be197563d8f6c75feec3ee57328507dac9673b9c01cbb6ac334a10a56	ac628f4cfda1bcac1aa445ee899b5f1385b4ffff94a36e6c4923a865bb69cfc6	172.18.0.0	2026-09-25 18:08:15.588396+00	2026-09-25 06:13:30.129676+00	2026-09-25 06:08:15.590722+00	2026-09-25 06:13:30.127613+00
\.


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.users (id, email, password_hash, role, account_status, display_name, last_login_at, created_at, updated_at) FROM stdin;
71d44ce9-a767-45f0-8d6d-650d598f9d4b	test12@gmail.com	$argon2id$v=19$m=65536,t=2,p=2$tGe+XkOmMNHkhmz5tRLiHA$RUw9HQIfZ9D8hrJCbM2Z7ndT3OSOHTsevJcTzxZMvzs	trainee	active	Test 12	2026-09-25 04:01:37.143188+00	2026-09-25 01:30:09.291079+00	2026-09-25 01:30:09.291079+00
786f62a0-39d6-40f9-a3d9-ba750c34d85e	test13@gmail.com	$argon2id$v=19$m=65536,t=2,p=2$tGe+XkOmMNHkhmz5tRLiHA$RUw9HQIfZ9D8hrJCbM2Z7ndT3OSOHTsevJcTzxZMvzs	trainee	active	Test 13	2026-09-25 05:24:07.064902+00	2026-09-25 01:30:09.291079+00	2026-09-25 01:30:09.291079+00
be8659af-d267-4217-8bdf-6573794960f8	mash@gmail.com	$argon2id$v=19$m=65536,t=2,p=2$ta5UMBBIjDIaMbnkjuuOUw$8hjgxfN04/hYEyi3lMTB+dBRsnv0EAofksZisTrEnng	coordinator	inactive	mash	2026-09-25 02:23:01.291351+00	2026-09-25 01:41:32.945256+00	2026-09-25 05:35:50.144644+00
2f131664-4d97-40af-965b-e0078b1fa9b3	admin@gmail.com	$argon2id$v=19$m=65536,t=2,p=2$tGe+XkOmMNHkhmz5tRLiHA$RUw9HQIfZ9D8hrJCbM2Z7ndT3OSOHTsevJcTzxZMvzs	admin	active	Admin	2026-09-25 06:08:15.597111+00	2026-09-25 01:20:46.303751+00	2026-09-25 03:46:45.764941+00
be2bc35a-2e2a-4465-aa72-8759621fade6	trainee1@ojt.local	$argon2id$v=19$m=65536,t=2,p=2$tGe+XkOmMNHkhmz5tRLiHA$RUw9HQIfZ9D8hrJCbM2Z7ndT3OSOHTsevJcTzxZMvzs	trainee	active	Trainee One	2026-09-25 06:16:10.116744+00	2026-09-25 01:20:50.842971+00	2026-09-25 03:33:39.937801+00
621742d6-7e67-4b9f-a3de-c63f12d4dfa5	jer@gmail.com	$argon2id$v=19$m=65536,t=2,p=2$tGe+XkOmMNHkhmz5tRLiHA$RUw9HQIfZ9D8hrJCbM2Z7ndT3OSOHTsevJcTzxZMvzs	coordinator	active	JershonC	2026-09-25 06:17:31.145888+00	2026-09-25 03:29:33.574148+00	2026-09-25 03:29:33.574148+00
\.


--
-- Name: attendance_adjustments attendance_adjustments_correction_request_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.attendance_adjustments
    ADD CONSTRAINT attendance_adjustments_correction_request_id_key UNIQUE (correction_request_id);


--
-- Name: attendance_adjustments attendance_adjustments_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.attendance_adjustments
    ADD CONSTRAINT attendance_adjustments_pkey PRIMARY KEY (id);


--
-- Name: attendance_evidence attendance_evidence_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.attendance_evidence
    ADD CONSTRAINT attendance_evidence_pkey PRIMARY KEY (id);


--
-- Name: attendance_sessions attendance_sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.attendance_sessions
    ADD CONSTRAINT attendance_sessions_pkey PRIMARY KEY (id);


--
-- Name: audit_events audit_events_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.audit_events
    ADD CONSTRAINT audit_events_pkey PRIMARY KEY (id);


--
-- Name: coordinator_scopes coordinator_scopes_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.coordinator_scopes
    ADD CONSTRAINT coordinator_scopes_pkey PRIMARY KEY (coordinator_user_id, trainee_id);


--
-- Name: correction_requests correction_requests_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.correction_requests
    ADD CONSTRAINT correction_requests_pkey PRIMARY KEY (id);


--
-- Name: daily_journals daily_journals_attendance_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.daily_journals
    ADD CONSTRAINT daily_journals_attendance_id_key UNIQUE (attendance_id);


--
-- Name: daily_journals daily_journals_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.daily_journals
    ADD CONSTRAINT daily_journals_pkey PRIMARY KEY (id);


--
-- Name: institution_settings institution_settings_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.institution_settings
    ADD CONSTRAINT institution_settings_pkey PRIMARY KEY (id);


--
-- Name: journal_evidence journal_evidence_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.journal_evidence
    ADD CONSTRAINT journal_evidence_pkey PRIMARY KEY (id);


--
-- Name: journal_reviews journal_reviews_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.journal_reviews
    ADD CONSTRAINT journal_reviews_pkey PRIMARY KEY (id);


--
-- Name: notifications notifications_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_pkey PRIMARY KEY (id);


--
-- Name: ojt_assignments ojt_assignments_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ojt_assignments
    ADD CONSTRAINT ojt_assignments_pkey PRIMARY KEY (id);


--
-- Name: ojt_sites ojt_sites_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ojt_sites
    ADD CONSTRAINT ojt_sites_pkey PRIMARY KEY (id);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: trainee_profiles trainee_profiles_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.trainee_profiles
    ADD CONSTRAINT trainee_profiles_pkey PRIMARY KEY (id);


--
-- Name: trainee_profiles trainee_profiles_user_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.trainee_profiles
    ADD CONSTRAINT trainee_profiles_user_id_key UNIQUE (user_id);


--
-- Name: user_sessions user_sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_sessions
    ADD CONSTRAINT user_sessions_pkey PRIMARY KEY (id);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: attendance_adjustments_attendance_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX attendance_adjustments_attendance_idx ON public.attendance_adjustments USING btree (attendance_id);


--
-- Name: attendance_evidence_action_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX attendance_evidence_action_key ON public.attendance_evidence USING btree (attendance_id, action);


--
-- Name: attendance_evidence_attendance_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX attendance_evidence_attendance_idx ON public.attendance_evidence USING btree (attendance_id);


--
-- Name: attendance_evidence_client_action_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX attendance_evidence_client_action_key ON public.attendance_evidence USING btree (trainee_id, client_action_id);


--
-- Name: attendance_sessions_assignment_date_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX attendance_sessions_assignment_date_idx ON public.attendance_sessions USING btree (assignment_id, attendance_date DESC);


--
-- Name: attendance_sessions_date_status_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX attendance_sessions_date_status_idx ON public.attendance_sessions USING btree (attendance_date, status);


--
-- Name: attendance_sessions_trainee_date_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX attendance_sessions_trainee_date_idx ON public.attendance_sessions USING btree (trainee_id, attendance_date DESC);


--
-- Name: attendance_sessions_trainee_date_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX attendance_sessions_trainee_date_key ON public.attendance_sessions USING btree (trainee_id, attendance_date);


--
-- Name: audit_events_actor_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX audit_events_actor_idx ON public.audit_events USING btree (actor_user_id, created_at DESC);


--
-- Name: audit_events_resource_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX audit_events_resource_idx ON public.audit_events USING btree (resource_type, resource_id, created_at DESC);


--
-- Name: coordinator_scopes_trainee_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX coordinator_scopes_trainee_idx ON public.coordinator_scopes USING btree (trainee_id);


--
-- Name: correction_requests_one_pending; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX correction_requests_one_pending ON public.correction_requests USING btree (attendance_id) WHERE (status = 'pending'::text);


--
-- Name: correction_requests_status_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX correction_requests_status_idx ON public.correction_requests USING btree (status, requested_at DESC);


--
-- Name: correction_requests_trainee_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX correction_requests_trainee_idx ON public.correction_requests USING btree (trainee_id, requested_at DESC);


--
-- Name: daily_journals_status_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX daily_journals_status_idx ON public.daily_journals USING btree (status, updated_at DESC);


--
-- Name: daily_journals_trainee_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX daily_journals_trainee_idx ON public.daily_journals USING btree (trainee_id, created_at DESC);


--
-- Name: institution_settings_singleton; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX institution_settings_singleton ON public.institution_settings USING btree ((true));


--
-- Name: journal_evidence_journal_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX journal_evidence_journal_idx ON public.journal_evidence USING btree (journal_id);


--
-- Name: journal_reviews_journal_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX journal_reviews_journal_idx ON public.journal_reviews USING btree (journal_id, created_at DESC);


--
-- Name: notifications_recipient_event_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX notifications_recipient_event_key ON public.notifications USING btree (recipient_user_id, event_key);


--
-- Name: notifications_recipient_read_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX notifications_recipient_read_idx ON public.notifications USING btree (recipient_user_id, read_at, created_at DESC);


--
-- Name: ojt_assignments_one_active_per_trainee; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX ojt_assignments_one_active_per_trainee ON public.ojt_assignments USING btree (trainee_id) WHERE (status = 'active'::text);


--
-- Name: ojt_assignments_site_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ojt_assignments_site_idx ON public.ojt_assignments USING btree (site_id);


--
-- Name: ojt_assignments_status_dates_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ojt_assignments_status_dates_idx ON public.ojt_assignments USING btree (status, start_date, end_date);


--
-- Name: ojt_assignments_trainee_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ojt_assignments_trainee_idx ON public.ojt_assignments USING btree (trainee_id);


--
-- Name: trainee_profiles_student_number_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX trainee_profiles_student_number_key ON public.trainee_profiles USING btree (student_number);


--
-- Name: user_sessions_token_hash_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX user_sessions_token_hash_key ON public.user_sessions USING btree (token_hash);


--
-- Name: user_sessions_user_expires_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX user_sessions_user_expires_idx ON public.user_sessions USING btree (user_id, expires_at);


--
-- Name: users_email_lower_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX users_email_lower_key ON public.users USING btree (lower(email));


--
-- Name: attendance_adjustments attendance_adjustments_approved_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.attendance_adjustments
    ADD CONSTRAINT attendance_adjustments_approved_by_fkey FOREIGN KEY (approved_by) REFERENCES public.users(id);


--
-- Name: attendance_adjustments attendance_adjustments_attendance_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.attendance_adjustments
    ADD CONSTRAINT attendance_adjustments_attendance_id_fkey FOREIGN KEY (attendance_id) REFERENCES public.attendance_sessions(id) ON DELETE CASCADE;


--
-- Name: attendance_adjustments attendance_adjustments_correction_request_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.attendance_adjustments
    ADD CONSTRAINT attendance_adjustments_correction_request_id_fkey FOREIGN KEY (correction_request_id) REFERENCES public.correction_requests(id);


--
-- Name: attendance_evidence attendance_evidence_attendance_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.attendance_evidence
    ADD CONSTRAINT attendance_evidence_attendance_id_fkey FOREIGN KEY (attendance_id) REFERENCES public.attendance_sessions(id) ON DELETE CASCADE;


--
-- Name: attendance_evidence attendance_evidence_trainee_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.attendance_evidence
    ADD CONSTRAINT attendance_evidence_trainee_id_fkey FOREIGN KEY (trainee_id) REFERENCES public.trainee_profiles(id);


--
-- Name: attendance_sessions attendance_sessions_assignment_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.attendance_sessions
    ADD CONSTRAINT attendance_sessions_assignment_id_fkey FOREIGN KEY (assignment_id) REFERENCES public.ojt_assignments(id);


--
-- Name: attendance_sessions attendance_sessions_trainee_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.attendance_sessions
    ADD CONSTRAINT attendance_sessions_trainee_id_fkey FOREIGN KEY (trainee_id) REFERENCES public.trainee_profiles(id);


--
-- Name: audit_events audit_events_actor_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.audit_events
    ADD CONSTRAINT audit_events_actor_user_id_fkey FOREIGN KEY (actor_user_id) REFERENCES public.users(id);


--
-- Name: coordinator_scopes coordinator_scopes_coordinator_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.coordinator_scopes
    ADD CONSTRAINT coordinator_scopes_coordinator_user_id_fkey FOREIGN KEY (coordinator_user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: coordinator_scopes coordinator_scopes_trainee_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.coordinator_scopes
    ADD CONSTRAINT coordinator_scopes_trainee_id_fkey FOREIGN KEY (trainee_id) REFERENCES public.trainee_profiles(id) ON DELETE CASCADE;


--
-- Name: correction_requests correction_requests_attendance_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.correction_requests
    ADD CONSTRAINT correction_requests_attendance_id_fkey FOREIGN KEY (attendance_id) REFERENCES public.attendance_sessions(id) ON DELETE CASCADE;


--
-- Name: correction_requests correction_requests_decided_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.correction_requests
    ADD CONSTRAINT correction_requests_decided_by_fkey FOREIGN KEY (decided_by) REFERENCES public.users(id);


--
-- Name: correction_requests correction_requests_trainee_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.correction_requests
    ADD CONSTRAINT correction_requests_trainee_id_fkey FOREIGN KEY (trainee_id) REFERENCES public.trainee_profiles(id);


--
-- Name: daily_journals daily_journals_attendance_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.daily_journals
    ADD CONSTRAINT daily_journals_attendance_id_fkey FOREIGN KEY (attendance_id) REFERENCES public.attendance_sessions(id) ON DELETE CASCADE;


--
-- Name: daily_journals daily_journals_trainee_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.daily_journals
    ADD CONSTRAINT daily_journals_trainee_id_fkey FOREIGN KEY (trainee_id) REFERENCES public.trainee_profiles(id);


--
-- Name: institution_settings institution_settings_updated_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.institution_settings
    ADD CONSTRAINT institution_settings_updated_by_fkey FOREIGN KEY (updated_by) REFERENCES public.users(id);


--
-- Name: journal_evidence journal_evidence_journal_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.journal_evidence
    ADD CONSTRAINT journal_evidence_journal_id_fkey FOREIGN KEY (journal_id) REFERENCES public.daily_journals(id) ON DELETE CASCADE;


--
-- Name: journal_reviews journal_reviews_journal_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.journal_reviews
    ADD CONSTRAINT journal_reviews_journal_id_fkey FOREIGN KEY (journal_id) REFERENCES public.daily_journals(id) ON DELETE CASCADE;


--
-- Name: journal_reviews journal_reviews_reviewer_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.journal_reviews
    ADD CONSTRAINT journal_reviews_reviewer_user_id_fkey FOREIGN KEY (reviewer_user_id) REFERENCES public.users(id);


--
-- Name: notifications notifications_recipient_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_recipient_user_id_fkey FOREIGN KEY (recipient_user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: ojt_assignments ojt_assignments_created_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ojt_assignments
    ADD CONSTRAINT ojt_assignments_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id);


--
-- Name: ojt_assignments ojt_assignments_site_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ojt_assignments
    ADD CONSTRAINT ojt_assignments_site_id_fkey FOREIGN KEY (site_id) REFERENCES public.ojt_sites(id);


--
-- Name: ojt_assignments ojt_assignments_trainee_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ojt_assignments
    ADD CONSTRAINT ojt_assignments_trainee_id_fkey FOREIGN KEY (trainee_id) REFERENCES public.trainee_profiles(id);


--
-- Name: ojt_sites ojt_sites_created_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ojt_sites
    ADD CONSTRAINT ojt_sites_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id);


--
-- Name: trainee_profiles trainee_profiles_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.trainee_profiles
    ADD CONSTRAINT trainee_profiles_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: user_sessions user_sessions_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_sessions
    ADD CONSTRAINT user_sessions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

\unrestrict qSCzibu9axTf34LROs8qHOeQrnYMAdQzaxVIrWEZa9kdWXc5bVDOMc7vkPDlhfe

