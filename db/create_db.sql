create database ourroots;
create user ourroots with encrypted password 'password';
create user ourroots_schema with encrypted password 'password';
grant all privileges on database ourroots to ourroots_schema;
-- grant all privileges on database ourroots to ourroots;
