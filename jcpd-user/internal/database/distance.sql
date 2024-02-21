# 创建 pointinfo 表, 使用空间索引优化
create table 5918_pointinfo
(
    id    int unsigned PRIMARY KEY,
    point geometry NOT NULL,
    SPATIAL INDEX idx_location (point)
);

# 范围距离查询
# 这里我们采用子查询的方式
# 因为 直接查距离的话，会导致两次计算，而使用 having的话，又会导致没有join的情况下having效率低
SELECT id, distance
FROM (SELECT *, ST_Distance_Sphere(POINT(0, 0), point) AS distance
      FROM 5918_pointinfo) AS distances
WHERE distance <= 1000000000
ORDER BY distance
LIMIT 10;
