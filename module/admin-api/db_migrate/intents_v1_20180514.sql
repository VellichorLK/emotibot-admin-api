# Dump of table intent_versions
# ------------------------------------------------------------

CREATE TABLE `intent_versions` (
  `intent_version_id` int(11) NOT NULL AUTO_INCREMENT,
  `app_id` varchar(128) COLLATE utf8_unicode_ci NOT NULL,
  `ie_model_id` varchar(128) COLLATE utf8_unicode_ci DEFAULT NULL,
  `re_model_id` varchar(128) COLLATE utf8_unicode_ci DEFAULT NULL,
  `orig_file_name` varchar(256) COLLATE utf8_unicode_ci NOT NULL,
  `file_name` varchar(256) COLLATE utf8_unicode_ci NOT NULL,
  `in_used` tinyint(1) NOT NULL DEFAULT '0',
  `created_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`intent_version_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;



# Dump of table intents
# ------------------------------------------------------------

CREATE TABLE `intents` (
  `intent_id` int(11) NOT NULL AUTO_INCREMENT,
  `app_id` varchar(128) NOT NULL,
  `name` varchar(256) NOT NULL,
  `intent_version_id` int(11) NOT NULL,
  PRIMARY KEY (`intent_id`),
  KEY `intent_version_id` (`intent_version_id`),
  CONSTRAINT `intents_ibfk_1` FOREIGN KEY (`intent_version_id`) REFERENCES `intent_versions` (`intent_version_id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;



# Dump of table intent_train_sets
# ------------------------------------------------------------

CREATE TABLE `intent_train_sets` (
  `intent_train_set_id` int(11) NOT NULL AUTO_INCREMENT,
  `sentence` varchar(256) NOT NULL,
  `intent_id` int(11) NOT NULL,
  `intent_version_id` int(11) NOT NULL,
  PRIMARY KEY (`intent_train_set_id`),
  KEY `intent_id` (`intent_id`),
  KEY `intent_version_id` (`intent_version_id`),
  CONSTRAINT `intent_train_sets_ibfk_1` FOREIGN KEY (`intent_id`) REFERENCES `intents` (`intent_id`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `intent_train_sets_ibfk_2` FOREIGN KEY (`intent_version_id`) REFERENCES `intent_versions` (`intent_version_id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=6080 DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;
