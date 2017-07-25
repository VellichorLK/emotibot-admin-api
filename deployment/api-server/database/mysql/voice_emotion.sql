-- phpMyAdmin SQL Dump
-- version 4.7.2
-- https://www.phpmyadmin.net/
--
-- 主機: db
-- 產生時間： 2017 年 07 月 24 日 10:00
-- 伺服器版本: 5.7.18
-- PHP 版本： 7.0.16

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
SET AUTOCOMMIT = 0;
START TRANSACTION;
SET time_zone = "+00:00";


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;

--
-- 資料庫： `voice_emotion`
--
CREATE DATABASE IF NOT EXISTS `voice_emotion` DEFAULT CHARACTER SET utf8 COLLATE utf8_general_ci;
USE `voice_emotion`;

-- --------------------------------------------------------

--
-- 資料表結構 `analysisInformation`
--

CREATE TABLE IF NOT EXISTS `analysisInformation` (
  `segment_id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT,
  `id` bigint(20) UNSIGNED NOT NULL,
  `segment_start_time` float DEFAULT NULL,
  `segment_end_time` float DEFAULT NULL,
  `channel` int(10) UNSIGNED DEFAULT NULL,
  `status` int(11) DEFAULT NULL,
  `extra_info` blob,
  PRIMARY KEY (`segment_id`),
  UNIQUE KEY `unique_compose` (`id`,`segment_start_time`,`segment_end_time`,`channel`),
  KEY `id` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=632 DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `channelScore`
--

CREATE TABLE IF NOT EXISTS `channelScore` (
  `chanScore_id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT,
  `id` bigint(20) UNSIGNED NOT NULL,
  `channel` int(11) DEFAULT NULL,
  `emotion_type` int(11) DEFAULT NULL,
  `score` float DEFAULT NULL,
  PRIMARY KEY (`chanScore_id`),
  UNIQUE KEY `unique_compose` (`id`,`channel`,`emotion_type`),
  KEY `score` (`score`),
  KEY `id` (`id`),
  KEY `channel` (`channel`),
  KEY `emotion_type` (`emotion_type`)
) ENGINE=InnoDB AUTO_INCREMENT=35 DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `emotionInformation`
--

CREATE TABLE IF NOT EXISTS `emotionInformation` (
  `emoInfo_id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT,
  `segment_id` bigint(20) UNSIGNED NOT NULL,
  `emotion_type` int(11) DEFAULT NULL,
  `score` float DEFAULT NULL,
  PRIMARY KEY (`emoInfo_id`),
  KEY `segment_id` (`segment_id`)
) ENGINE=InnoDB AUTO_INCREMENT=1257 DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `emotionMap`
--

CREATE TABLE IF NOT EXISTS `emotionMap` (
  `emotion_type` int(11) NOT NULL,
  `emotion` varchar(32) NOT NULL,
  PRIMARY KEY (`emotion_type`),
  UNIQUE KEY `id_UNIQUE` (`emotion_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `fileInformation`
--

CREATE TABLE IF NOT EXISTS `fileInformation` (
  `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT,
  `file_id` varchar(48) NOT NULL,
  `path` varchar(256) NOT NULL,
  `file_name` varchar(256) DEFAULT NULL,
  `file_type` varchar(8) DEFAULT NULL,
  `size` int(10) UNSIGNED DEFAULT NULL,
  `duration` int(10) UNSIGNED DEFAULT NULL,
  `created_time` bigint(20) UNSIGNED DEFAULT NULL,
  `checksum` varchar(32) DEFAULT '',
  `tag1` varchar(128) DEFAULT '',
  `tag2` varchar(128) DEFAULT '',
  `priority` smallint(5) UNSIGNED DEFAULT '0',
  `appid` varchar(32) NOT NULL,
  `analysis_start_time` bigint(20) UNSIGNED DEFAULT '0',
  `analysis_end_time` bigint(20) UNSIGNED DEFAULT '0',
  `analysis_result` int(11) DEFAULT '-1',
  `upload_time` bigint(20) UNSIGNED NOT NULL,
  `real_duration` int(11) DEFAULT '-1',
  PRIMARY KEY (`id`),
  UNIQUE KEY `file_id_UNIQUE` (`file_id`),
  UNIQUE KEY `id_UNIQUE` (`id`),
  KEY `create_time` (`created_time`),
  KEY `appid` (`appid`),
  KEY `tag` (`tag1`),
  KEY `analysis_result` (`analysis_result`),
  KEY `tag2` (`tag2`)
) ENGINE=InnoDB AUTO_INCREMENT=14 DEFAULT CHARSET=utf8;

--
-- 已匯出資料表的限制(Constraint)
--

--
-- 資料表的 Constraints `analysisInformation`
--
ALTER TABLE `analysisInformation`
  ADD CONSTRAINT `id` FOREIGN KEY (`id`) REFERENCES `fileInformation` (`id`) ON DELETE NO ACTION ON UPDATE NO ACTION;

--
-- 資料表的 Constraints `emotionInformation`
--
ALTER TABLE `emotionInformation`
  ADD CONSTRAINT `segment_id` FOREIGN KEY (`segment_id`) REFERENCES `analysisInformation` (`segment_id`) ON DELETE NO ACTION ON UPDATE NO ACTION;
COMMIT;

/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
