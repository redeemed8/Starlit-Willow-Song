# 创建 message_counter表
CREATE TABLE 1455_message_counter
(
    id           INT AUTO_INCREMENT PRIMARY KEY,
    send_to_rece VARCHAR(22),
    unread_num   SMALLINT UNSIGNED DEFAULT 0,
    INDEX send_rece (send_to_rece, unread_num)
);
