/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
SET time_zone = "+00:00";

CREATE DATABASE IF NOT EXISTS `authentication` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
USE `authentication`;

# Dump of table apps
# ------------------------------------------------------------
DROP TABLE IF EXISTS `apps`;

CREATE TABLE `apps` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `uuid` char(36) NOT NULL,
  `name` char(64) NOT NULL DEFAULT '',
  `start` timestamp NULL DEFAULT NULL,
  `end` timestamp NULL DEFAULT NULL,
  `count` bigint(20) DEFAULT NULL,
  `enterprise` char(36) NOT NULL DEFAULT '',
  `created_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `status` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uuid` (`uuid`),
  KEY `enterprise of app` (`enterprise`),
  CONSTRAINT `enterprise of app` FOREIGN KEY (`enterprise`) REFERENCES `enterprises` (`uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

LOCK TABLES `apps` WRITE;
/*!40000 ALTER TABLE `apps` DISABLE KEYS */;

INSERT INTO `apps` (`id`, `uuid`, `name`, `start`, `end`, `count`, `enterprise`, `created_time`, `status`)
VALUES
	(1,'csbot','csbot',NULL,NULL,NULL,'bb3e3925-f0ad-11e7-bd86-0242ac120003','2018-04-05 15:21:02',1);

/*!40000 ALTER TABLE `apps` ENABLE KEYS */;
UNLOCK TABLES;


# Dump of table enterprises
# ------------------------------------------------------------

DROP TABLE IF EXISTS `enterprises`;

CREATE TABLE `enterprises` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `uuid` char(36) NOT NULL,
  `name` varchar(64) NOT NULL DEFAULT '',
  `admin_user` char(36) NOT NULL DEFAULT '',
  `created_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uuid` (`uuid`),
  KEY `admin of enterprise` (`admin_user`),
  CONSTRAINT `admin of enterprise` FOREIGN KEY (`admin_user`) REFERENCES `users` (`uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

LOCK TABLES `enterprises` WRITE;
/*!40000 ALTER TABLE `enterprises` DISABLE KEYS */;

INSERT INTO `enterprises` (`id`, `uuid`, `name`, `admin_user`, `created_time`)
VALUES
	(1,'bb3e3925-f0ad-11e7-bd86-0242ac120003','emotibot','4b21158a-3953-11e8-8a71-0242ac110003','2018-04-05 15:21:02');

/*!40000 ALTER TABLE `enterprises` ENABLE KEYS */;
UNLOCK TABLES;


# Dump of table modules
# ------------------------------------------------------------

DROP TABLE IF EXISTS `modules`;

CREATE TABLE `modules` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `code` char(36) NOT NULL DEFAULT '',
  `name` varchar(36) NOT NULL DEFAULT '',
  `enterprise` char(36) NOT NULL DEFAULT '',
  `cmd_list` char(64) NOT NULL,
  `discription` varchar(200) NOT NULL DEFAULT '',
  `created_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `status` tinyint(1) NOT NULL DEFAULT '1',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

LOCK TABLES `modules` WRITE;
/*!40000 ALTER TABLE `modules` DISABLE KEYS */;

INSERT INTO `modules` (`code`, `name`, `enterprise`, `cmd_list`, `discription`, `created_time`, `status`)
VALUES
	('statistic_dash', '', '', 'view', '', '2018-04-12 14:47:45', 1),
	('ssm', '', '', 'view,edit', '', '2018-05-15 23:04:37', 1),
	('qa_label', '', '', 'view,edit', '', '2018-05-15 23:05:01', 1),
	('task_engine', '', '', 'edit', '', '2018-05-15 22:59:01', 1),
	('wordbank', '', '', 'view,edit,create,delete,export,import', '', '2018-04-12 14:47:45', 1),
	('statistic_daily', '', '', 'view,export', '', '2018-04-12 14:47:45', 1),
	('statistic_analysis', '', '', 'view,export', '', '2018-05-15 23:05:25', 1),
	('statistic_audit', '', '', 'view,export', '', '2018-04-12 14:47:45', 1),
	('robot_profile', '', '', 'view,edit', '', '2018-04-12 14:47:45', 1),
	('robot_chat_skill', '', '', 'view,edit', '', '2018-04-12 14:47:45', 1),
	('robot_function', '', '', 'view,edit', '', '2018-05-15 23:06:30', 1),
	('robot_command', '', '', 'view,edit', '', '2018-05-15 23:06:41', 1),
	('management', '', '', 'edit', '', '2018-04-12 14:47:45', 1);

/*!40000 ALTER TABLE `modules` ENABLE KEYS */;
UNLOCK TABLES;


# Dump of table privileges
# ------------------------------------------------------------

DROP TABLE IF EXISTS `privileges`;

CREATE TABLE `privileges` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `role` bigint(20) NOT NULL,
  `module` bigint(20) NOT NULL,
  `cmd_list` char(64) NOT NULL DEFAULT '',
  `created_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `id of role` (`role`),
  KEY `id of module` (`module`),
  CONSTRAINT `id of module` FOREIGN KEY (`module`) REFERENCES `modules` (`id`),
  CONSTRAINT `id of role` FOREIGN KEY (`role`) REFERENCES `roles` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;



# Dump of table roles
# ------------------------------------------------------------

DROP TABLE IF EXISTS `roles`;

CREATE TABLE `roles` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `uuid` char(36) NOT NULL,
  `name` char(36) NOT NULL,
  `enterprise` char(36) NOT NULL DEFAULT '',
  `discription` varchar(200) NOT NULL DEFAULT '',
  `created_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `uuid` (`uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;



# Dump of table user_column
# ------------------------------------------------------------

DROP TABLE IF EXISTS `user_column`;

CREATE TABLE `user_column` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `column` char(32) NOT NULL DEFAULT '',
  `display_name` varchar(64) NOT NULL DEFAULT '',
  `enterprise` char(36) NOT NULL DEFAULT '',
  `note` varchar(64) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  KEY `enterprise of custom column` (`enterprise`),
  CONSTRAINT `enterprise of custom column` FOREIGN KEY (`enterprise`) REFERENCES `enterprises` (`uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

LOCK TABLES `user_column` WRITE;
/*!40000 ALTER TABLE `user_column` DISABLE KEYS */;

INSERT INTO `user_column` (`id`, `column`, `display_name`, `enterprise`, `note`)
VALUES
	(1,'custom1','自訂屬性1','bb3e3925-f0ad-11e7-bd86-0242ac120003','示範自訂屬性效果');

/*!40000 ALTER TABLE `user_column` ENABLE KEYS */;
UNLOCK TABLES;


# Dump of table user_info
# ------------------------------------------------------------

DROP TABLE IF EXISTS `user_info`;

CREATE TABLE `user_info` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `user_id` char(36) NOT NULL DEFAULT '',
  `column_id` bigint(64) NOT NULL,
  `value` text NOT NULL,
  PRIMARY KEY (`id`),
  KEY `user of info` (`user_id`),
  KEY `column of info` (`column_id`),
  CONSTRAINT `column of info` FOREIGN KEY (`column_id`) REFERENCES `user_column` (`id`),
  CONSTRAINT `user of info` FOREIGN KEY (`user_id`) REFERENCES `users` (`uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

LOCK TABLES `user_info` WRITE;
/*!40000 ALTER TABLE `user_info` DISABLE KEYS */;

INSERT INTO `user_info` (`id`, `user_id`, `column_id`, `value`)
VALUES
	(1,'4b21158a-3953-11e8-8a71-0242ac110003',1,'custom_value1');

/*!40000 ALTER TABLE `user_info` ENABLE KEYS */;
UNLOCK TABLES;


# Dump of table users
# ------------------------------------------------------------

DROP TABLE IF EXISTS `users`;

CREATE TABLE `users` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `uuid` char(36) NOT NULL DEFAULT '',
  `display_name` varchar(64) NOT NULL DEFAULT '',
  `user_name` char(32) NOT NULL DEFAULT '',
  `email` char(255) NOT NULL DEFAULT '',
  `enterprise` char(36) NOT NULL DEFAULT '',
  `type` tinyint(1) unsigned NOT NULL DEFAULT '2',
  `password` char(32) NOT NULL DEFAULT '',
  `role` char(36) NOT NULL DEFAULT '',
  `created_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `status` tinyint(1) NOT NULL DEFAULT '1',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uuid` (`uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

LOCK TABLES `users` WRITE;
/*!40000 ALTER TABLE `users` DISABLE KEYS */;

INSERT INTO `users` (`id`, `uuid`, `display_name`, `user_name`, `email`, `enterprise`, `type`, `password`, `role`, `created_time`, `status`)
VALUES
	(1,'4b21158a-3953-11e8-8a71-0242ac110003','CSBOT','csbotadmin','csbotadmin@emotibot.com','bb3e3925-f0ad-11e7-bd86-0242ac120003',1,'ac04367d3155bb651df2e4220bdb8303','','2018-04-05 15:21:54',1);

/*!40000 ALTER TABLE `users` ENABLE KEYS */;
UNLOCK TABLES;



/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;
/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
