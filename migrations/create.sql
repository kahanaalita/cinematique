-- Таблица фильмов
CREATE TABLE films (
                       id SERIAL PRIMARY KEY,
                       title VARCHAR(150)  NOT NULL CHECK (LENGTH(title) >= 1),
                       description VARCHAR(1000),
                       release_year INT NOT NULL,
                       rating FLOAT CHECK (rating BETWEEN 0 AND 10)
                   );

-- Таблица актеров
CREATE TABLE actors (
                        id SERIAL PRIMARY KEY,
                        name VARCHAR(100) NOT NULL,
                        gender VARCHAR(100) NOT NULL,
                        birth_date DATE
);

-- Связывающая таблица для фильмов и актеров
CREATE TABLE film_actor (
                            film_id INT REFERENCES films(id) ON DELETE CASCADE,
                            actor_id INT REFERENCES actors(id) ON DELETE CASCADE,
                            PRIMARY KEY (film_id, actor_id)
);