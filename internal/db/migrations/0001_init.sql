CREATE TABLE patients (
    id          BIGSERIAL PRIMARY KEY,
    last_name   TEXT        NOT NULL,
    first_name  TEXT        NOT NULL,
    middle_name TEXT,
    phone       TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX patients_name_idx ON patients (lower(last_name), lower(first_name));

CREATE TABLE services (
    id               BIGSERIAL PRIMARY KEY,
    name             TEXT        NOT NULL,
    duration_minutes INTEGER     NOT NULL CHECK (duration_minutes > 0),
    -- цена в рублях
    price            INTEGER     NOT NULL DEFAULT 0 CHECK (price >= 0),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX services_name_uniq ON services (lower(name));

CREATE TABLE appointments (
    id         BIGSERIAL PRIMARY KEY,
    patient_id BIGINT      NOT NULL REFERENCES patients (id) ON DELETE RESTRICT,
    service_id BIGINT      REFERENCES services (id) ON DELETE RESTRICT,
    starts_at  TIMESTAMPTZ NOT NULL,
    ends_at    TIMESTAMPTZ NOT NULL,
    note       TEXT        NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT appointments_time_check CHECK (ends_at > starts_at),
    -- Две записи не могут пересекаться по времени. Проверка на уровне БД,
    -- поэтому даже одновременная запись с телефона и компьютера не создаст накладку.
    CONSTRAINT appointments_no_overlap EXCLUDE USING gist (tstzrange(starts_at, ends_at) WITH &&)
);

CREATE INDEX appointments_patient_idx ON appointments (patient_id, starts_at);
