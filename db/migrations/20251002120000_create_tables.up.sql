-- 1. App User
CREATE TABLE IF NOT EXISTS app_user (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    role TEXT NOT NULL CHECK (role IN ('admin', 'staff', 'manager')),
    created_at TIMESTAMP DEFAULT now()
);

-- 2. Company
CREATE TABLE IF NOT EXISTS company (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    gstin TEXT NULL,
    created_at TIMESTAMP DEFAULT now()
);

-- 3. Company Address
CREATE TABLE IF NOT EXISTS company_address (
    id BIGSERIAL PRIMARY KEY,
    company_id BIGINT REFERENCES company(id) ON DELETE CASCADE,
    address_line TEXT NOT NULL,
    city TEXT NOT NULL,
    state TEXT NOT NULL,
    pincode TEXT NOT NULL,
    is_default BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT now()
);

-- 4. Bilty Address (snapshot)
CREATE TABLE IF NOT EXISTS bilty_address (
    id BIGSERIAL PRIMARY KEY,
    company_id BIGINT REFERENCES company(id),
    address_line TEXT NOT NULL,
    city TEXT NOT NULL,
    state TEXT NOT NULL,
    pincode TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT now()
);

-- 5. Bilty
CREATE TABLE IF NOT EXISTS bilty (
    id BIGSERIAL PRIMARY KEY,
    bilty_no BIGSERIAL UNIQUE NOT NULL, -- optional, starts from 1
    consignor_company_id BIGINT REFERENCES company(id),
    consignee_company_id BIGINT REFERENCES company(id),
    consignor_address_id BIGINT REFERENCES bilty_address(id),
    consignee_address_id BIGINT REFERENCES bilty_address(id),
    from_location TEXT NOT NULL,
    to_location TEXT NOT NULL,
    date DATE NOT NULL,
    to_pay NUMERIC(12,2) DEFAULT 0,
    gstin TEXT NULL,
    inv_no TEXT,
    pvt_marks TEXT,
    permit_no TEXT,
    value_rupees NUMERIC(12,2),
    remarks TEXT,
    hamali NUMERIC(12,2),
    dd_charges NUMERIC(12,2),
    other_charges NUMERIC(12,2),
    fov NUMERIC(12,2),
    statistical TEXT,
    created_by BIGINT REFERENCES app_user(id),
    created_at TIMESTAMP DEFAULT now(),
    status TEXT NOT NULL CHECK (status IN ('draft', 'complete'))
);

-- 6. Goods
CREATE TABLE IF NOT EXISTS goods (
    id BIGSERIAL PRIMARY KEY,
    bilty_id BIGINT REFERENCES bilty(id) ON DELETE CASCADE,
    particulars TEXT NOT NULL,
    num_of_pkts INTEGER NOT NULL,
    weight_kg NUMERIC(10,2),
    rate NUMERIC(10,2),
    per TEXT,
    amount NUMERIC(12,2)
);

-- 7. Initial Details
CREATE TABLE IF NOT EXISTS initial_setup (
    id BIGSERIAL PRIMARY KEY,
    company_name TEXT NOT NULL,
    gstin TEXT NULL,
    address TEXT NOT NULL,
    city TEXT NOT NULL,
    state TEXT NOT NULL,
    pincode TEXT NOT NULL,
    mobile JSONB,
    footnote TEXT,
    created_at TIMESTAMP DEFAULT now()
);
