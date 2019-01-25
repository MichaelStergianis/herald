-- Music related
CREATE TABLE artists (
       id SERIAL PRIMARY KEY,
       artist_name VARCHAR NOT NULL
);

CREATE TABLE genres (
       id SERIAL PRIMARY KEY,
       genre_name VARCHAR NOT NULL
);

CREATE TABLE images (
       id SERIAL PRIMARY KEY,
       file_path VARCHAR NOT NULL
);

CREATE TABLE albums (
       id SERIAL PRIMARY KEY,

       artist INTEGER REFERENCES artists(id),

       release_year INTEGER,
       n_tracks INTEGER NOT NULL, -- number of songs
       title VARCHAR NOT NULL,
);

CREATE TABLE images_in_album (
       album_id REFERENCES albums(id),
       image_id REFERENCES images(id)
       primary_image BOOLEAN NOT NULL
);

CREATE TABLE songs (
       id SERIAL PRIMARY KEY,

       album INTEGER REFERENCES albums(id),

       fs_path VARCHAR UNIQUE NOT NULL,
       title VARCHAR NOT NULL,
       song_size INTEGER NOT NULL, -- bytes
       artist VARCHAR,
       duration INTEGER NOT NULL -- seconds
);

CREATE TABLE libraries (
       id SERIAL PRIMARY KEY,
       library_name VARCHAR UNIQUE NOT NULL,
       fs_path VARCHAR NOT NULL
);

CREATE TABLE songs_in_library (
       song_id INTEGER REFERENCES songs(id),
       library_id INTEGER REFERENCES libraries(id)
);


-- Settings related
CREATE TABLE preferences (
       id SERIAL PRIMARY KEY
);

CREATE TABLE users (
      id SERIAL PRIMARY KEY,

      preference_id INTEGER REFERENCES preferences(id),

      user_name VARCHAR UNIQUE NOT NULL,
      email VARCHAR NOT NULL,
      password VARCHAR NOT NULL
);
