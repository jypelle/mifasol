-- +migrate Up

-- Album

create table album
(
    album_id    text    not null primary key,
    creation_ts integer not null,
    update_ts   integer not null,
    name        text    not null
);

create table deleted_album
(
    album_id  text    not null primary key,
    delete_ts integer not null
);

-- Artist

create table artist
(
    artist_id   text    not null primary key,
    creation_ts integer not null,
    update_ts   integer not null,
    name        text    not null
);

create table deleted_artist
(
    artist_id text    not null primary key,
    delete_ts integer not null
);

create table artist_song
(
    artist_id text not null,
    song_id   text not null,
    primary key (artist_id, song_id)
);

-- Favorite playlist

create table favorite_playlist
(
    user_id     text    not null,
    playlist_id text    not null,
    update_ts   integer not null,
    primary key (user_id, playlist_id)
);

create table deleted_favorite_playlist
(
    user_id     text    not null,
    playlist_id text    not null,
    delete_ts   integer not null,
    primary key (user_id, playlist_id)
);

-- Favorite song

create table favorite_song
(
    user_id   text    not null,
    song_id   text    not null,
    update_ts integer not null,
    primary key (user_id, song_id)
);

create table deleted_favorite_song
(
    user_id   text    not null,
    song_id   text    not null,
    delete_ts integer not null,
    primary key (user_id, song_id)
);

-- Playlist

create table playlist
(
    playlist_id       text    not null,
    creation_ts       integer not null,
    update_ts         integer not null,
    content_update_ts integer not null,
    name              text    not null,
    primary key (playlist_id)
);

create table playlist_song
(
    playlist_id text    not null,
    position    integer not null,
    song_id     text    not null,
    primary key (playlist_id, position)
);

create table playlist_owned_user
(
    playlist_id text not null,
    user_id     text not null,
    primary key (playlist_id, user_id)
);

create table deleted_playlist
(
    playlist_id text    not null,
    delete_ts   integer not null,
    primary key (playlist_id)
);

-- Song

create table song
(
    song_id          text    not null primary key,
    creation_ts      integer not null,
    update_ts        integer not null,
    name             text    not null,
    format           integer not null,
    size             integer not null,
    bit_depth        integer not null,
    publication_year integer,
    album_id         text    not null,
    track_number     integer,
    explicit_fg      bool    not null
);

create table deleted_song
(
    song_id   text    not null,
    delete_ts integer not null,
    primary key (song_id)
);

-- User

create table user
(
    user_id          text    not null primary key,
    creation_ts      integer not null,
    update_ts        integer not null,
    name             text    not null,
    hide_explicit_fg bool    not null,
    admin_fg         bool    not null,
    password         text    not null
);

create unique index user_name_uindex on user (name);

create table deleted_user
(
    user_id   text    not null,
    delete_ts integer not null,
    primary key (user_id)
);
