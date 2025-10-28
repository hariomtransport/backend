-- 1. App User
CREATE TABLE IF NOT EXISTS app_user (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    role TEXT NOT NULL CHECK (role IN ('admin', 'staff', 'manager')),
    created_at TIMESTAMP DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_app_user_role ON app_user(role);


-- 2. Company
CREATE TABLE IF NOT EXISTS company (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    gstin TEXT NULL,
    created_at TIMESTAMP DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_company_name ON company(name);
CREATE INDEX IF NOT EXISTS idx_company_gstin ON company(gstin);


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
CREATE INDEX IF NOT EXISTS idx_company_address_company_id ON company_address(company_id);
CREATE INDEX IF NOT EXISTS idx_company_address_city_state ON company_address(city, state);
CREATE INDEX IF NOT EXISTS idx_company_address_is_default ON company_address(is_default);


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
CREATE INDEX IF NOT EXISTS idx_bilty_address_company_id ON bilty_address(company_id);
CREATE INDEX IF NOT EXISTS idx_bilty_address_city_state ON bilty_address(city, state);


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
    gstin TEXT,
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
    updated_at TIMESTAMP DEFAULT now(),
    pdf_created_at TIMESTAMP,
    pdf_path TEXT,
    status TEXT NOT NULL CHECK (status IN ('draft', 'complete'))
);

-- ⚡ Performance indexes for fast bilty fetch
CREATE INDEX IF NOT EXISTS idx_bilty_status ON bilty(status);
CREATE INDEX IF NOT EXISTS idx_bilty_date ON bilty(date);
CREATE INDEX IF NOT EXISTS idx_bilty_consignor_company_id ON bilty(consignor_company_id);
CREATE INDEX IF NOT EXISTS idx_bilty_consignee_company_id ON bilty(consignee_company_id);
CREATE INDEX IF NOT EXISTS idx_bilty_created_by ON bilty(created_by);
CREATE INDEX IF NOT EXISTS idx_bilty_from_to_location ON bilty(from_location, to_location);
CREATE INDEX IF NOT EXISTS idx_bilty_inv_no ON bilty(inv_no);


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
CREATE INDEX IF NOT EXISTS idx_goods_bilty_id ON goods(bilty_id);
CREATE INDEX IF NOT EXISTS idx_goods_particulars ON goods(particulars);


-- 7. Initial Setup
CREATE TABLE IF NOT EXISTS initial_setup (
    id BIGSERIAL PRIMARY KEY,
    company_name TEXT NOT NULL,
    gstin TEXT NULL,
    address TEXT NOT NULL,
    city TEXT NOT NULL,
    state TEXT NOT NULL,
    pincode TEXT NOT NULL,
    mobile JSONB,
    footnote JSONB,
    created_at TIMESTAMP DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_initial_setup_company_name ON initial_setup(company_name);


-- ✅ Insert default company details (only if not exists)
INSERT INTO initial_setup (
    company_name,
    gstin,
    address,
    city,
    state,
    pincode,
    mobile,
    footnote
)
SELECT
    'HARI OM TRANSPORT AGENCY',
    '10BFCPK7445G2ZQ',
    'H.O.-N.H.31,BYPASS ROAD,NEAR-MANGLA ASTHAN BIHAR SHARIF(NALANDA)',
    'Patna',
    'Bihar',
    '800007',
    '[{"number": "9835892439", "label": "H.O"}, {"number": "9608241192", "label": "PHR"}]'::jsonb,
    '[
        "N.B. We have read the terms & conditions stipulated over leaf and hereby declared that the particulars furnished are correct.Carriers are not responsible for Leakage,Breakage & Damages.",
        "1.NO ANY RECALLING ACCEPTED AFTER 30 DAYS(1 MONTH).",
        "2.THE TRANSPORT SHALL NOT BOUND TO GIVE ANY FURTHER RECORD AFTER TWO MONTHS."
    ]'::jsonb
WHERE NOT EXISTS (
    SELECT 1 FROM initial_setup WHERE company_name = 'HARI OM TRANSPORT AGENCY'
);
