Create Table: CREATE TABLE `class` (
   `id` int NOT NULL AUTO_INCREMENT,
   `class_name` varchar(200) NOT NULL DEFAULT '班级名称',
    `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;


Create Table: CREATE TABLE `student` (
     `id` int NOT NULL AUTO_INCREMENT,
     `name` varchar(200) NOT NULL DEFAULT '名字',
    `sex` tinyint NOT NULL DEFAULT '0' COMMENT '性别',
    `class_id` int NOT NULL DEFAULT '0' COMMENT '班级id',
    `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
