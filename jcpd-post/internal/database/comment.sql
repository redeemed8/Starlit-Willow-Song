# 创建评论表 commentinfo
CREATE TABLE IF NOT EXISTS comment_infos
(
    id             INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at     DATETIME     NOT NULL,
    post_id        INT UNSIGNED NOT NULL,
    publisher_id   INT UNSIGNED NOT NULL,
    publisher_name VARCHAR(255) NOT NULL,
    body           TEXT         NOT NULL,
    INDEX cp (created_at, post_id)
) ENGINE = InnoDB;



