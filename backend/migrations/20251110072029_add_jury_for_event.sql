-- +goose Up
INSERT INTO users (login, password, role)
VALUES
('ngarin', 'ngarin4821', 2),
('dkhairulin', 'dkhairulin7392', 2),
('amaximenko', 'amaximenko1064', 2),
('asapakhov', 'asapakhov9253', 2),
('tmukhadzhinov', 'tmukhadzhinov5810', 2),
('vushakov', 'vushakov4477', 2),
('atatulyan', 'atatulyan9318', 2),
('ishumilina', 'ishumilina6420', 2),
('vkravchenko', 'vkravchenko2895', 2),
('nkosova', 'nkosova3762', 2),
('akhlyustova', 'akhlyustova8174', 2),
('miraliev', 'miraliev9035', 2),
('sfedorov', 'sfedorov1958', 2),
('yusupov', 'yusupov8649', 2),
('ashadeev', 'ashadeev5716', 2),
('nbayramaliev', 'nbayramaliev7482', 2),
('vryzhov', 'vryzhov4190', 2),
('doskolkov', 'doskolkov2387', 2),
('abondarenko', 'abondarenko9934', 2);

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
