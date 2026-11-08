CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

CREATE ROLE pgo_mcp_reader
	LOGIN
	PASSWORD 'reader_ci_password';

GRANT CONNECT ON DATABASE pgo_mcp_test TO pgo_mcp_reader;
GRANT USAGE ON SCHEMA public TO pgo_mcp_reader;
GRANT pg_read_all_stats TO pgo_mcp_reader;

CREATE TABLE public.procurements (
	id bigserial PRIMARY KEY,
	status text NOT NULL,
	amount numeric NOT NULL DEFAULT 0
);

CREATE INDEX procurements_status_idx
	ON public.procurements (status);

INSERT INTO public.procurements (status, amount)
VALUES
	('active', 100),
	('closed', 250);

GRANT SELECT ON ALL TABLES IN SCHEMA public TO pgo_mcp_reader;
GRANT SELECT, USAGE ON ALL SEQUENCES IN SCHEMA public TO pgo_mcp_reader;
