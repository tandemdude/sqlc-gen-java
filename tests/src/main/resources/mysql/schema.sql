-- taken from sqlc-gen-kotlin, booktest --
CREATE TABLE authors (
    author_id integer NOT NULL AUTO_INCREMENT PRIMARY KEY,
    name text NOT NULL
) ENGINE=InnoDB;

CREATE INDEX authors_name_idx ON authors(name(255));

CREATE TABLE books (
    book_id integer NOT NULL AUTO_INCREMENT PRIMARY KEY,
    author_id integer NOT NULL,
    isbn varchar(255) NOT NULL DEFAULT '' UNIQUE,
    book_type ENUM('FICTION', 'NONFICTION') NOT NULL DEFAULT 'FICTION',
    title text NOT NULL,
    yr integer NOT NULL DEFAULT 2000,
    available datetime NOT NULL DEFAULT NOW(),
    tags text NOT NULL
) ENGINE=InnoDB;

CREATE INDEX books_title_idx ON books(title(255), yr);
-- end --

CREATE TABLE instant_test (
    t_id integer NOT NULL AUTO_INCREMENT PRIMARY KEY,
    ts timestamp DEFAULT NULL
) ENGINE=InnoDB;
