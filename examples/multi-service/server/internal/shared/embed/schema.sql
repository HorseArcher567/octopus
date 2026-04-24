DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS products;

CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL,
    email TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    UNIQUE (username),
    UNIQUE (email)
);

CREATE TABLE orders (
    order_id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    product_name TEXT NOT NULL,
    amount REAL NOT NULL,
    status TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);

CREATE INDEX idx_orders_user_id ON orders(user_id);

CREATE TABLE products (
    product_id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    price REAL NOT NULL,
    stock INTEGER NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);

INSERT INTO products (name, description, price, stock, created_at, updated_at) VALUES
    ('book', 'golang book', 39.90, 100, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('keyboard', 'mechanical keyboard', 199.00, 20, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('mouse', 'wireless mouse', 89.00, 40, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);
