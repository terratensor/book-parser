CREATE TABLE public.books
(
    id         bigint      NOT NULL,
    name       varchar,
    filename   varchar,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    deleted_at timestamptz NULL
--     CONSTRAINT books_pk PRIMARY KEY (uuid)
);

CREATE TABLE public.db_pg_books
(
    id         bigserial   NOT NULL,
    name       varchar,
    filename   varchar,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    deleted_at timestamptz NULL,
    CONSTRAINT db_pg_books_pkey PRIMARY KEY (id)
);

CREATE TABLE public.book_paragraphs
(
    id         bigint      NOT NULL,
    book_id    bigint      NOT NULL,
    book_name  text,
    text       text,
    position   int,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    deleted_at timestamptz NULL
--     CONSTRAINT book_paragraphs_pk PRIMARY KEY (uuid)
);

CREATE TABLE public.db_pg_paragraphs
(
    id         bigserial   NOT NULL,
    book_id    int8        NOT NULL,
    book_name  text,
    text       text,
    position   int,
    length     int,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    deleted_at timestamptz NULL,
    CONSTRAINT db_pg_paragraphs_pkey PRIMARY KEY (id)
);
