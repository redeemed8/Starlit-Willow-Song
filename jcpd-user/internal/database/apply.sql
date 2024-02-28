# 创建 apply 表
CREATE TABLE IF NOT EXISTS 5176_apply
(
    id             INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `created_at`   TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`   TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `sender_id`    INT UNSIGNED NOT NULL,
    `receiver_id`  INT UNSIGNED NOT NULL,
    `apply_type`   TEXT         NOT NULL,
    `status`       TEXT         NOT NULL,
    `introduction` TEXT         NOT NULL
) ENGINE = InnoDB;

# 采用4个索引而不是联合索引，是因为这个表要经常写入和修改
# 而且数据量大时，联合索引占的内存可能比普通索引多，而且普通索引使用更灵活
CREATE INDEX `idx_field1` ON `5176_apply` (`sender_id`);
CREATE INDEX `idx_field2` ON `5176_apply` (`receiver_id`);
CREATE INDEX `idx_field3` ON `5176_apply` (`apply_type`);
CREATE INDEX `idx_field4` ON `5176_apply` (`status`);



