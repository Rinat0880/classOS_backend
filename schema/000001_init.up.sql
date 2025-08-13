CREATE TABLE
    users (
        id SERIAL PRIMARY KEY,
        fullname varchar(255) not null unique,
        role TEXT NOT NULL CHECK (role IN ('admin', 'client')),
        password_hash varchar(255) not null
    );

CREATE TABLE
    groups (
        id serial not null unique,
        name varchar(255) not null
    );

CREATE TABLE
    users_lists (
        id serial not null unique,
        user_id int references users (id) on delete cascade,
        group_id int references groups (id) on delete cascade,
        PRIMARY KEY (user_id, group_id)
    );

CREATE TABLE
    whitelist_global (
        id SERIAL PRIMARY KEY,
        group_id INT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
        resource TEXT NOT NULL
    );