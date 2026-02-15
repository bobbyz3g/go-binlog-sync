USE gbs_test;

INSERT INTO users (name, email) VALUES ("carol", "carol@example.com");
UPDATE users SET email = "alice+1@example.com" WHERE id = 1;
DELETE FROM orders WHERE id = 2;
