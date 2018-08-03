-- Music related
CREATE TABLE artists (
       id SERIAL PRIMARY KEY,
       artist_name VARCHAR NOT NULL
);

CREATE TABLE genres (
       id SERIAL PRIMARY KEY,
       genre_name VARCHAR NOT NULL
);

CREATE TABLE albums (
       id SERIAL PRIMARY KEY,

       artist INTEGER REFERENCES artists(id),

       album_size INTEGER NOT NULL,
       title VARCHAR NOT NULL,
       duration INTERVAL DAY TO SECOND
);

CREATE TABLE songs (
       id SERIAL PRIMARY KEY,

       album INTEGER REFERENCES albums(id),

       fs_path VARCHAR UNIQUE NOT NULL,
       title VARCHAR NOT NULL,
       song_size INTEGER NOT NULL,
       artist VARCHAR,
       duration INTERVAL DAY TO SECOND
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
