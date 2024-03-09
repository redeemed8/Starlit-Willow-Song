# 创建 like_infos表
CREATE TABLE `6481_likeinfo`
(
    `id`      int(11)          NOT NULL AUTO_INCREMENT,
    `user_id` int(10) UNSIGNED NOT NULL,
    `post_id` int(10) UNSIGNED NOT NULL,
    PRIMARY KEY (`id`),
    KEY `likeid` (`user_id`, `post_id`)
) ENGINE = InnoDB;


select id from 6481_likeinfo where user_id = 11 and post_id in (11, 22, 33);

