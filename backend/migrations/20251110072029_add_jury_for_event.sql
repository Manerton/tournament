-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION hash_password(password TEXT)
RETURNS TEXT AS $$
BEGIN
    -- Используем crypt из pgcrypto для bcrypt-хеширования
    RETURN crypt(password, gen_salt('bf', 14)); -- bf = blowfish (bcrypt), 14 - стоимость
END;
$$ LANGUAGE plpgsql;


INSERT INTO users (login, password, role)
VALUES
('ngarin', hash_password('ngarin4821'), 2),
('dkhairulin', hash_password('dkhairulin7392'), 2),
('amaximenko', hash_password('amaximenko1064'), 2),
('asapakhov', hash_password('asapakhov9253'), 2),
('tmukhadzhinov',hash_password('tmukhadzhinov5810'), 2),
('vushakov', hash_password('vushakov4477'), 2),
('atatulyan', hash_password('atatulyan9318'), 2),
('ishumilina', hash_password('ishumilina6420'), 2),
('vkravchenko', hash_password('vkravchenko2895'), 2),
('nkosova', hash_password('nkosova3762'), 2),
('akhlyustova', hash_password('akhlyustova8174'), 2),
('miraliev', hash_password('miraliev9035'), 2),
('sfedorov', hash_password('sfedorov1958'), 2),
('yusupov', hash_password('yusupov8649'), 2),
('ashadeev', hash_password('ashadeev5716'), 2),
('nbayramaliev', hash_password('nbayramaliev7482'), 2),
('vryzhov', hash_password('vryzhov4190'), 2),
('doskolkov', hash_password('doskolkov2387'), 2),
('abondarenko', hash_password('abondarenko9934'), 2);
-- +goose StatementEnd
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
