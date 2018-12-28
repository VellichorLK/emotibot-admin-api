CREATE DATABASE `QISYS`;
use `QISYS`;

CREATE TABLE `RuleGroup` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `is_delete` tinyint(4) NOT NULL DEFAULT '0',
  `name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `enterprise` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `description` varchar(1024) COLLATE utf8mb4_unicode_ci DEFAULT '',
  `create_time` bigint(20) DEFAULT '0',
  `update_time` bigint(20) DEFAULT '0',
  `is_enable` tinyint(4) DEFAULT '1',
  `limit_speed` int(11) DEFAULT '0',
  `limit_silence` float DEFAULT '0',
  `type` tinyint(4) NOT NULL DEFAULT '0' COMMENT '0 for default, 1 for flow usage',
  `uuid` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_is_deleted` (`is_delete`),
  KEY `idx_enterprise` (`enterprise`),
  KEY `idx_is_enable` (`is_enable`),
  KEY `idx_uuid` (`uuid`),
  KEY `idx_create_time` (`create_time`),
  KEY `idx_update_time` (`update_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

LOCK TABLES `RuleGroup` WRITE;
INSERT INTO `RuleGroup` (`id`, `is_delete`, `name`, `enterprise`, `description`, `create_time`, `update_time`, `is_enable`, `limit_speed`, `limit_silence`, `type`, `uuid`)
VALUES
	(1,0,'testing','123456789','this is an integration test data',0,0,1,0,0,0, '1234'),
	(2,0,'testing2','123456789','this is another integration test data',0,0,1,0,0,1, '');

UNLOCK TABLES;

# Dump of table Tag
# ------------------------------------------------------------

DROP TABLE IF EXISTS `Tag`;

CREATE TABLE `Tag` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `is_delete` tinyint(4) NOT NULL DEFAULT '0',
  `name` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `type` tinyint(4) DEFAULT '0' COMMENT '0 is default, 1 for keyword, 2 for intent. 3 for user response',
  `pos_sentences` longtext COLLATE utf8mb4_unicode_ci,
  `neg_sentences` mediumtext COLLATE utf8mb4_unicode_ci,
  `create_time` bigint(20) NOT NULL DEFAULT '0',
  `update_time` bigint(20) NOT NULL DEFAULT '0',
  `enterprise` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  `uuid` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `tag_id_UNIQUE` (`id`),
  KEY `idx_id_delete` (`is_delete`),
  KEY `idx_type` (`type`),
  KEY `idx_enterprise` (`enterprise`),
  KEY `idx_create_time` (`create_time`),
  KEY `idx_update_time` (`update_time`),
  KEY `idx_uuid` (`uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

LOCK TABLES `Tag` WRITE;
/*!40000 ALTER TABLE `Tag` DISABLE KEYS */;

INSERT INTO `Tag` (`id`, `is_delete`, `name`, `type`, `pos_sentences`, `neg_sentences`, `create_time`, `update_time`, `enterprise`, `uuid`)
VALUES
	(1,1,'Test1',0,'[\"test1-1\", \"test1-2\", \"test1-3\"]','[]',1545901909,1545901927,'csbot','94d58cb937f34291be262095ce974f2e'),
	(2,0,'Test2',1,'[\"test2-1\"]','[\"test2-2\", \"test2-3\"]',1545901951,1545901959,'csbot','5e46b3ee737c45afb29f2a243c1aae7e');

/*!40000 ALTER TABLE `Tag` ENABLE KEYS */;
UNLOCK TABLES;