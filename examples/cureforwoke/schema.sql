-- ============================================================
-- Ron DeSantis's Cure for Woke — Database Schema
-- A scheduling & invoicing system for a torture dungeon.
-- ============================================================

-- Customers who book sessions (and often become forgetful after)
CREATE TABLE customers (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  name        TEXT    NOT NULL,
  alias       TEXT,                  -- some prefer pseudonyms
  phone       TEXT,
  email       TEXT,
  address     TEXT,
  birth_date  TEXT,                  -- YYYY-MM-DD
  risk_score  INTEGER DEFAULT 0,     -- 0=reliable, 100=will definitely "forget"
  notes       TEXT,
  created_at  TEXT    DEFAULT (datetime('now')),
  updated_at  TEXT    DEFAULT (datetime('now'))
);

-- Available services (torture methods)
CREATE TABLE services (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  name        TEXT    NOT NULL,
  description TEXT,
  duration    INTEGER NOT NULL,      -- minutes
  base_price  REAL    NOT NULL,
  requires_waiver INTEGER DEFAULT 1, -- 1 = must sign waiver
  intensity   INTEGER DEFAULT 5,     -- 1-10 scale
  active      INTEGER DEFAULT 1
);

-- Standard service catalog
INSERT INTO services (name, description, duration, base_price, intensity) VALUES
  ('Waterboarding',          'Classic re-education technique. Refreshing and persuasive.',                   45,  299.99,  7),
  ('Iron Maiden',            'A relaxing fitting session. Custom-tailored discomfort.',                     60,  499.99,  9),
  ('Dolly Parton Rock',            'Locked in a small room while "9 to 5" plays on loop. Psychological warfare.', 120, 399.99,  6),
  ('Drawing & Quartering',   'Historical reenactment. Horses not included but available as add-on.',         90,  799.99,  10),
  ('The Stockades',          'Public ridicule in the town square. Produce-throwing extra.',                  30,  149.99,  3),
  ('Rack & Roll',            'Our signature stretching program. Increases height temporarily.',              60,  599.99,  8),
  ('Dunking Stool',          'Hydration therapy with a medieval twist. Repeated as needed.',                 30,  199.99,  5),
  ('Scold''s Bridle',        'Silence is golden. Speech-restricting headwear for the chatty.',               20,  89.99,   2),
  ('Pear of Anguish',        'A unique oral expansion experience. Dentist recommended.',                     15,  249.99,  4),
  ('Breaking Wheel',         'Rotisserie-style relaxation. Comes with complementary seasoning.',              90,  699.99,  9);

-- Scheduled appointments
CREATE TABLE appointments (
  id            INTEGER PRIMARY KEY AUTOINCREMENT,
  customer_id   INTEGER NOT NULL REFERENCES customers(id),
  service_id    INTEGER NOT NULL REFERENCES services(id),
  scheduled_for TEXT    NOT NULL,      -- ISO datetime
  duration      INTEGER NOT NULL,     -- minutes (derived from service, can be overridden)
  status        TEXT    DEFAULT 'pending',  -- pending, confirmed, completed, cancelled, no_show
  waiver_signed INTEGER DEFAULT 0,
  notes         TEXT,
  created_at    TEXT    DEFAULT (datetime('now'))
);

-- Invoices generated after sessions
CREATE TABLE invoices (
  id            INTEGER PRIMARY KEY AUTOINCREMENT,
  customer_id   INTEGER NOT NULL REFERENCES customers(id),
  appointment_id INTEGER REFERENCES appointments(id),
  invoice_date  TEXT    DEFAULT (date('now')),
  due_date      TEXT    NOT NULL,
  subtotal      REAL    NOT NULL DEFAULT 0,
  tax           REAL    NOT NULL DEFAULT 0,
  total         REAL    NOT NULL DEFAULT 0,
  paid          REAL    DEFAULT 0,
  balance       REAL    DEFAULT 0,
  status        TEXT    DEFAULT 'pending',  -- pending, sent, overdue, paid, written_off
  dunning_level INTEGER DEFAULT 0,          -- 0=none, 1=reminder, 2=final, 3=collections
  notes         TEXT,
  created_at    TEXT    DEFAULT (datetime('now'))
);

-- Line items on invoices
CREATE TABLE invoice_items (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  invoice_id  INTEGER NOT NULL REFERENCES invoices(id),
  description TEXT    NOT NULL,
  quantity    INTEGER DEFAULT 1,
  unit_price  REAL    NOT NULL,
  total       REAL    NOT NULL
);

-- Payments received
CREATE TABLE payments (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  invoice_id  INTEGER NOT NULL REFERENCES invoices(id),
  amount      REAL    NOT NULL,
  method      TEXT    DEFAULT 'cash',   -- cash, card, crypto, barter, found_in_pocket
  received_at TEXT    DEFAULT (datetime('now')),
  notes       TEXT
);

-- Collection actions (for chasing down forgetful customers)
CREATE TABLE collection_actions (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  invoice_id  INTEGER NOT NULL REFERENCES invoices(id),
  action_type TEXT    NOT NULL,  -- email, phone, visit, psychic_message, subpoena
  contact     TEXT,              -- where we reached them (or tried to)
  response    TEXT,              -- what they said (if anything coherent)
  result      TEXT,              -- paid, promised, ignored, fled_jurisdiction
  acted_at    TEXT    DEFAULT (datetime('now')),
  created_at  TEXT    DEFAULT (datetime('now'))
);

-- Helpful views
CREATE VIEW overdue_invoices AS
  SELECT
    i.id AS invoice_id,
    c.name AS customer_name,
    c.phone,
    c.email,
    i.invoice_date,
    i.due_date,
    i.total,
    i.paid,
    i.balance,
    i.dunning_level,
    julianday('now') - julianday(i.due_date) AS days_overdue
  FROM invoices i
  JOIN customers c ON c.id = i.customer_id
  WHERE i.status NOT IN ('paid', 'written_off')
    AND i.balance > 0
    AND i.due_date < date('now')
  ORDER BY days_overdue DESC;
