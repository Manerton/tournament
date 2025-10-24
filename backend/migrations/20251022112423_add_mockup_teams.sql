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
    'МБОУ Астрахани СОШ № 23',
    'МБОУ Астрахани СОШ № 53 им. Н. М. Скоморохова',
    'МКОУ СОШ Заволжское',
    'ГБ профессиональное ОУ Астраханской области Астраханский государственный колледж профессиональных технологий',
    'МКОУ СОШ № 3 Харабали',
    'МБОУ Седлистинская СОШ',
    'МКОУ Винновская ООШ',
    'Прикаспийский филиал муниципального казённого образовательного учреждения СОШ Трусово',
    'МБОУ СОШ №10 Верхний Баскунчак',
    'МБОУ Новоурусовская СОШ муниципального образования Красноярский район Астраханской области',
    'ГКОУ Астраханской области Школа-интернат №7 для обучающихся с ограниченными возможностями здоровья',
    'Муниципальное бюджетное дошкольное образовательное учреждения Астрахани детский сад № 130',
    'МКОУ Верхнекалиновская СОШ Камызякский район',
    'МКОУ СОШ Восток Муниципального образования Енотаевский район',
    'МКОУ Копановская ООШ муниципального образования Енотаевский район',
    'МБОУ Цветновская СОШ (филиал Большой Могой)',
    'МБОУ Забузанская СОШ им. Турченко Э. П. муниципального образования Красноярский район Астраханской области',
    'МБОУ Астрахани СОШ № 22',
    'МБОУ Астрахани СОШ № 12',
    'МКОУ Волго-Каспийская СОШ',
    'МКОУ Зензелинская СОШ',
    'МКОУ СОШ Речное',
    'ГАОУ Астраханской области дополнительного профессионального образования Институт развития образования',
    'МБОУ Красноярская СОШ №1 муниципального образования Красноярский район Астраханской области',
    'МКОУ Камызякская СОШ №4',
    'Школа Альфа дети',
    'МБОУ Астрахани СОШ № 49',
    'МБОУ Астрахани СОШ № 33 им. Н. А. Мордовиной',
    'Государственное областное учреждение ДО Астраханский областной центр развития творчества',
    'МБОУ Астрахани СОШ № 55',
    'МБОУ Ново-Булгаринская СОШ',
    'МБОУ Астрахани Гимназия №2',
    'МКОУ Олинская СОШ',
    'МКОУ Лиманская СОШ №2'
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
