USE gbs_test;

START TRANSACTION;
INSERT INTO txn_tbl (note) VALUES ("txn-1"), ("txn-2");
UPDATE orders SET status = "paid" WHERE id = 1;
COMMIT;
