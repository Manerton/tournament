-- +goose Up
-- +goose StatementBegin
-- Test data: 16 accepted teams with 5 participants each for the quiz-show event.
-- Participant initial_score mirrors the Swiss mockup style (50-100 per participant,
-- summed into teams.initial_score). Quiz team rows with access codes are created
-- directly, so heralds can log in without going through the accept flow.
DO $$
DECLARE
  quiz_event_id uuid := '11111111-1111-1111-1111-111111111111';
  school_rec record;
  t_id uuid;
  p_id uuid;
  idx int := 0;
  sub int;
  total_score int;
BEGIN
  FOR school_rec IN SELECT id FROM public.schools ORDER BY name LIMIT 16 LOOP
    idx := idx + 1;

    INSERT INTO public.teams (name, responsible_name, responsible_phone, event_id, school_id, status)
    VALUES (
      'Квиз-команда ' || idx,
      'Ответственный ' || idx,
      '+7' || lpad((910000000 + idx)::text, 9, '0'),
      quiz_event_id,
      school_rec.id,
      2  -- accepted
    ) RETURNING id INTO t_id;

    FOR sub IN 1..5 LOOP
      INSERT INTO public.participants (fio, class, gender, school_id, initial_score)
      VALUES (
        'Квиз-участник ' || idx || '_' || sub,
        (8 + ((idx * sub) % 4))::smallint,
        ((idx + sub) % 2)::smallint,
        school_rec.id,
        (50 + ((idx * sub) % 51))::int  -- initial_score от 50 до 100
      ) RETURNING id INTO p_id;

      INSERT INTO public.team_participants (team_id, participant_id)
      VALUES (t_id, p_id);
    END LOOP;

    -- Суммарный стартовый балл команды
    SELECT COALESCE(SUM(p.initial_score), 0)
    INTO total_score
    FROM public.participants p
    JOIN public.team_participants tp ON tp.participant_id = p.id
    WHERE tp.team_id = t_id;

    UPDATE public.teams
    SET initial_score = total_score
    WHERE id = t_id;

    -- Код доступа для герольда: 0001..0016
    INSERT INTO public.quiz_teams (event_id, team_id, access_code)
    VALUES (quiz_event_id, t_id, lpad(idx::text, 4, '0'));
  END LOOP;
END $$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DO $$
DECLARE
  quiz_event_id uuid := '11111111-1111-1111-1111-111111111111';
BEGIN
  DELETE FROM public.quiz_teams WHERE event_id = quiz_event_id;

  DELETE FROM public.participants
  WHERE id IN (
    SELECT tp.participant_id
    FROM public.team_participants tp
    JOIN public.teams t ON t.id = tp.team_id
    WHERE t.event_id = quiz_event_id
  );

  DELETE FROM public.team_participants
  WHERE team_id IN (SELECT id FROM public.teams WHERE event_id = quiz_event_id);

  DELETE FROM public.teams WHERE event_id = quiz_event_id;
END $$;
-- +goose StatementEnd
