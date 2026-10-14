CREATE TABLE resources (
    id          BIGSERIAL PRIMARY KEY,
    name        TEXT NOT NULL UNIQUE,
    kind        TEXT NOT NULL CHECK (kind IN ('material', 'machine', 'labour')),
    unit        TEXT NOT NULL,
    capacity    DOUBLE PRECISION CHECK (capacity IS NULL OR capacity >= 0),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE processes (
    id          BIGSERIAL PRIMARY KEY,
    name        TEXT NOT NULL UNIQUE,
    position    INTEGER NOT NULL,
    wip_limit   INTEGER CHECK (wip_limit IS NULL OR wip_limit > 0),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX processes_position_idx ON processes (position);

CREATE TABLE process_resources (
    id          BIGSERIAL PRIMARY KEY,
    process_id  BIGINT NOT NULL REFERENCES processes (id) ON DELETE CASCADE,
    resource_id BIGINT NOT NULL REFERENCES resources (id) ON DELETE RESTRICT,
    quantity    DOUBLE PRECISION NOT NULL CHECK (quantity > 0),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (process_id, resource_id)
);

CREATE INDEX process_resources_process_idx ON process_resources (process_id);

CREATE TABLE orders (
    id          BIGSERIAL PRIMARY KEY,
    reference   TEXT NOT NULL,
    quantity    INTEGER NOT NULL CHECK (quantity > 0),
    status      TEXT NOT NULL DEFAULT 'pending'
                CHECK (status IN ('pending', 'in_progress', 'done', 'cancelled')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX orders_status_idx ON orders (status);

CREATE TABLE order_steps (
    id          BIGSERIAL PRIMARY KEY,
    order_id    BIGINT NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    process_id  BIGINT NOT NULL REFERENCES processes (id) ON DELETE RESTRICT,
    position    INTEGER NOT NULL,
    status      TEXT NOT NULL DEFAULT 'queued'
                CHECK (status IN ('queued', 'running', 'done')),
    started_at  TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (order_id, process_id)
);

CREATE INDEX order_steps_order_idx ON order_steps (order_id);
CREATE INDEX order_steps_process_status_idx ON order_steps (process_id, status);

CREATE TABLE consumptions (
    id            BIGSERIAL PRIMARY KEY,
    order_step_id BIGINT NOT NULL REFERENCES order_steps (id) ON DELETE CASCADE,
    resource_id   BIGINT NOT NULL REFERENCES resources (id) ON DELETE RESTRICT,
    quantity      DOUBLE PRECISION NOT NULL CHECK (quantity > 0),
    occurred_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX consumptions_order_step_idx ON consumptions (order_step_id);
CREATE INDEX consumptions_resource_idx ON consumptions (resource_id);
