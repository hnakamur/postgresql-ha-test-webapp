CREATE SEQUENCE IF NOT EXISTS pings_id_seq;

CREATE TABLE IF NOT EXISTS pings (
    id bigint NOT NULL DEFAULT nextval('pings_id_seq'),
    created_at timestamp with time zone NOT NULL,
    PRIMARY KEY (id)
);
ALTER SEQUENCE pings_id_seq OWNED BY pings.id;
