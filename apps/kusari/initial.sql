CREATE SCHEMA atlas_schema_revisions;


CREATE TABLE atlas_schema_revisions.atlas_schema_revisions (
  version character varying NOT NULL,
  description character varying NOT NULL,
  type bigint NOT NULL DEFAULT 2,
  applied bigint NOT NULL DEFAULT 0,
  total bigint NOT NULL DEFAULT 0,
  executed_at timestamp with time zone NOT NULL,
  execution_time bigint NOT NULL,
  error text NULL,
  error_stmt text NULL,
  hash character varying NOT NULL,
  partial_hashes jsonb NULL,
  operator_version character varying NOT NULL
);


ALTER TABLE atlas_schema_revisions.atlas_schema_revisions
ADD CONSTRAINT atlas_schema_revisions_pkey PRIMARY KEY (version);

INSERT INTO
  "atlas_schema_revisions"."atlas_schema_revisions" (
    "applied",
    "description",
    "error",
    "error_stmt",
    "executed_at",
    "execution_time",
    "hash",
    "operator_version",
    "partial_hashes",
    "total",
    "type",
    "version"
  )
VALUES
  (
    '15',
    'initial',
    '',
    '',
    '2026-05-29 19:11:04.887374+00',
    '213244',
    'Yf2XjELq60rBWdfIV7w2dpuVbjdbcHgHsw/nTT9UWrk=',
    'Atlas CLI v1.2.1-f7a44b2-canary',
    NULL,
    '15',
    '2',
    '20260530000000'
  )
