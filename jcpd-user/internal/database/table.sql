# 创建 userinfo 表
CREATE TABLE IF NOT EXISTS 5613_userinfo
(
    id         int AUTO_INCREMENT PRIMARY KEY,
    phone      varchar(12) UNIQUE,
    username   varchar(30) UNIQUE,
    password   varchar(33),
    uuid       varchar(37) NOT NULL,
    sex        varchar(2),
    sign       longtext,
    created_at timestamp DEFAULT CURRENT_TIMESTAMP
) ENGINE = InnoDB;

