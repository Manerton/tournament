-- +goose Up
-- +goose StatementBegin
-- 1. Добавляем поле initial_score, если его нет в 2 таблицы
ALTER TABLE public.participants
ADD COLUMN IF NOT EXISTS initial_score int DEFAULT 0;

ALTER TABLE public.teams
ADD COLUMN IF NOT EXISTS initial_score int DEFAULT 0;


DO $$
DECLARE
  school_names varchar[] := ARRAY[
    'ЧОУ "Школа Сергеевская"',
    'ФГБПОУ "Астраханское СУВУ"',
    'ГБОУ АО "Школа-интернат № 3 для обучающихся с ограниченными возможностями здоровья"',
    'ГКОУ АО "ОШИ № 5"',
    'ГКОУ АО "ОШИ № 6"',
    'ГКОУ АО "Школа - интернат №1"',
    'ГКОУ АО "Школа-интернат № 3 для обучающихся с ограниченными возможностями здоровья"',
    'ГКОУ АО «Школа-интернат № 7»',
   ' ГБПОУ АО "Астраханский губернский техникум"',
    'ГБОУ АО "Школа-интернат одаренных детей им. А.П. Гужвина"',
    'ГБОУ АО "Инженерная школа"',
    'ГБОУ АО "АТЛ"',
    'ГБОУ АО "АКШИ им. П.О. Сухого"',
    'ГАПОУ АО "Черноярский губернский колледж"',
    'ГБОУ АО "АЛГ"',
    'ГКОУ АО "Средняя общеобразовательная школа № 4"',
    'ГАОУ АО "Казачий кадетский корпус имени атамана И.А. Бирюкова"'
  ];
  school_ids uuid[] := '{}';
  event_id uuid;
  temp_id uuid;
  t_id uuid;
  p_id uuid;
  idx int;
  sub int;
  total_score INT;
BEGIN
  -- 2. Создаём новое событие
  INSERT INTO public.events (
    name,
    start_date,
    end_date,
    previous_event_id,
    type_draws,
    type,
    status,
    stage
  )
  VALUES (
    'Математический турнир 2025',
    '2025-11-12 10:00:00+03',
    '2025-11-12 15:00:00+03',
    NULL,
    3,  -- Swiss
    1,  -- тип события
    3,  -- активен
    0
  )
  RETURNING id INTO event_id;

  -- 3. Сопоставляем школы
  FOR idx IN 1..array_length(school_names, 1) LOOP
    SELECT id INTO temp_id
    FROM public.schools
    WHERE name = school_names[idx]
    LIMIT 1;

    IF temp_id IS NULL THEN
      RAISE NOTICE 'Школа не найдена: %', school_names[idx];
    ELSE
      school_ids := school_ids || temp_id;
    END IF;
  END LOOP;

  -- 4. Создаём команды и участников с initial_score
  FOR idx IN 1..array_length(school_ids, 1) LOOP
    INSERT INTO public.teams (name, responsible_name, responsible_phone, event_id, school_id, status)
    VALUES (
      'Команда ' || idx,
      'Ответственный ' || idx,
      '+7' || lpad((900000000 + idx)::text, 9, '0'),
      event_id,
      school_ids[idx],
      2
    ) RETURNING id INTO t_id;

    FOR sub IN 1..5 LOOP
      INSERT INTO public.participants (fio, class, gender, school_id, initial_score)
      VALUES (
        'Участник ' || idx || '_' || sub,
        (8 + ((idx * sub) % 4))::smallint,
        ((idx + sub) % 2)::smallint,
        school_ids[idx],
        (50 + ((idx * sub) % 51))::int  -- initial_score от 50 до 100
      ) RETURNING id INTO p_id;

      INSERT INTO public.team_participants (team_id, participant_id)
      VALUES (t_id, p_id);
    END LOOP;

    -- Считаем суммарный стартовый балл участников этой команды
    SELECT COALESCE(SUM(p.initial_score), 0)
    INTO total_score
    FROM public.participants p
    JOIN public.team_participants tp ON tp.participant_id = p.id
    WHERE tp.team_id = t_id;

    -- 🔹 Обновляем поле initial_score в таблице teams
    UPDATE public.teams
    SET initial_score = total_score
    WHERE id = t_id;
  END LOOP;
END $$;
-- +goose StatementEnd

-- +goose Down
-- Удаляем созданное событие и связанные с ним команды и участников
DO $$
DECLARE
  eid uuid;
BEGIN
  SELECT id INTO eid FROM public.events WHERE name = 'Математический турнир 2025' LIMIT 1;

  IF eid IS NOT NULL THEN
    DELETE FROM public.team_participants
    WHERE team_id IN (SELECT id FROM public.teams WHERE event_id = eid);

    DELETE FROM public.participants
    WHERE id IN (
      SELECT participant_id FROM public.team_participants
      WHERE team_id IN (SELECT id FROM public.teams WHERE event_id = eid)
    );

    DELETE FROM public.teams WHERE event_id = eid;
    DELETE FROM public.events WHERE id = eid;
  END IF;
END $$;
