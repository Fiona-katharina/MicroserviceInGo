CREATE TABLE products
(
    id SERIAL,
    name TEXT NOT NULL,
    price NUMERIC(10,2) NOT NULL DEFAULT 0.00,
    CONSTRAINT products_pkey PRIMARY KEY (id)
)

CREATE TABLE users
(
    id SERIAL,
    name TEXT NOT NULL,
    cartID NUMERIC DEFAULT -1,
    CONSTRAINT users_pkey PRIMARY KEY (id)
)
