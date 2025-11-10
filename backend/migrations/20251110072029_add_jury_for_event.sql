-- +goose Up
INSERT INTO users (login, password, role)
VALUES
('ngarin', 'ngarin', 2),
('dkhairulin', 'dkhairulin', 2),
('amaximenko', 'amaximenko', 2),
('asapakhov', 'asapakhov', 2),
('tmukhadzhinov', 'tmukhadzhinov', 2),
('vushakov', 'vushakov', 2),
('atatulyan', 'atatulyan', 2),
('ishumilina', 'ishumilina', 2),
('vkravchenko', 'vkravchenko', 2),
('nkosova', 'nkosova', 2),
('akhlyustova', 'akhlyustova', 2),
('miraliev', 'miraliev', 2),
('sfedorov', 'sfedorov', 2),
('yusupov', 'yusupov', 2),
('ashadeev', 'ashadeev', 2),
('nbayramaliev', 'nbayramaliev', 2),
('vryzhov', 'vryzhov', 2),
('doskolkov', 'doskolkov', 2),
('abondarenko', 'abondarenko', 2);

-- +goose Down
DELETE FROM users
WHERE login IN (
  'ngarin',
  'dkhairulin',
  'amaximenko',
  'asapakhov',
  'tmukhadzhinov',
  'vushakov',
  'atatulyan',
  'ishumilina',
  'vkravchenko',
  'nkosova',
  'akhlyustova',
  'miraliev',
  'sfedorov',
  'yusupov',
  'ashadeev',
  'nbayramaliev',
  'vryzhov',
  'doskolkov',
  'abondarenko'
);

SELECT 'down SQL query';
