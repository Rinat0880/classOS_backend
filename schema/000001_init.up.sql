CREATE TABLE
    users (
        id SERIAL PRIMARY KEY,
        name varchar(255) not null,
        username varchar(255) not null unique,
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
    whitelist (
        id SERIAL PRIMARY KEY,
        group_id INT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
        resource TEXT NOT NULL
    );

INSERT INTO users (name, username, role, password_hash)
VALUES (
    'Super Admin',
    'admin01',
    'admin',
    '64687569676664677339383334356b6a6e6664393867646a37cc096b78234d87087d0b45986bea7ad2614b72'
)
ON CONFLICT (username) DO NOTHING;

CREATE OR REPLACE FUNCTION prevent_superadmin_delete()
RETURNS trigger AS $$
BEGIN
    IF OLD.username = 'admin01' THEN
        RAISE EXCEPTION 'Суперадмина удалить нельзя';
    END IF;
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_prevent_superadmin_delete
BEFORE DELETE ON users
FOR EACH ROW EXECUTE FUNCTION prevent_superadmin_delete();

CREATE OR REPLACE FUNCTION prevent_superadmin_update()
RETURNS trigger AS $$
BEGIN
    IF OLD.username = 'admin01' AND NEW.role <> 'admin' THEN
        RAISE EXCEPTION 'Нельзя изменить роль суперадмина';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_prevent_superadmin_update
BEFORE UPDATE ON users
FOR EACH ROW EXECUTE FUNCTION prevent_superadmin_update();