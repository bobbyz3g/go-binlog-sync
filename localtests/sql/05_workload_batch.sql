USE gbs_test;

WITH RECURSIVE seq AS (
  SELECT 1 AS n
  UNION ALL
  SELECT n + 1 FROM seq WHERE n < 1000
)
INSERT INTO txn_tbl (note)
SELECT CONCAT("batch-", n) FROM seq;
