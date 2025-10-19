create type verification_status as enum ('new', 'verified', 'in_development');

alter type verification_status owner to postgres;

create type source_website as enum ('2gis', 'google_maps', 'instagram', 'olx', 'yandex', 'booking', 'manual');

alter type source_website owner to postgres;

create table accommodations
(
    id                  serial
        primary key,
    name                varchar(500)   not null,
    latitude            numeric(10, 8),
    longitude           numeric(11, 8),
    address             text,
    phone               varchar(50),
    email               varchar(100),
    social_media_links  jsonb,
    website_url         text,
    social_media_page   text,
    service_description text,
    room_count          integer,
    capacity            integer,
    price_range_min     numeric(10, 2),
    price_range_max     numeric(10, 2),
    price_currency      varchar(3)               default 'KZT'::character varying,
    photos              jsonb,
    rating              numeric(3, 2),
    review_count        integer                  default 0,
    reviews             jsonb,
    amenities           jsonb,
    verification_status verification_status      default 'new'::verification_status,
    last_updated        timestamp with time zone default CURRENT_TIMESTAMP,
    source_website      source_website not null,
    source_url          text,
    external_id         varchar(100),
    created_at          timestamp with time zone default CURRENT_TIMESTAMP,
    deleted_at          timestamp with time zone,
    accommodation_type  varchar(50),
    constraint unique_source_external_id
        unique (source_website, external_id)
);

alter table accommodations
    owner to postgres;

create index idx_accommodations_source
    on accommodations (source_website);

create index idx_accommodations_status
    on accommodations (verification_status);

create index idx_accommodations_created_at
    on accommodations (created_at);

create index idx_accommodations_last_updated
    on accommodations (last_updated);

create index idx_accommodations_location
    on accommodations (latitude, longitude);

create table parsing_logs
(
    id               serial
        primary key,
    source_website   source_website not null,
    operation        varchar(20)    not null, -- 'insert' or 'update'
    status           varchar(20)    not null default 'pending',
    error_message    text,
    external_id      varchar(100),             -- external ID from source
    started_at       timestamp with time zone default CURRENT_TIMESTAMP,
    completed_at     timestamp with time zone,
    duration_ms      integer
);

alter table parsing_logs
    owner to postgres;

create function update_last_updated_column() returns trigger
    language plpgsql
as
$$
BEGIN
    NEW.last_updated = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$;

alter function update_last_updated_column() owner to postgres;

create trigger update_accommodations_last_updated
    before update
    on accommodations
    for each row
execute procedure update_last_updated_column();

