# ************************************************************
# Sequel Pro SQL dump
# Version 4541
#
# http://www.sequelpro.com/
# https://github.com/sequelpro/sequelpro
#
# Host: 127.0.0.1 (MySQL 5.7.22)
# Database: anothertest
# Generation Time: 2018-06-12 09:32:07 +0000
# ************************************************************


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;


# Dump of table app_group
# ------------------------------------------------------------

DROP TABLE IF EXISTS `app_group`;

CREATE TABLE `app_group` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `robot_group` char(32) NOT NULL DEFAULT '',
  `app` char(32) NOT NULL,
  `created_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `app exist` (`robot_group`),
  KEY `app exists` (`app`),
  CONSTRAINT `app exists` FOREIGN KEY (`app`) REFERENCES `apps` (`uuid`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `group exist` FOREIGN KEY (`robot_group`) REFERENCES `robot_groups` (`uuid`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;



# Dump of table apps
# ------------------------------------------------------------

DROP TABLE IF EXISTS `apps`;

CREATE TABLE `apps` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `uuid` char(32) NOT NULL DEFAULT '',
  `name` char(64) NOT NULL DEFAULT '',
  `status` tinyint(1) NOT NULL DEFAULT '0',
  `description` varchar(200) NOT NULL DEFAULT '',
  `start` timestamp NULL DEFAULT NULL,
  `end` timestamp NULL DEFAULT NULL,
  `count` bigint(20) DEFAULT NULL,
  `enterprise` char(32) NOT NULL DEFAULT '',
  `created_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uuid` (`uuid`),
  KEY `enterprise of app` (`enterprise`),
  CONSTRAINT `app is type of machine` FOREIGN KEY (`uuid`) REFERENCES `machine` (`uuid`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `enterprise of app` FOREIGN KEY (`enterprise`) REFERENCES `enterprises` (`uuid`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

LOCK TABLES `apps` WRITE;
/*!40000 ALTER TABLE `apps` DISABLE KEYS */;

INSERT INTO `apps` (`id`, `uuid`, `name`, `status`, `description`, `start`, `end`, `count`, `enterprise`, `created_time`)
VALUES
	(1,'csbot','csbot',1,'',NULL,NULL,NULL,'bb3e3925f0ad11e7bd860242ac120003','2018-06-12 17:25:44');

/*!40000 ALTER TABLE `apps` ENABLE KEYS */;
UNLOCK TABLES;


# Dump of table columns
# ------------------------------------------------------------

DROP TABLE IF EXISTS `columns`;

CREATE TABLE `columns` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `column` char(32) NOT NULL DEFAULT '',
  `display_name` varchar(64) NOT NULL DEFAULT '',
  `enterprise` char(32) NOT NULL DEFAULT '',
  `note` varchar(64) NOT NULL DEFAULT '',
  `ref` char(32) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  KEY `enterprise of custom column` (`enterprise`),
  CONSTRAINT `enterprise of column` FOREIGN KEY (`enterprise`) REFERENCES `enterprises` (`uuid`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;



# Dump of table enterprises
# ------------------------------------------------------------

DROP TABLE IF EXISTS `enterprises`;

CREATE TABLE `enterprises` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `uuid` char(32) NOT NULL DEFAULT '',
  `name` varchar(64) NOT NULL DEFAULT '',
  `description` varchar(200) NOT NULL DEFAULT '',
  `created_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uuid` (`uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

LOCK TABLES `enterprises` WRITE;
/*!40000 ALTER TABLE `enterprises` DISABLE KEYS */;

INSERT INTO `enterprises` (`id`, `uuid`, `name`, `description`, `created_time`)
VALUES
	(1,'bb3e3925f0ad11e7bd860242ac120003','emotibot','','2018-06-12 17:23:15');

/*!40000 ALTER TABLE `enterprises` ENABLE KEYS */;
UNLOCK TABLES;


# Dump of table robot_groups
# ------------------------------------------------------------

DROP TABLE IF EXISTS `robot_groups`;

