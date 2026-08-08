CREATE TYPE contact_type AS ENUM ('email', 'phone', 'whatsapp');

CREATE TYPE document_type AS ENUM ('cpf', 'rg', 'cnh', 'passport', 'cnpj', 'state_registration');

CREATE TYPE user_role AS ENUM ('admin', 'employee', 'visitor');

CREATE TYPE address_type AS ENUM ('comercial', 'residencial');

CREATE TYPE stock_movement_type AS ENUM ('in', 'out', 'adjustment');

CREATE TYPE service_order_status AS ENUM ('open', 'in_progress', 'completed', 'cancelled');
