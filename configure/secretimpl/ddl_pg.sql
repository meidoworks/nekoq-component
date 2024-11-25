create table secret_level1
(
    key_id        bigserial    not null,
    key_name      varchar(256) not null,
    key_version   bigint       not null,
    key_status    bigint       not null,
    use_key_id    varchar(64)  not null,
    key_encrypted bytea        not null,
    expired_time  bigint       not null,
    time_created  bigint       not null,
    time_update   bigint       not null,
    primary key (key_id)
);

create unique index on secret_level1 (key_name, key_version);

comment on column secret_level1.key_status is '0-valid, 1-decrypt_only';

create table secret_level2
(
    key_id        bigserial    not null,
    key_name      varchar(256) not null,
    key_version   bigint       not null,
    key_status    bigint       not null,
    use_key_id    bigint       not null,
    key_type      varchar(64)  not null,
    key_encrypted bytea        not null,
    expired_time  bigint       not null,
    time_created  bigint       not null,
    time_update   bigint       not null,
    primary key (key_id)
);


create unique index on secret_level2 (key_name, key_version);

comment on column secret_level2.key_status is '0-valid, 1-decrypt_only';

comment on column secret_level2.key_type is 'keyset-KeySet, aes256-AES256, etc.';

