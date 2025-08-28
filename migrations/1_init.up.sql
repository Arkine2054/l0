CREATE TABLE orders (
    order_uid          TEXT PRIMARY KEY,
    track_number       TEXT NOT NULL,
    entry              TEXT NOT NULL,
    locale             TEXT NOT NULL,
    internal_signature TEXT,
    customer_id        TEXT NOT NULL,
    delivery_service   TEXT NOT NULL,
    shardkey           TEXT NOT NULL,
    sm_id              INT NOT NULL,
    date_created       TIMESTAMPTZ NOT NULL,
    oof_shard          TEXT NOT NULL
    );

CREATE TABLE deliveries (
    id        SERIAL PRIMARY KEY,
    order_uid TEXT NOT NULL REFERENCES orders(order_uid) ON DELETE CASCADE,
    name      TEXT NOT NULL,
    phone     TEXT NOT NULL,
    zip       TEXT NOT NULL,
    city      TEXT NOT NULL,
    address   TEXT NOT NULL,
    region    TEXT NOT NULL,
    email     TEXT NOT NULL,
CONSTRAINT deliveries_order_uid_unique UNIQUE (order_uid) -- уникальная доставка для заказа
    );

CREATE TABLE payments (
    id            SERIAL PRIMARY KEY,
    order_uid     TEXT NOT NULL REFERENCES orders(order_uid) ON DELETE CASCADE,
    transaction   TEXT NOT NULL,
    request_id    TEXT,
    currency      TEXT NOT NULL,
    provider      TEXT NOT NULL,
    amount        INT NOT NULL,
    payment_dt    BIGINT NOT NULL,
    bank          TEXT NOT NULL,
    delivery_cost INT NOT NULL,
    goods_total   INT NOT NULL,
    custom_fee    INT NOT NULL,
CONSTRAINT payments_order_uid_unique UNIQUE (order_uid) -- уникальный платёж для заказа
    );

CREATE TABLE items (
    id           SERIAL PRIMARY KEY,
    order_uid    TEXT NOT NULL REFERENCES orders(order_uid) ON DELETE CASCADE,
    chrt_id      INT NOT NULL,
    track_number TEXT NOT NULL,
    price        INT NOT NULL,
    rid          TEXT NOT NULL,
    name         TEXT NOT NULL,
    sale         INT NOT NULL,
    size         TEXT NOT NULL,
    total_price  INT NOT NULL,
    nm_id        INT NOT NULL,
    brand        TEXT NOT NULL,
    status       INT NOT NULL
    );
ALTER TABLE items
    ADD CONSTRAINT uniq_item_per_order UNIQUE (order_uid, rid);
