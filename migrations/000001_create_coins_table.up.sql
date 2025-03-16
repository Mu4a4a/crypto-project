CREATE TABLE IF NOT EXISTS coins (
    title varchar(50) primary key,
    cost real not null,
    actual_at timestamp not null
);