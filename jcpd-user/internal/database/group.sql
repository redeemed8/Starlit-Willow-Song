# 创建 group 表
CREATE TABLE IF NOT EXISTS `5433_group`
(
    `id`             INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `group_name`     TEXT         NOT NULL,
    `group_post`     TEXT         NOT NULL,
    `lord_id`        INT UNSIGNED NOT NULL,
    `admin_ids`      TEXT,
    `member_ids`     TEXT,
    `cur_person_num` INT                   DEFAULT 0,
    `max_person_num` INT          NOT NULL DEFAULT 100,
    `created_at`     TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`     TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE = InnoDB;

CREATE INDEX `idx_lord_id` ON `5433_group` (`lord_id`);
