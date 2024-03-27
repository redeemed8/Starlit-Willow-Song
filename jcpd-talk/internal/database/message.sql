# 创建 message表
CREATE TABLE 1765_message
(
    id          INT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    sender_id   INT UNSIGNED,
    receiver_id INT UNSIGNED,
    content     TEXT,
    status      VARCHAR(1)
);

CREATE INDEX message_sender_id ON 1765_message (sender_id);
CREATE INDEX message_receiver_id ON 1765_message (receiver_id);
