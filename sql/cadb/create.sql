

\set ON_ERROR_STOP on

-- Database: cadb
SELECT
    EXISTS(SELECT datname  FROM pg_catalog.pg_database WHERE datname = 'cadb') as cadb_exists \gset

\if :cadb_exists
\echo 'cadb already exists!'
\q
\endif

-- template0: see https://blog.dbi-services.com/what-the-hell-are-these-template0-and-template1-databases-in-postgresql/
CREATE DATABASE cadb
    WITH
    OWNER = postgres
    ENCODING = 'UTF8'
    LC_COLLATE = 'en_US.UTF-8'
    LC_CTYPE = 'en_US.UTF-8'
    TEMPLATE template0
    CONNECTION LIMIT = -1;
