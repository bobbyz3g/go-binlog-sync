USE gbs_test;

INSERT INTO users (name, email) VALUES
  ("alice", "alice@example.com"),
  ("bob", "bob@example.com");

INSERT INTO orders (user_id, amount, status) VALUES
  (1, 19.99, "paid"),
  (2, 5.50, "pending");

INSERT INTO types (c_int, c_big, c_varchar, c_text, c_json, c_decimal, c_bool, c_blob, c_date, c_time, c_ts)
VALUES
  (1, 100, "alpha", "text", JSON_OBJECT("k", "v"), 10.50, 1, X'616263', '2024-01-01', '12:34:56', '2024-01-01 12:34:56');

USE gbs_filter;
INSERT INTO allow_table (note) VALUES ("allow-1"), ("allow-2");
INSERT INTO deny_table (note) VALUES ("deny-1"), ("deny-2");

USE gbs_reserved;
INSERT INTO `order` (note) VALUES ("o1");
INSERT INTO `user-log` (note) VALUES ("u1");
