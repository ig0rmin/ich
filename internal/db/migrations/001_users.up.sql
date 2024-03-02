CREATE TABLE users (
    id serial PRIMARY KEY,
    username varchar NOT NULL,
    email varchar NOT NULL,
    password varchar NOT NULL
);

CREATE INDEX users_email ON users(email);
