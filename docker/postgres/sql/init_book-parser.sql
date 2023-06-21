CREATE TABLE public.books
(
    id         bigint   NOT NULL,
    name       varchar,
    filename   varchar,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    deleted_at timestamptz NULL
--     CONSTRAINT books_pk PRIMARY KEY (uuid)
);

CREATE TABLE public.book_paragraphs
(
    id         bigint         NOT NULL,
    book_id    bigint         NOT NULL,
    text       text,
    position   int,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    deleted_at timestamptz NULL
--     CONSTRAINT book_paragraphs_pk PRIMARY KEY (uuid)
);
