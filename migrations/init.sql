create table if not exists users (
    id uuid primary key not null,
    email text unique not null,
    password text not null,
    role text not null,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


create table if not exists pvz (
    id uuid primary key not null,
    registration_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    city text not null
);

create table if not exists reception (
    id uuid primary key not null,
    datetime TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    pvz_id uuid not null,
    status text not null
);

create table if not exists product (
    id uuid primary key not null,
    datetime TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    type text not null,
    reception_id uuid not null
);
