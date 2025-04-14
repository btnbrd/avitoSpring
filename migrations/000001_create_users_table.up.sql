CREATE TABLE IF NOT EXISTS users (
       id UUID PRIMARY KEY,
       email TEXT NOT NULL UNIQUE CHECK (email ~* '^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$'),
       password_hash TEXT NOT NULL,
       role TEXT NOT NULL CHECK (role IN ('employee', 'moderator'))
);


CREATE TABLE IF NOT EXISTS pvz (
       id UUID PRIMARY KEY,
       registration_date TIMESTAMP NOT NULL,
       city TEXT NOT NULL CHECK (city IN ('Москва', 'Санкт-Петербург', 'Казань'))
);

CREATE TABLE IF NOT EXISTS  receptions (
     id UUID PRIMARY KEY,
     datetime TIMESTAMP NOT NULL,
     pvz_id UUID NOT NULL REFERENCES pvz(id),
     status TEXT NOT NULL CHECK (status IN ('in_progress', 'close'))
 );

CREATE TABLE IF NOT EXISTS products (
     id UUID PRIMARY KEY,
     datetime TIMESTAMP NOT NULL,
     type TEXT NOT NULL CHECK (type IN ('электроника', 'одежда', 'обувь')),
     reception_id UUID NOT NULL REFERENCES receptions(id)
);
