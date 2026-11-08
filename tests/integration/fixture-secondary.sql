CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

CREATE TABLE public.procurements (
	id bigserial PRIMARY KEY,
	status text NOT NULL,
	amount numeric NOT NULL DEFAULT 0
);

CREATE INDEX procurements_status_idx
	ON public.procurements (status);

INSERT INTO public.procurements (status, amount)
VALUES
	('secondary', 500);

GRANT USAGE ON SCHEMA public TO pgo_mcp_reader;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO pgo_mcp_reader;
GRANT SELECT, USAGE ON ALL SEQUENCES IN SCHEMA public TO pgo_mcp_reader;