CREATE TABLE `robot_groups` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `uuid` char(32) NOT NULL DEFAULT '',
  `name` varchar(64) NOT NULL DEFAULT '',
  `enterprise` char(32) NOT NULL DEFAULT '',
  `status` tinyint(1) NOT NULL DEFAULT '1',
  `created_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uuid` (`uuid`),
  KEY `enterprise of group` (`enterprise`),
  CONSTRAINT `enterprise of group` FOREIGN KEY (`enterprise`) REFERENCES `enterprises` (`uuid`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `group is type of machine` FOREIGN KEY (`uuid`) REFERENCES `machine` (`uuid`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;



# Dump of table human
# ------------------------------------------------------------

DROP TABLE IF EXISTS `human`;

CREATE TABLE `human` (
  `id` int(20) NOT NULL AUTO_INCREMENT,
  `uuid` char(32) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uuid` (`uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;



# Dump of table machine
# ------------------------------------------------------------

DROP TABLE IF EXISTS `machine`;

CREATE TABLE `machine` (
  `id` int(20) NOT NULL AUTO_INCREMENT,
  `uuid` char(32) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uuid` (`uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;



# Dump of table modules
# ------------------------------------------------------------

DROP TABLE IF EXISTS `modules`;

CREATE TABLE `modules` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `code` char(32) NOT NULL DEFAULT '',
  `name` varchar(36) NOT NULL DEFAULT '',
  `enterprise` char(32) DEFAULT '',
  `cmd_list` char(64) NOT NULL,
  `description` varchar(200) NOT NULL DEFAULT '',
  `status` tinyint(1) NOT NULL DEFAULT '1',
  `created_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `enterprise of modules` (`enterprise`),
  CONSTRAINT `enterprise of modules` FOREIGN KEY (`enterprise`) REFERENCES `enterprises` (`uuid`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

/*!40000 ALTER TABLE `modules` DISABLE KEYS */;
INSERT INTO `modules` (`id`, `code`, `name`, `enterprise`, `cmd_list`, `description`, `status`, `created_time`)
VALUES
	(1, 'statistic_dash', '', NULL, 'view', '', 1, '2018-04-12 22:47:45'),
	(2, 'ssm', '', NULL, 'view,edit,export,import', '', 1, '2018-05-16 07:04:37'),
	(3, 'qa_label', '', NULL, 'view,edit', '', 1, '2018-05-16 07:05:01'),
	(4, 'task_engine', '', NULL, 'edit', '', 1, '2018-05-16 06:59:01'),
	(5, 'wordbank', '', NULL, 'view,edit,create,delete,export,import', '', 1, '2018-04-12 22:47:45'),
	(6, 'statistic_daily', '', NULL, 'view,export', '', 1, '2018-04-12 22:47:45'),
	(7, 'statistic_analysis', '', NULL, 'view,export', '', 1, '2018-05-16 07:05:25'),
	(8, 'statistic_audit', '', NULL, 'view,export', '', 1, '2018-04-12 22:47:45'),
	(9, 'robot_profile', '', NULL, 'view,edit', '', 1, '2018-04-12 22:47:45'),
	(10, 'robot_chat_skill', '', NULL, 'view,edit', '', 1, '2018-04-12 22:47:45'),
	(11, 'robot_function', '', NULL, 'view,edit', '', 1, '2018-05-16 07:06:30'),
	(12, 'robot_command', '', NULL, 'view,edit', '', 1, '2018-05-16 07:06:41'),
	(13, 'management', '', NULL, 'edit', '', 1, '2018-04-12 22:47:45'),
	(14, 'statistic_dash', '', 'bb3e3925f0ad11e7bd860242ac120003', 'view', '', 1, '2018-04-12 22:47:45'),
	(15, 'ssm', '', 'bb3e3925f0ad11e7bd860242ac120003', 'view,edit,export,import', '', 1, '2018-05-16 07:04:37'),
	(16, 'qa_label', '', 'bb3e3925f0ad11e7bd860242ac120003', 'view,edit', '', 1, '2018-05-16 07:05:01'),
	(17, 'task_engine', '', 'bb3e3925f0ad11e7bd860242ac120003', 'edit', '', 1, '2018-05-16 06:59:01'),
	(18, 'wordbank', '', 'bb3e3925f0ad11e7bd860242ac120003', 'view,edit,create,delete,export,import', '', 1, '2018-04-12 22:47:45'),
	(19, 'statistic_daily', '', 'bb3e3925f0ad11e7bd860242ac120003', 'view,export', '', 1, '2018-04-12 22:47:45'),
	(20, 'statistic_analysis', '', 'bb3e3925f0ad11e7bd860242ac120003', 'view,export', '', 1, '2018-05-16 07:05:25'),
	(21, 'statistic_audit', '', 'bb3e3925f0ad11e7bd860242ac120003', 'view,export', '', 1, '2018-04-12 22:47:45'),
	(22, 'robot_profile', '', 'bb3e3925f0ad11e7bd860242ac120003', 'view,edit', '', 1, '2018-04-12 22:47:45'),
	(23, 'robot_chat_skill', '', 'bb3e3925f0ad11e7bd860242ac120003', 'view,edit', '', 1, '2018-04-12 22:47:45'),
	(24, 'robot_function', '', 'bb3e3925f0ad11e7bd860242ac120003', 'view,edit', '', 1, '2018-05-16 07:06:30'),
	(25, 'robot_command', '', 'bb3e3925f0ad11e7bd860242ac120003', 'view,edit', '', 1, '2018-05-16 07:06:41'),
	(26, 'management', '', 'bb3e3925f0ad11e7bd860242ac120003', 'edit', '', 1, '2018-04-12 22:47:45');
/*!40000 ALTER TABLE `modules` ENABLE KEYS */;

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
  `uuid` char(32) NOT NULL DEFAULT '',
  `name` char(36) NOT NULL,
  `enterprise` char(32) NOT NULL DEFAULT '',
  `description` varchar(200) NOT NULL DEFAULT '',
  `created_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uuid` (`uuid`),
  KEY `enterprise of role` (`enterprise`),
  CONSTRAINT `enterprise of role` FOREIGN KEY (`enterprise`) REFERENCES `enterprises` (`uuid`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;



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
  CONSTRAINT `column of info` FOREIGN KEY (`column_id`) REFERENCES `columns` (`id`),
  CONSTRAINT `user of info` FOREIGN KEY (`user_id`) REFERENCES `users` (`uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;



# Dump of table user_privileges
# ------------------------------------------------------------

DROP TABLE IF EXISTS `user_privileges`;

CREATE TABLE `user_privileges` (
  `id` int(20) NOT NULL AUTO_INCREMENT,
  `human` char(32) NOT NULL DEFAULT '',
  `machine` char(32) NOT NULL DEFAULT '',
  `role` char(32) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  KEY `human exist` (`human`),
  KEY `machine exist` (`machine`),
  KEY `role exist` (`role`),
  CONSTRAINT `human exist` FOREIGN KEY (`human`) REFERENCES `human` (`uuid`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `machine exist` FOREIGN KEY (`machine`) REFERENCES `machine` (`uuid`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `role exist` FOREIGN KEY (`role`) REFERENCES `roles` (`uuid`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;



# Dump of table users
# ------------------------------------------------------------

DROP TABLE IF EXISTS `users`;

CREATE TABLE `users` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `uuid` char(32) NOT NULL DEFAULT '',
  `display_name` varchar(64) NOT NULL DEFAULT '',
  `user_name` char(32) NOT NULL DEFAULT '',
  `email` char(255) NOT NULL DEFAULT '',
  `phone` char(20) NOT NULL DEFAULT '',
  `enterprise` char(32) DEFAULT NULL,
  `type` tinyint(1) unsigned NOT NULL DEFAULT '2',
  `password` char(32) NOT NULL DEFAULT '',
  `status` tinyint(1) NOT NULL DEFAULT '1',
  `created_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uuid` (`uuid`),
  KEY `enterprise of user` (`enterprise`),
  CONSTRAINT `enterprise of user` FOREIGN KEY (`enterprise`) REFERENCES `enterprises` (`uuid`) ON DELETE SET NULL ON UPDATE CASCADE,
  CONSTRAINT `user is human` FOREIGN KEY (`uuid`) REFERENCES `human` (`uuid`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

LOCK TABLES `users` WRITE;
/*!40000 ALTER TABLE `users` DISABLE KEYS */;

INSERT INTO `users` (`id`, `uuid`, `display_name`, `user_name`, `email`, `phone`, `enterprise`, `type`, `password`, `status`, `created_time`)
VALUES
	(1,'4b21158a395311e88a710242ac110003','CSBOT','csbotadmin','csbotadmin@emotibot.com','','bb3e3925f0ad11e7bd860242ac120003',1,'ac04367d3155bb651df2e4220bdb8303',1,'2018-06-12 17:23:58');

/*!40000 ALTER TABLE `users` ENABLE KEYS */;
UNLOCK TABLES;



/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;
/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
