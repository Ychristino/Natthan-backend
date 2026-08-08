-- ─── Seed: primeiro administrador ─────────────────────────────────────────────
-- Senha do admin: admin123
-- Gerado com: bcrypt cost=10

DO $$
DECLARE
  v_person_id UUID;
  v_user_id   UUID;
BEGIN

  IF EXISTS (SELECT 1 FROM users WHERE email = 'admin@natthan.com') THEN
    RAISE NOTICE 'Usuário admin já existe. Pulando.';
  ELSE
    INSERT INTO persons (id, full_name)
    VALUES (gen_random_uuid(), 'Administrador')
    RETURNING id INTO v_person_id;

    v_user_id := gen_random_uuid();

    INSERT INTO users (id, person_id, username, email, password_hash, active)
    VALUES (
      v_user_id,
      v_person_id,
      'admin',
      'admin@natthan.com',
      '$2b$10$X82ujjGs0HLvXygi8U0l1.GhB6nN5hqxLbaPty3ul7nOSZj6q34KK',
      true
    );

    INSERT INTO user_roles (id, user_id, role, active)
    VALUES (gen_random_uuid(), v_user_id, 'admin', true);

    RAISE NOTICE 'Admin criado: admin@natthan.com / admin123';
  END IF;

END $$;
