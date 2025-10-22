-- +goose Up
-- +goose StatementBegin
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
BEGIN
  SELECT id INTO event_id
  FROM public.events
  WHERE name = 'Второй региональный математический турнир школьных команд «Королевская битва»'
  LIMIT 1;

  IF event_id IS NULL THEN
    RAISE EXCEPTION 'Event not found with name: Второй региональный математический турнир школьных команд «Королевская битва»';
  END IF;

  FOR idx IN 1..34 LOOP
    SELECT id INTO temp_id
    FROM public.schools
    WHERE name = school_names[idx]
    LIMIT 1;

    IF temp_id IS NULL THEN
      RAISE EXCEPTION 'School not found with name: %', school_names[idx];
    END IF;

    school_ids := school_ids || temp_id;
  END LOOP;

  FOR idx IN 1..34 LOOP
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
      INSERT INTO public.participants (fio, class, gender, school_id)
      VALUES (
        'Участник ' || idx || '_' || sub,
        (8 + ((idx * sub) % 4))::smallint,
        ((idx + sub) % 2)::smallint,
        school_ids[idx]
      ) RETURNING id INTO p_id;

      INSERT INTO public.team_participants (team_id, participant_id)
      VALUES (t_id, p_id);
    END LOOP;
  END LOOP;
END $$;
-- +goose StatementEnd

-- +goose Down
-- No rollback for mock data, or add DELETE statements if needed.