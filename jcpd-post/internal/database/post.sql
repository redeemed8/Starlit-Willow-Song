# 创建 postinfo表，字段虽采用text类型，也会进行参数对应的检查，将字符串控制在合理的范围内，不会出现很大的字符串
CREATE TABLE 3491_postinfo
(
    id             INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at     DATETIME     NOT NULL,
    updated_at     DATETIME     NOT NULL,
    title          TEXT         NOT NULL,
    topic_tag      VARCHAR(60)  NOT NULL,
    body           TEXT         NOT NULL,
    publisher_id   INT UNSIGNED NOT NULL,
    publisher_name VARCHAR(31)  NOT NULL,
    likes          INT        DEFAULT 0,
    review_status  VARCHAR(1) DEFAULT '0',
    reason         TEXT
) ENGINE = InnoDB;

# 普通索引
CREATE INDEX ttt ON 3491_postinfo (topic_tag);
CREATE INDEX ppp ON 3491_postinfo (publisher_id);

# 联合索引
CREATE INDEX like_time ON 3491_postinfo (likes, updated_at);
