-- +goose Up
-- +goose StatementBegin
CREATE TABLE live_score_by_tours (
                                     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                     team_id UUID NOT NULL,
                                     event_id UUID NOT NULL,
                                     participant_id UUID NOT NULL,
                                     score INT NOT NULL DEFAULT 0 CHECK (score >= 0),

                                     CONSTRAINT fk_team
                                         FOREIGN KEY (team_id)
                                             REFERENCES teams (id)
                                             ON UPDATE CASCADE
                                             ON DELETE CASCADE,

                                     CONSTRAINT fk_event
                                         FOREIGN KEY (event_id)
                                             REFERENCES events (id)
                                             ON UPDATE CASCADE
                                             ON DELETE CASCADE,

                                     CONSTRAINT fk_participant
                                         FOREIGN KEY (participant_id)
                                             REFERENCES participants (id)
                                             ON UPDATE CASCADE
                                             ON DELETE CASCADE,

                                     CONSTRAINT unique_team_event_participant UNIQUE (team_id, event_id, participant_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS live_score_by_tours;
-- +goose StatementEnd
