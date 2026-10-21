-- +goose Up
CREATE TABLE meetings (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    title text NOT NULL CHECK (char_length(title) BETWEEN 1 AND 200),
    kind text NOT NULL DEFAULT 'REGULAR' CHECK (kind IN ('REGULAR', 'EXTRAORDINARY')),
    held_at timestamptz NOT NULL,
    ends_at timestamptz,
    location text CHECK (char_length(location) <= 200),
    quorum_required boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CHECK (ends_at IS NULL OR ends_at > held_at)
);

CREATE TRIGGER meetings_set_updated_at BEFORE UPDATE ON meetings
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- +goose Down
DROP TABLE meetings;
