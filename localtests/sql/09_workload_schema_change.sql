USE gbs_test;

ALTER TABLE users ADD COLUMN nickname VARCHAR(64) NULL;
INSERT INTO users (name, email, nickname) VALUES ("dave", "dave@example.com", "d");
