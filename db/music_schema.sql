-- Music related
CREATE SCHEMA IF NOT EXISTS music;

CREATE TABLE IF NOT EXISTS music.libraries (
       id SERIAL PRIMARY KEY,
       name VARCHAR UNIQUE NOT NULL,
       fs_path VARCHAR UNIQUE NOT NULL
);

CREATE INDEX IF NOT EXISTS ix_libraries ON music.libraries (id, name);

CREATE TABLE IF NOT EXISTS music.artists (
       id SERIAL PRIMARY KEY,
       name VARCHAR NOT NULL,
       fs_path VARCHAR UNIQUE NOT NULL
);

CREATE INDEX IF NOT EXISTS ix_artists ON music.artists (id, name);

CREATE TABLE IF NOT EXISTS music.genres (
       id SERIAL PRIMARY KEY,
       name VARCHAR UNIQUE NOT NULL
);

CREATE INDEX IF NOT EXISTS ix_genres ON music.genres (id, name);

CREATE TABLE IF NOT EXISTS music.images (
       id SERIAL PRIMARY KEY,
       fs_path VARCHAR UNIQUE NOT NULL
);

CREATE INDEX IF NOT EXISTS ix_images ON music.images (id);

CREATE TABLE IF NOT EXISTS music.albums (
       id SERIAL PRIMARY KEY,

       artist INTEGER REFERENCES music.artists(id),

       release_year INTEGER,
       n_tracks INTEGER NOT NULL, -- number of songs
       n_disks INTEGER NOT NULL,  -- number of disks
       title VARCHAR NOT NULL,
       fs_path VARCHAR UNIQUE NOT NULL,
       duration DOUBLE PRECISION -- seconds
);

CREATE INDEX IF NOT EXISTS ix_albums ON music.albums (id, title);

CREATE TABLE IF NOT EXISTS music.images_in_album (
       album_id INTEGER REFERENCES music.albums(id),
       image_id INTEGER REFERENCES music.images(id),
       primary_image BOOLEAN NOT NULL
);

CREATE TABLE IF NOT EXISTS music.songs (
       id SERIAL PRIMARY KEY,

       album INTEGER REFERENCES music.albums(id),
       genre INTEGER REFERENCES music.genres(id),

       fs_path VARCHAR UNIQUE NOT NULL,
       title VARCHAR NOT NULL,
       track INTEGER NOT NULL,
       num_tracks INTEGER NOT NULL,
       disk INTEGER NOT NULL,
       num_disks INTEGER NOT NULL,
       song_size INTEGER NOT NULL,           -- bytes
       duration DOUBLE PRECISION NOT NULL,   -- seconds
       artist VARCHAR
);

CREATE INDEX IF NOT EXISTS ix_songs ON music.songs (id, title);

CREATE TABLE IF NOT EXISTS music.songs_in_library (
       song_id INTEGER REFERENCES music.songs(id),
       library_id INTEGER REFERENCES music.libraries(id)
);


