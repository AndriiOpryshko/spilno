-- Bootstrap: the one association of Stage 1. Deterministic id, used by the meetings backfill
-- (20261027120008) and by dev/test seeding. The name is renamed by the board in the app.
INSERT INTO associations (id, name)
VALUES ('00000000-0000-7000-8000-000000000001', 'Test Association')
ON CONFLICT DO NOTHING;
