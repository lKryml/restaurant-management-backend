CREATE TABLE users(
    id      uuid not null primary key default gen_random_uuid(),
    name    varchar(255) not null,
    img     varchar(255),
    phone   varchar(255) not null,
    email   varchar(255) unique not null,
    password varchar(255) not null,
    created_at timestamp not null default current_timestamp,
    updated_at timestamp not null default current_timestamp

)