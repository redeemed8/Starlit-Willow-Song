# 创建 like_infos表
CREATE TABLE `6481_likeinfo`
(
    `id`      int(11)          NOT NULL AUTO_INCREMENT,
    `user_id` int(10) UNSIGNED NOT NULL,
    `post_id` int(10) UNSIGNED NOT NULL,
    PRIMARY KEY (`id`),
    KEY `likeid` (`user_id`, `post_id`)
) ENGINE = InnoDB;


# 可以使用联合索引加速
select id
from 6481_likeinfo
where user_id = 11
  and post_id in (11, 22, 33);

# 批量插入
insert into 6481_likeinfo
    (user_id, post_id)
values (1, 2),
       (2, 4),
       (3, 5);

# 批量删除
DELETE
FROM 6481_likeinfo
WHERE (user_id = 1 AND post_id = 2)
   OR (user_id = 5 AND post_id = 19);
