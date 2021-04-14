

\set ON_ERROR_STOP on

-- Database: trustydb
SELECT
    EXISTS(SELECT datname  FROM pg_catalog.pg_database WHERE datname = 'trustydb') as trustydb_exists \gset

\if :trustydb_exists
\echo 'trustydb already exists!'
\q
\endif

-- template0: see https://blog.dbi-services.com/what-the-hell-are-these-template0-and-template1-databases-in-postgresql/
CREATE DATABASE trustydb
    WITH
    OWNER = postgres
    ENCODING = 'UTF8'
    LC_COLLATE = 'en_US.UTF-8'
    LC_CTYPE = 'en_US.UTF-8'
    TEMPLATE template0
    CONNECTION LIMIT = -1;
