CREATE TABLE IF NOT EXISTS coins (
    id serial primary key,
    title varchar(50),
    cost real not null,
    actual_at timestamp not null
);