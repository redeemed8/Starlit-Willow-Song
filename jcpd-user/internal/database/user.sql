# 创建 userinfo 表
CREATE TABLE IF NOT EXISTS 5613_userinfo
(
    id          INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    phone       varchar(12) UNIQUE,
    username    varchar(30) UNIQUE,
    password    varchar(33),
    uuid        varchar(37) NOT NULL,
    sex         varchar(2),
    sign        text,
    friend_list text,
    group_list  text,
    created_at  timestamp DEFAULT CURRENT_TIMESTAMP
) ENGINE = InnoDB;

# 批量更新用户的所在群列表
UPDATE 5613_userinfo
SET group_list = REPLACE(group_list, ',2,', ',')
WHERE id IN (1);

