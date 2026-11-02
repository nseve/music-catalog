CREATE SCHEMA mc;

CREATE TABLE mc.genre (
    id          SERIAL          PRIMARY KEY,
    name        VARCHAR(255)    NOT NULL UNIQUE
);

CREATE TABLE mc.artist (
    id          SERIAL          PRIMARY KEY,
    name        VARCHAR(255)    NOT NULL,
    country     CHAR(2)
);

CREATE TABLE mc.album (
    id              SERIAL          PRIMARY KEY,
    artist_id       INTEGER         NOT NULL REFERENCES mc.artist(id) ON DELETE RESTRICT,
    title           VARCHAR(255)    NOT NULL,
    album_type      VARCHAR(20)     NOT NULL DEFAULT 'album',
    release_date    DATE            NOT NULL,
    cover_path      VARCHAR(255),

    CONSTRAINT album_type_check CHECK (album_type IN ('album', 'single', 'ep', 'compilation'))
);

CREATE TABLE mc.track (
    id              SERIAL          PRIMARY KEY,
    album_id        INTEGER         NOT NULL REFERENCES mc.album(id) ON DELETE CASCADE,
    genre_id        INTEGER                  REFERENCES mc.genre(id) ON DELETE SET NULL,
    title           VARCHAR(255)    NOT NULL,
    duration        INTEGER         NOT NULL CHECK (duration > 0),
    track_number    INTEGER         NOT NULL CHECK (track_number > 0),
    is_explicit     BOOLEAN         NOT NULL DEFAULT FALSE,

    UNIQUE (album_id, track_number)
);

CREATE INDEX idx_album_artist_id ON mc.album(artist_id);
CREATE INDEX idx_track_genre_id  ON mc.track(genre_id);