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

DROP TABLE IF EXISTS `call`;

CREATE TABLE `call` (
  `call_id` bigint(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'call id',
  `status` tinyint(4) NOT NULL DEFAULT '0' COMMENT 'is the call status',
  `call_uuid` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'uuid for the call',
  `file_name` varchar(200) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT 'file name',
  `file_path` varchar(1024) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'file path',
  `description` varchar(1024) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'description the file',
  `duration` int(11) NOT NULL DEFAULT '0' COMMENT 'duration of the call',
  `upload_time` bigint(20) NOT NULL DEFAULT '0' COMMENT 'when the file is uploaded',
  `call_time` bigint(20) NOT NULL DEFAULT '0' COMMENT 'the happened time to this call',
  `staff_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'id of the staff who serve the call',
  `staff_name` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'name of the staff who serve the call',
  `extension` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'staff extension',
  `department` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'staff deparmtment',
  `customer_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'id of the customer in the call',
  `customer_name` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'name of the customer in the call',
  `customer_phone` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT 'phone number of the customer in the call',
  `enterprise` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'enterprise who owns the file',
  `uploader` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'who upload the file',
  `left_silence_time` float DEFAULT NULL COMMENT 'silence time of the left channel',
  `right_silence_time` float DEFAULT NULL COMMENT 'silence time of the right channel',
  `left_speed` float DEFAULT NULL COMMENT 'speak speed of the left channel',
  `right_speed` float DEFAULT NULL COMMENT 'speak speed of the right channel',
  `type` tinyint(4) NOT NULL DEFAULT '0' COMMENT 'the type of this task is created, 0 for default which is from upload audio file, 1 for task from flow',
  `left_channel` tinyint(4) NOT NULL DEFAULT '0' COMMENT '0 for default, 1 for host, 2 for guest',
  `right_channel` tinyint(4) NOT NULL DEFAULT '0' COMMENT '0 for default, 1 for host, 2 for guest',
  `apply_group_list` varchar(512) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'the list, id of group that this call would to be apply to.',
  PRIMARY KEY (`call_id`),
  UNIQUE KEY `call_uuid_UNIQUE` (`call_uuid`),
  KEY `idx_status` (`status`),
  KEY `idx_call_time` (`call_time`),
  KEY `idx_upload_time` (`upload_time`),
  KEY `idx_staff_id` (`staff_id`),
  KEY `idx_staff_name` (`staff_name`),
  KEY `idx_extension` (`extension`),
  KEY `idx_department` (`department`),
  KEY `idx_customer_id` (`customer_id`),
  KEY `idx_customer_name` (`customer_name`),
  KEY `idx_customer_phone` (`customer_phone`),
  KEY `idx_enterprise` (`enterprise`),
  KEY `idx_user` (`uploader`),
  KEY `idx_duration` (`duration`),
  KEY `index_type` (`type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `Group` (
  `app_id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `is_delete` TINYINT NULL DEFAULT 0,
  `group_name` VARCHAR(64) CHARACTER SET 'utf8mb4' COLLATE 'utf8mb4_unicode_ci' NOT NULL DEFAULT '',
  `enterprise` VARCHAR(32) CHARACTER SET 'utf8mb4' COLLATE 'utf8mb4_unicode_ci' NULL DEFAULT '',
  `description` VARCHAR(1024) CHARACTER SET 'utf8mb4' COLLATE 'utf8mb4_unicode_ci' NULL DEFAULT '',
  `create_time` BIGINT NULL DEFAULT 0,
  `update_time` BIGINT NULL DEFAULT 0,
  `is_enable` TINYINT NULL DEFAULT 1,
  `limit_speed` INT NULL DEFAULT 0,
  `limit_silence` FLOAT NULL DEFAULT 0,
  PRIMARY KEY (`app_id`),
  INDEX `idx_is_deleted` (`is_delete` ASC),
  INDEX `idx_enterprise` (`enterprise` ASC),
  INDEX `idx_is_enable` (`is_enable` ASC))
ENGINE = InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

ALTER TABLE `Group`
ADD COLUMN `type` TINYINT NOT NULL DEFAULT 0 COMMENT '0 for default, 1 for flow usage';
