-- Seed: Brasil — país, estados e capitais
-- Executado apenas em desenvolvimento (profile dev no docker-compose)

DO $$
DECLARE
  v_country_id UUID;
  v_state      RECORD;
BEGIN

  -- ── País ──────────────────────────────────────────────────────────────────
  INSERT INTO countries (name, code)
  VALUES ('Brasil', 'BR')
  ON CONFLICT (code) DO NOTHING;

  SELECT id INTO v_country_id FROM countries WHERE code = 'BR';

  -- ── Estados ───────────────────────────────────────────────────────────────
  INSERT INTO states (name, code, country_id) VALUES
    ('Acre',                'AC', v_country_id),
    ('Alagoas',             'AL', v_country_id),
    ('Amapá',               'AP', v_country_id),
    ('Amazonas',            'AM', v_country_id),
    ('Bahia',               'BA', v_country_id),
    ('Ceará',               'CE', v_country_id),
    ('Distrito Federal',    'DF', v_country_id),
    ('Espírito Santo',      'ES', v_country_id),
    ('Goiás',               'GO', v_country_id),
    ('Maranhão',            'MA', v_country_id),
    ('Mato Grosso',         'MT', v_country_id),
    ('Mato Grosso do Sul',  'MS', v_country_id),
    ('Minas Gerais',        'MG', v_country_id),
    ('Pará',                'PA', v_country_id),
    ('Paraíba',             'PB', v_country_id),
    ('Paraná',              'PR', v_country_id),
    ('Pernambuco',          'PE', v_country_id),
    ('Piauí',               'PI', v_country_id),
    ('Rio de Janeiro',      'RJ', v_country_id),
    ('Rio Grande do Norte', 'RN', v_country_id),
    ('Rio Grande do Sul',   'RS', v_country_id),
    ('Rondônia',            'RO', v_country_id),
    ('Roraima',             'RR', v_country_id),
    ('Santa Catarina',      'SC', v_country_id),
    ('São Paulo',           'SP', v_country_id),
    ('Sergipe',             'SE', v_country_id),
    ('Tocantins',           'TO', v_country_id)
  ON CONFLICT (code, country_id) DO NOTHING;

  -- ── Capitais (uma cidade por estado para uso inicial) ─────────────────────
  FOR v_state IN SELECT id, code FROM states WHERE country_id = v_country_id LOOP
    INSERT INTO cities (name, state_id)
    SELECT capital, v_state.id FROM (VALUES
      ('AC', 'Rio Branco'),
      ('AL', 'Maceió'),
      ('AP', 'Macapá'),
      ('AM', 'Manaus'),
      ('BA', 'Salvador'),
      ('CE', 'Fortaleza'),
      ('DF', 'Brasília'),
      ('ES', 'Vitória'),
      ('GO', 'Goiânia'),
      ('MA', 'São Luís'),
      ('MT', 'Cuiabá'),
      ('MS', 'Campo Grande'),
      ('MG', 'Belo Horizonte'),
      ('PA', 'Belém'),
      ('PB', 'João Pessoa'),
      ('PR', 'Curitiba'),
      ('PE', 'Recife'),
      ('PI', 'Teresina'),
      ('RJ', 'Rio de Janeiro'),
      ('RN', 'Natal'),
      ('RS', 'Porto Alegre'),
      ('RO', 'Porto Velho'),
      ('RR', 'Boa Vista'),
      ('SC', 'Florianópolis'),
      ('SP', 'São Paulo'),
      ('SE', 'Aracaju'),
      ('TO', 'Palmas')
    ) AS t(code, capital)
    WHERE t.code = v_state.code
    ON CONFLICT DO NOTHING;
  END LOOP;

  RAISE NOTICE 'Localização do Brasil carregada.';
END $$;
