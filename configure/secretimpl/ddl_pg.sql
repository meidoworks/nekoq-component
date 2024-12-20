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
    time_updated   bigint       not null,
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
    time_updated   bigint       not null,
    primary key (key_id)
);

create unique index on secret_level2 (key_name, key_version);

comment on column secret_level2.key_status is '0-valid, 1-decrypt_only';

comment on column secret_level2.key_type is 'keyset-KeySet, aes256-AES256, etc.';


create table secret_cert
(
    cert_id           varchar(256) not null,
    cert_type         bigint       not null,
    cert_name         varchar(256) not null,
    cert_version      bigint       not null,
    cert_status       bigint       not null,
    parent_ca_cert_id varchar(256) not null,
    cert_key_level    bigint       not null,
    cert_key_name     varchar(256) not null,
    cert_content      bytea        not null,
    expired_time      bigint       not null,
    time_created      bigint       not null,
    time_updated      bigint       not null,
    primary key (cert_id)
);

create unique index on secret_cert (cert_type, cert_name, cert_version);

comment on column secret_cert.cert_status is '0-valid, 1-archived';

create sequence cert_id_seq increment by 16 minvalue 1 maxvalue 9223372036854775807 start 1 cache 1 no cycle;
