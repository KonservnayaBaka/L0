CREATE TABLE IF NOT EXISTS orders (
    order_uid VARCHAR(255) PRIMARY KEY,
    track_number VARCHAR(255) NOT NULL UNIQUE,
    entry VARCHAR(255),
    locale VARCHAR(10),
    internal_signature TEXT,
    customer_id VARCHAR(255),
    delivery_service VARCHAR(255),
    shardkey VARCHAR(10),
    sm_id INT,
    date_created TIMESTAMP,
    oof_shard VARCHAR(10)
);

CREATE TABLE IF NOT EXISTS delivery (
    order_uid VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(20) NOT NULL UNIQUE,
    zip INT NOT NULL,
    city VARCHAR(255) NOT NULL,
    address TEXT NOT NULL,
    region VARCHAR(255),
    email VARCHAR(255) UNIQUE,

    FOREIGN KEY (order_uid) REFERENCES orders(order_uid)
);

CREATE TABLE IF NOT EXISTS payment (
    order_uid VARCHAR(255) PRIMARY KEY,
    transaction VARCHAR(255) NOT NULL UNIQUE,
    request_id VARCHAR(255),
    currency VARCHAR(10) NOT NULL,
    provider VARCHAR(255) NOT NULL,
    amount INT NOT NULL,
    payment_dt TIMESTAMP NOT NULL,
    bank VARCHAR(255) NOT NULL,
    delivery_cost NUMERIC NOT NULL,
    goods_total NUMERIC NOT NULL,
    custom_fee NUMERIC DEFAULT 0,

    FOREIGN KEY (order_uid) REFERENCES orders(order_uid)
);

CREATE TABLE IF NOT EXISTS items (
    id SERIAL PRIMARY KEY,
    order_uid VARCHAR(255),
    chrt_id INT NOT NULL UNIQUE,
    track_number VARCHAR(255) NOT NULL UNIQUE,
    price NUMERIC NOT NULL,
    rid VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    sale INT DEFAULT 0,
    size VARCHAR(10) DEFAULT 0,
    total_price NUMERIC NOT NULL,
    nm_id INT NOT NULL UNIQUE,
    brand VARCHAR(255),
    status INT,

    FOREIGN KEY (order_uid) REFERENCES orders(order_uid)
);