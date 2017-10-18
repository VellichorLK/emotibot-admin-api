-- phpMyAdmin SQL Dump
-- version 4.7.2
-- https://www.phpmyadmin.net/
--
-- 主機: db
-- 產生時間： 2017 年 07 月 24 日 09:59
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
-- 資料庫： `emotibot`
--
CREATE DATABASE IF NOT EXISTS `emotibot` DEFAULT CHARACTER SET utf8 COLLATE utf8_general_ci;
USE `emotibot`;

-- --------------------------------------------------------

--
-- 資料表結構 `accountuser`
--

CREATE TABLE IF NOT EXISTS `accountuser` (
  `AUserID` int(11) NOT NULL AUTO_INCREMENT,
  `UserID` varchar(50) NOT NULL,
  `3RD` varchar(20) CHARACTER SET utf8 COLLATE utf8_unicode_ci NOT NULL DEFAULT '',
  `Account` varchar(40) CHARACTER SET utf8 COLLATE utf8_unicode_ci NOT NULL DEFAULT '',
  `CreateTime` datetime NOT NULL,
  PRIMARY KEY (`AUserID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `answer`
--

CREATE TABLE IF NOT EXISTS `answer` (
  `a_id` int(4) NOT NULL AUTO_INCREMENT,
  `parent_q_id` int(4) NOT NULL,
  `content` text COLLATE utf8_unicode_ci NOT NULL,
  `user` varchar(255) COLLATE utf8_unicode_ci NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `score` smallint(5) DEFAULT NULL,
  `feature_words` text COLLATE utf8_unicode_ci,
  `modal` varchar(20) COLLATE utf8_unicode_ci DEFAULT NULL,
  `intention` varchar(200) COLLATE utf8_unicode_ci DEFAULT NULL,
  `mood` varchar(20) COLLATE utf8_unicode_ci DEFAULT NULL,
  `is_girl_friend` tinyint(1) DEFAULT NULL,
  `is_sutiable` tinyint(1) DEFAULT NULL,
  `rule` varchar(20) COLLATE utf8_unicode_ci DEFAULT NULL,
  `seg_word` text COLLATE utf8_unicode_ci,
  `feature_words_1` text COLLATE utf8_unicode_ci,
  `feature_words_2` text COLLATE utf8_unicode_ci,
  `is_modified` tinyint(1) DEFAULT '0',
  `modal_updated` varchar(50) COLLATE utf8_unicode_ci DEFAULT NULL,
  `mood1` varchar(100) COLLATE utf8_unicode_ci DEFAULT NULL,
  `emotion` text COLLATE utf8_unicode_ci,
  `topic` text COLLATE utf8_unicode_ci,
  `act` text COLLATE utf8_unicode_ci,
  `intent` text COLLATE utf8_unicode_ci,
  `qa_score` text COLLATE utf8_unicode_ci,
  `CUOutput` text COLLATE utf8_unicode_ci,
  `created_user` int(11) DEFAULT NULL,
  `updated_user` int(11) DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  `status` int(11) NOT NULL DEFAULT '1',
  PRIMARY KEY (`a_id`),
  KEY `content` (`content`(255)),
  KEY `IDX_a_id` (`a_id`),
  KEY `answer_parent_q_id` (`parent_q_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

-- --------------------------------------------------------

--
-- 資料表結構 `api_functions`
--

CREATE TABLE IF NOT EXISTS `api_functions` (
  `FunctionId` int(11) NOT NULL AUTO_INCREMENT,
  `FunctionName` varchar(50) NOT NULL,
  `FunctionType` int(11) NOT NULL DEFAULT '0',
  `CreatedTime` datetime DEFAULT NULL,
  `Icon` varchar(100) DEFAULT NULL,
  `CodePath` varchar(200) DEFAULT NULL,
  `Rank` int(11) DEFAULT NULL,
  `Status` int(11) DEFAULT '1',
  PRIMARY KEY (`FunctionId`)
) ENGINE=InnoDB AUTO_INCREMENT=51 DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `api_generator`
--

CREATE TABLE IF NOT EXISTS `api_generator` (
  `AppId` varchar(100) NOT NULL,
  `G_UserId` varchar(100) NOT NULL,
  `UserId` varchar(100) DEFAULT NULL,
  PRIMARY KEY (`AppId`,`G_UserId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `api_key_ip`
--

CREATE TABLE IF NOT EXISTS `api_key_ip` (
  `ApiKeyIpId` int(11) NOT NULL AUTO_INCREMENT,
  `AppId` varchar(50) NOT NULL,
  `NickName` varchar(100) DEFAULT NULL,
  `IP` varchar(50) NOT NULL,
  `Status` int(11) NOT NULL DEFAULT '0',
  `CreatedTime` datetime DEFAULT NULL,
  `Tag` varchar(50) DEFAULT NULL,
  `DB_User` varchar(50) DEFAULT NULL,
  `DB_Passwd` varchar(50) DEFAULT NULL,
  `DB_IP` varchar(50) DEFAULT NULL,
  PRIMARY KEY (`ApiKeyIpId`),
  KEY `chatstatus_createdtime_status` (`AppId`,`Tag`,`NickName`,`IP`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `api_preduct`
--

CREATE TABLE IF NOT EXISTS `api_preduct` (
  `PreductId` int(11) NOT NULL AUTO_INCREMENT,
  `PreductName` varchar(200) DEFAULT NULL,
  `PreductRemark` varchar(1000) DEFAULT NULL,
  `CreatedUser` int(11) NOT NULL,
  `CreatedTime` datetime NOT NULL,
  `PreductVersion` int(11) NOT NULL,
  `Status` int(11) NOT NULL,
  PRIMARY KEY (`PreductId`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `api_talkingdata`
--

CREATE TABLE IF NOT EXISTS `api_talkingdata` (
  `TD_AppId` varchar(100) NOT NULL,
  `TD_UserId` varchar(100) NOT NULL,
  `UserId` varchar(100) DEFAULT NULL,
  PRIMARY KEY (`TD_AppId`,`TD_UserId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `api_user`
--

CREATE TABLE IF NOT EXISTS `api_user` (
  `UserId` varchar(50) NOT NULL,
  `Phone` varchar(100) DEFAULT NULL,
  `Email` varchar(100) DEFAULT NULL,
  `CreatedTime` datetime NOT NULL,
  `Password` varchar(255) NOT NULL,
  `NickName` varchar(255) NOT NULL,
  `Gender` int(11) DEFAULT NULL,
  `Type` int(11) NOT NULL,
  `Status` int(11) NOT NULL,
  `UpdatedTime` datetime NOT NULL,
  `Owner` varchar(100) DEFAULT NULL,
  `Remark` varchar(1000) DEFAULT NULL,
  `AiNickName` varchar(100) DEFAULT NULL,
  `Msg` varchar(200) DEFAULT '你好，我是你的机器人XX，我可以陪你聊天，为你答疑解惑哦！',
  PRIMARY KEY (`UserId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `api_usercount`
--

CREATE TABLE IF NOT EXISTS `api_usercount` (
  `AppId` varchar(50) DEFAULT NULL,
  `UserId` varchar(50) DEFAULT NULL,
  `ChatCount` int(11) DEFAULT NULL,
  `CreatedTime` datetime DEFAULT NULL,
  `NickName` varchar(100) DEFAULT NULL,
  KEY `user_count` (`AppId`,`CreatedTime`,`UserId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `api_userkey`
--

CREATE TABLE IF NOT EXISTS `api_userkey` (
  `UserId` varchar(50) NOT NULL,
  `Count` int(11) NOT NULL,
  `Version` int(11) NOT NULL,
  `CreatedTime` datetime NOT NULL,
  `PreductName` varchar(255) NOT NULL,
  `ApiKey` varchar(50) NOT NULL,
  `Status` int(11) NOT NULL,
  `MaxCount` int(11) DEFAULT '50000',
  `AutoUserID` int(11) DEFAULT '0',
  `NickName` varchar(255) DEFAULT NULL,
  `CommonFunctionIds` varchar(200) DEFAULT NULL,
  `AreaIds` varchar(200) DEFAULT NULL,
  `Type` int(11) DEFAULT '1',
  `MsgType` int(11) DEFAULT '1',
  `Msg` varchar(200) DEFAULT '你好，我是你的机器人XX，我可以陪你聊天，为你答疑解惑哦！',
  `MsgJson` text,
  PRIMARY KEY (`UserId`,`PreductName`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `api_usertypecount`
--

CREATE TABLE IF NOT EXISTS `api_usertypecount` (
  `AppId` varchar(50) DEFAULT NULL,
  `Type` varchar(50) DEFAULT NULL,
  `TypeCount` double(8,2) DEFAULT NULL,
  `CreatedTime` datetime DEFAULT NULL,
  KEY `user_count` (`AppId`,`CreatedTime`,`Type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `appidlog`
--

CREATE TABLE IF NOT EXISTS `appidlog` (
  `LogId` int(11) NOT NULL AUTO_INCREMENT,
  `UserId` varchar(200) DEFAULT NULL,
  `FunctionName` varchar(200) DEFAULT NULL,
  `Module` varchar(200) DEFAULT NULL,
  `AppId` varchar(200) DEFAULT NULL,
  `CreatedTime` datetime NOT NULL,
  PRIMARY KEY (`LogId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `appid_list`
--

CREATE TABLE IF NOT EXISTS `appid_list` (
  `app_id` char(32) NOT NULL,
  `created_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `api_cnt` int(11) NOT NULL COMMENT 'daily count limitation',
  `expiration_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'client decide the usage',
  `analysis_time` int(11) NOT NULL COMMENT 'total voice limitation, in second',
  `activation` tinyint(1) NOT NULL,
  PRIMARY KEY (`app_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `appversion`
--

CREATE TABLE IF NOT EXISTS `appversion` (
  `AppVersionId` int(11) NOT NULL AUTO_INCREMENT,
  `Title` varchar(200) NOT NULL,
  `Description` longtext,
  `Status` int(11) NOT NULL DEFAULT '1',
  `CreatedUser` int(11) DEFAULT NULL,
  `CreatedTime` datetime DEFAULT NULL,
  `EditUser` int(11) DEFAULT NULL,
  `EditTime` datetime DEFAULT NULL,
  `VersionCode` int(11) NOT NULL,
  `isForced` int(11) NOT NULL,
  `appSource` varchar(50) DEFAULT NULL,
  PRIMARY KEY (`AppVersionId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `appvoice`
--

CREATE TABLE IF NOT EXISTS `appvoice` (
  `VoiceID` int(11) NOT NULL AUTO_INCREMENT,
  `UserID` varchar(200) NOT NULL,
  `msg` varchar(200) DEFAULT NULL,
  `gender` int(11) DEFAULT NULL,
  `CreateTime` datetime NOT NULL,
  `Status` int(11) NOT NULL,
  `tag` varchar(20) DEFAULT NULL,
  `Reply` varchar(1000) DEFAULT NULL,
  PRIMARY KEY (`VoiceID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `area`
--

CREATE TABLE IF NOT EXISTS `area` (
  `AreaId` int(11) NOT NULL AUTO_INCREMENT,
  `AreaName` varchar(50) NOT NULL,
  `Remark` varchar(1000) DEFAULT NULL,
  `CreatedTime` datetime DEFAULT NULL,
  `Status` int(11) NOT NULL DEFAULT '1',
  `Icon` varchar(100) DEFAULT NULL,
  PRIMARY KEY (`AreaId`)
) ENGINE=InnoDB AUTO_INCREMENT=10 DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `auth`
--

CREATE TABLE IF NOT EXISTS `auth` (
  `AuthId` int(11) NOT NULL AUTO_INCREMENT,
  `Phone` varchar(100) NOT NULL,
  `AuthNum` varchar(50) DEFAULT NULL,
  `CreatedTime` datetime DEFAULT NULL,
  `Status` int(11) NOT NULL DEFAULT '1',
  PRIMARY KEY (`AuthId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `baidu_song`
--

CREATE TABLE IF NOT EXISTS `baidu_song` (
  `SongId` int(11) NOT NULL AUTO_INCREMENT,
  `Search_Song` varchar(200) NOT NULL,
  `Singer` varchar(200) NOT NULL,
  `Status` int(11) NOT NULL DEFAULT '1',
  `Baidu_Song` varchar(200) NOT NULL,
  `Full_Match` tinyint(1) DEFAULT NULL,
  `CreatedTime` datetime DEFAULT NULL,
  `EditTime` datetime DEFAULT NULL,
  PRIMARY KEY (`SongId`),
  UNIQUE KEY `BaiduSong` (`Singer`,`Baidu_Song`)
) ENGINE=InnoDB AUTO_INCREMENT=52504 DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `categories`
--

CREATE TABLE IF NOT EXISTS `categories` (
  `CategoryId` int(11) NOT NULL AUTO_INCREMENT,
  `Name` varchar(200) NOT NULL,
  `Value` longtext,
  `Status` int(11) NOT NULL DEFAULT '1',
  `CreatedUser` int(11) DEFAULT NULL,
  `CreatedTime` datetime DEFAULT NULL,
  `EditUser` int(11) DEFAULT NULL,
  `EditTime` datetime DEFAULT NULL,
  `TopicId` varchar(100) NOT NULL,
  `Type` int(11) NOT NULL,
  PRIMARY KEY (`CategoryId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `chatstatus`
--

CREATE TABLE IF NOT EXISTS `chatstatus` (
  `UUId` int(11) NOT NULL AUTO_INCREMENT,
  `UserId` varchar(50) DEFAULT NULL,
  `Text` varchar(1000) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `Answer` varchar(1000) DEFAULT NULL,
  `Status` int(11) NOT NULL,
  `CreatedTime` datetime DEFAULT NULL,
  `WeChatId` varchar(100) DEFAULT NULL,
  `Source` varchar(200) DEFAULT NULL,
  `Score` float DEFAULT NULL,
  `Module` varchar(200) DEFAULT NULL,
  `OldText` varchar(1000) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `IFormat` varchar(20) DEFAULT NULL,
  `OFormat` varchar(20) DEFAULT NULL,
  `FeedBack` varchar(100) DEFAULT NULL,
  `tag` varchar(20) DEFAULT NULL,
  `Owner` varchar(100) DEFAULT NULL,
  `SubModule` varchar(100) DEFAULT NULL,
  PRIMARY KEY (`UUId`),
  KEY `chatstatus_createdtime_status` (`CreatedTime`,`Status`),
  KEY `chatstatus_UserId_IDX` (`UserId`),
  KEY `chatstatus_WeChatId_IDX` (`WeChatId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `commonfunction`
--

CREATE TABLE IF NOT EXISTS `commonfunction` (
  `CommonFunctionId` int(11) NOT NULL AUTO_INCREMENT,
  `CommonFunctionName` varchar(50) NOT NULL,
  `Remark` varchar(1000) DEFAULT NULL,
  `CreatedTime` datetime DEFAULT NULL,
  `Status` int(11) NOT NULL DEFAULT '1',
  `Icon` varchar(100) DEFAULT NULL,
  `Rank` int(11) DEFAULT '0',
  PRIMARY KEY (`CommonFunctionId`),
  KEY `commonfunction_CommonFunctionName_IDX` (`CommonFunctionName`)
) ENGINE=InnoDB AUTO_INCREMENT=9 DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `configfile`
--

CREATE TABLE IF NOT EXISTS `configfile` (
  `ModuleID` int(11) NOT NULL AUTO_INCREMENT,
  `ParentID` varchar(255) DEFAULT NULL,
  `ExclusiveID` varchar(255) DEFAULT NULL,
  `ModuleName` varchar(255) DEFAULT NULL,
  `FunctionName` varchar(255) NOT NULL,
  `Possibility` int(11) NOT NULL,
  `CreatedUser` int(11) DEFAULT NULL,
  `CreatedTime` datetime DEFAULT NULL,
  `EditUser` int(11) DEFAULT NULL,
  `EditTime` datetime DEFAULT NULL,
  `OccurTime` datetime DEFAULT NULL,
  `ParameterNum` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`ModuleID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `copy_temple`
--

CREATE TABLE IF NOT EXISTS `copy_temple` (
  `AppId` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `Module` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `CreatedTime` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `Status` varchar(10) COLLATE utf8mb4_unicode_ci NOT NULL,
  `Message` text COLLATE utf8mb4_unicode_ci,
  `TmpAppId` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`AppId`,`Module`),
  KEY `IDX_app_module` (`AppId`,`Module`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- 資料表結構 `corpussupplement`
--

CREATE TABLE IF NOT EXISTS `corpussupplement` (
  `CCId` int(11) NOT NULL AUTO_INCREMENT,
  `ChatId` int(11) NOT NULL,
  `Q` varchar(200) DEFAULT NULL,
  `Status` int(11) NOT NULL,
  `CreateTime` datetime NOT NULL,
  PRIMARY KEY (`CCId`)
) ENGINE=InnoDB AUTO_INCREMENT=10 DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `down_log`
--

CREATE TABLE IF NOT EXISTS `down_log` (
  `LogId` int(11) NOT NULL AUTO_INCREMENT,
  `UserName` varchar(200) NOT NULL,
  `remake_log` varchar(1000) NOT NULL,
  `created_user` int(11) NOT NULL,
  `created_at` datetime NOT NULL,
  `source_name` varchar(100) NOT NULL,
  PRIMARY KEY (`LogId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `emoticon_mapping`
--

CREATE TABLE IF NOT EXISTS `emoticon_mapping` (
  `id` int(8) NOT NULL AUTO_INCREMENT,
  `emoticon` varchar(120) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `meaning` varchar(120) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=42667 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- 資料表結構 `enterprise`
--

CREATE TABLE IF NOT EXISTS `enterprise` (
  `EnterpriseId` varchar(50) NOT NULL,
  `Account` varchar(50) NOT NULL,
  `CreatedTime` datetime NOT NULL,
  `Password` varchar(255) NOT NULL,
  `NickName` varchar(255) NOT NULL,
  `Location` varchar(200) DEFAULT NULL,
  `PeopleNumber` int(11) DEFAULT '0',
  `Industry` varchar(100) DEFAULT '',
  `LinkName` varchar(100) NOT NULL,
  `LinkPhone` varchar(100) NOT NULL,
  `LinkEmail` varchar(100) NOT NULL,
  `Status` int(11) DEFAULT '0' COMMENT '-1:审核失败; 0:未审核; 1:已审核',
  `UpdatedTime` datetime NOT NULL,
  `Remark` varchar(1000) DEFAULT NULL,
  `UserId` varchar(50) NOT NULL,
  PRIMARY KEY (`EnterpriseId`),
  KEY `enterprise_Account_IDX` (`Account`,`CreatedTime`,`Password`,`LinkEmail`,`LinkPhone`,`LinkName`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `enterprise_list`
--

CREATE TABLE IF NOT EXISTS `enterprise_list` (
  `enterprise_id` char(32) NOT NULL,
  `enterprise_name` varchar(64) NOT NULL,
  `created_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `industry` varchar(32) DEFAULT NULL,
  `phone_number` varchar(16) DEFAULT NULL COMMENT '+86 23471234',
  `address` varchar(256) DEFAULT NULL,
  `people_numbers` int(5) NOT NULL,
  `app_id` char(32) NOT NULL,
  PRIMARY KEY (`enterprise_id`),
  UNIQUE KEY `app_id_idx` (`app_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `enterprise_user`
--

CREATE TABLE IF NOT EXISTS `enterprise_user` (
  `EnterpriseId` varchar(50) NOT NULL,
  `UserId` varchar(50) NOT NULL,
  `CreatedTime` datetime NOT NULL,
  `Status` int(11) DEFAULT '0' COMMENT '-1:解除绑定; 0:绑定要请; 1:管理员(admin); 2:普通用户',
  `UpdatedTime` datetime NOT NULL,
  `EnterpriseRemark` varchar(1000) DEFAULT NULL,
  `UserRemark` varchar(1000) DEFAULT NULL,
  PRIMARY KEY (`EnterpriseId`,`UserId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `faceimage`
--

CREATE TABLE IF NOT EXISTS `faceimage` (
  `FImageID` int(11) NOT NULL AUTO_INCREMENT,
  `UserID` varchar(50) NOT NULL,
  `smile` int(11) NOT NULL,
  `male` int(11) NOT NULL,
  `human` int(11) NOT NULL,
  `flag` int(11) NOT NULL,
  `Status` int(11) NOT NULL,
  `Type` int(11) NOT NULL,
  `CreateTime` datetime NOT NULL,
  `confidence` varchar(100) DEFAULT NULL,
  `sourceImg` int(11) DEFAULT NULL,
  PRIMARY KEY (`FImageID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `feedback`
--

CREATE TABLE IF NOT EXISTS `feedback` (
  `FeedBackID` int(11) NOT NULL AUTO_INCREMENT,
  `NickName` varchar(200) NOT NULL,
  `MSType` varchar(200) DEFAULT NULL,
  `MSG` varchar(300) DEFAULT NULL,
  `Phone` varchar(50) DEFAULT NULL,
  `CreateTime` datetime NOT NULL,
  `UserID` varchar(50) NOT NULL,
  PRIMARY KEY (`FeedBackID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `finalranker`
--

CREATE TABLE IF NOT EXISTS `finalranker` (
  `FRId` int(11) NOT NULL AUTO_INCREMENT,
  `Match_Question` varchar(1000) DEFAULT NULL,
  `Score` float DEFAULT NULL,
  `CreatedTime` datetime DEFAULT NULL,
  `UUId` int(11) NOT NULL,
  `Answer` varchar(1000) DEFAULT NULL,
  `FunctionName` varchar(100) DEFAULT NULL,
  PRIMARY KEY (`FRId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `finalrankermyscore`
--

CREATE TABLE IF NOT EXISTS `finalrankermyscore` (
  `MyScoreId` int(11) NOT NULL AUTO_INCREMENT,
  `FRId` int(11) NOT NULL,
  `UserID` int(11) NOT NULL,
  `UUId` int(11) NOT NULL,
  `UpdatedTime` datetime DEFAULT NULL,
  `Score` float NOT NULL,
  PRIMARY KEY (`MyScoreId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `finalrankermysort`
--

CREATE TABLE IF NOT EXISTS `finalrankermysort` (
  `MySortId` int(11) NOT NULL AUTO_INCREMENT,
  `FRIds` varchar(2000) NOT NULL,
  `UserId` int(11) NOT NULL,
  `UUId` int(11) NOT NULL,
  `CreatedTime` datetime DEFAULT NULL,
  `num` int(11) NOT NULL,
  PRIMARY KEY (`MySortId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `freemeguidance`
--

CREATE TABLE IF NOT EXISTS `freemeguidance` (
  `tag` varchar(50) NOT NULL,
  `reload` int(11) NOT NULL,
  `freememsg` varchar(1000) DEFAULT NULL,
  PRIMARY KEY (`tag`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `functiononoff`
--

CREATE TABLE IF NOT EXISTS `functiononoff` (
  `FunctionName_En` varchar(50) NOT NULL,
  `FunctionName_Zh` varchar(50) NOT NULL,
  `Remark` varchar(1000) DEFAULT NULL,
  `CreatedTime` datetime DEFAULT NULL,
  `CommonFunctionId` int(11) DEFAULT NULL,
  `Rank` int(11) DEFAULT '0',
  `Intent` varchar(500) DEFAULT NULL,
  `Url` varchar(500) DEFAULT NULL,
  `FunctionOnOff` int(11) DEFAULT '1',
  PRIMARY KEY (`FunctionName_En`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `functions`
--

CREATE TABLE IF NOT EXISTS `functions` (
  `FunctionId` int(11) NOT NULL AUTO_INCREMENT,
  `FunctionName` varchar(50) NOT NULL,
  `FunctionType` int(11) NOT NULL DEFAULT '0',
  `CreatedTime` datetime DEFAULT NULL,
  `Icon` varchar(100) DEFAULT NULL,
  `CodePath` varchar(200) DEFAULT NULL,
  `Rank` int(11) DEFAULT NULL,
  `Status` int(11) DEFAULT '1',
  PRIMARY KEY (`FunctionId`)
) ENGINE=InnoDB AUTO_INCREMENT=49 DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `function_share`
--

CREATE TABLE IF NOT EXISTS `function_share` (
  `AppId` varchar(50) NOT NULL,
  `FunctionName_En` varchar(50) NOT NULL,
  `FunctionName_Zh` varchar(50) DEFAULT NULL,
  `Remark` varchar(1000) DEFAULT NULL,
  `Intent` varchar(200) DEFAULT NULL,
  `Url` varchar(500) DEFAULT NULL,
  `FunctionId` varchar(100) NOT NULL,
  `Tag` varchar(200) DEFAULT NULL,
  `Status` int(11) DEFAULT '1' COMMENT '-1:删除; 0:停止; 1:启动',
  `CreatedTime` datetime DEFAULT NULL,
  `Sample1` varchar(200) NOT NULL,
  `Sample2` varchar(200) NOT NULL,
  `RejectRemark` varchar(1000) DEFAULT NULL,
  PRIMARY KEY (`AppId`,`FunctionName_En`,`FunctionId`),
  KEY `user_count` (`AppId`,`FunctionName_En`,`FunctionId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `function_share_categories`
--

CREATE TABLE IF NOT EXISTS `function_share_categories` (
  `AppId` varchar(50) NOT NULL,
  `FunctionName_En` varchar(50) NOT NULL,
  `FunctionId` varchar(100) NOT NULL,
  `CategoryId` int(11) NOT NULL,
  KEY `function_share_categories` (`AppId`,`FunctionName_En`,`FunctionId`,`CategoryId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `function_share_intent`
--

CREATE TABLE IF NOT EXISTS `function_share_intent` (
  `AppId` varchar(50) NOT NULL,
  `FunctionName_En` varchar(50) NOT NULL,
  `FunctionId` varchar(100) NOT NULL,
  `IntentId` int(11) NOT NULL,
  KEY `function_share_intent` (`AppId`,`FunctionName_En`,`FunctionId`,`IntentId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `greeting`
--

CREATE TABLE IF NOT EXISTS `greeting` (
  `GreetingId` int(11) NOT NULL AUTO_INCREMENT,
  `Subject` varchar(200) NOT NULL,
  `Type` int(11) NOT NULL,
  `StartTime` datetime DEFAULT NULL,
  `EndTime` datetime DEFAULT NULL,
  `Message` varchar(1000) NOT NULL,
  `Expression` varchar(200) DEFAULT NULL,
  `CreatedUser` int(11) DEFAULT NULL,
  `CreatedTime` datetime DEFAULT NULL,
  `EditUser` int(11) DEFAULT NULL,
  `EditTime` datetime DEFAULT NULL,
  `Status` int(11) NOT NULL,
  PRIMARY KEY (`GreetingId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `ibb_qa`
--

CREATE TABLE IF NOT EXISTS `ibb_qa` (
  `IBBQAId` int(11) NOT NULL AUTO_INCREMENT,
  `Module` varchar(50) NOT NULL,
  `CreatedTime` datetime DEFAULT NULL,
  PRIMARY KEY (`IBBQAId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `imagetagcategory`
--

CREATE TABLE IF NOT EXISTS `imagetagcategory` (
  `ImageTagId` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(20) CHARACTER SET utf8 NOT NULL,
  `status` int(11) NOT NULL,
  PRIMARY KEY (`ImageTagId`)
) ENGINE=InnoDB AUTO_INCREMENT=10 DEFAULT CHARSET=latin1;

-- --------------------------------------------------------

--
-- 資料表結構 `imagetagname`
--

CREATE TABLE IF NOT EXISTS `imagetagname` (
  `ImageTagNameId` int(11) NOT NULL AUTO_INCREMENT,
  `ImageTagId` int(11) NOT NULL,
  `name` varchar(20) CHARACTER SET utf8 NOT NULL,
  `synonym` varchar(1000) CHARACTER SET utf8 NOT NULL,
  `status` int(11) NOT NULL,
  PRIMARY KEY (`ImageTagNameId`)
) ENGINE=InnoDB AUTO_INCREMENT=86 DEFAULT CHARSET=latin1;

-- --------------------------------------------------------

--
-- 資料表結構 `image_authenticate`
--

CREATE TABLE IF NOT EXISTS `image_authenticate` (
  `ImgId` int(11) NOT NULL AUTO_INCREMENT,
  `ImgStatus` varchar(100) DEFAULT NULL,
  `ImgFilter` varchar(100) DEFAULT NULL,
  `ImgReview` varchar(100) DEFAULT NULL,
  `ImgMessage` varchar(100) DEFAULT NULL,
  `ImgProcessedTime` float(10,10) DEFAULT NULL,
  `ImgImage` varchar(200) DEFAULT NULL,
  `Status` int(11) NOT NULL DEFAULT '1',
  `UserId` varchar(50) DEFAULT NULL,
  `CreatedTime` datetime DEFAULT NULL,
  `WechatId` varchar(50) DEFAULT NULL,
  `Owner` varchar(200) DEFAULT NULL,
  `response` varchar(100) DEFAULT NULL,
  `ImgPossibleActions` varchar(200) DEFAULT NULL,
  `UploadId` varchar(200) DEFAULT NULL,
  PRIMARY KEY (`ImgId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `image_category`
--

CREATE TABLE IF NOT EXISTS `image_category` (
  `pic_id` int(11) NOT NULL AUTO_INCREMENT,
  `tag` varchar(255) CHARACTER SET utf8 NOT NULL,
  `modified_time` datetime NOT NULL,
  `user` varchar(65) CHARACTER SET utf8 NOT NULL,
  `status` int(11) NOT NULL,
  `file_path` varchar(255) NOT NULL,
  PRIMARY KEY (`pic_id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- --------------------------------------------------------

--
-- 資料表結構 `intentanswer`
--

CREATE TABLE IF NOT EXISTS `intentanswer` (
  `IntentAnswerId` int(11) NOT NULL AUTO_INCREMENT,
  `IntentAnswerName` varchar(50) NOT NULL,
  `CreatedUser` varchar(50) DEFAULT NULL,
  `CreatedTime` datetime DEFAULT NULL,
  `UpdatedUser` varchar(50) DEFAULT NULL,
  `UpdatedTime` datetime DEFAULT NULL,
  `Status` int(11) NOT NULL DEFAULT '1',
  PRIMARY KEY (`IntentAnswerId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `intentquestion`
--

CREATE TABLE IF NOT EXISTS `intentquestion` (
  `IntentQuestionId` int(11) NOT NULL AUTO_INCREMENT,
  `IntentQuestionName` varchar(1000) DEFAULT NULL,
  `IntentAnswerId` varchar(11) NOT NULL,
  `CreatedUser` varchar(50) DEFAULT NULL,
  `CreatedTime` datetime DEFAULT NULL,
  `UpdatedUser` varchar(50) DEFAULT NULL,
  `UpdatedTime` datetime DEFAULT NULL,
  `Status` int(11) NOT NULL DEFAULT '1',
  PRIMARY KEY (`IntentQuestionId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `interntable`
--

CREATE TABLE IF NOT EXISTS `interntable` (
  `InternID` int(11) NOT NULL AUTO_INCREMENT,
  `Name` varchar(50) DEFAULT NULL,
  `UserNum` varchar(20) DEFAULT NULL,
  `Phone` varchar(20) DEFAULT NULL,
  `WeChatNum` varchar(50) DEFAULT NULL,
  `CreatedUser` int(11) NOT NULL,
  `CreatedTime` datetime NOT NULL,
  `EditUser` int(11) NOT NULL,
  `EditTime` datetime NOT NULL,
  `Status` int(11) NOT NULL,
  `UserCategory` int(11) NOT NULL,
  PRIMARY KEY (`InternID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `keywordgrouping`
--

CREATE TABLE IF NOT EXISTS `keywordgrouping` (
  `KeywordGroupingId` int(11) NOT NULL AUTO_INCREMENT,
  `Name` varchar(200) NOT NULL,
  `Value` varchar(200) DEFAULT NULL,
  `Status` int(11) NOT NULL DEFAULT '1',
  `CreatedUser` int(11) DEFAULT NULL,
  `CreatedTime` datetime DEFAULT NULL,
  `EditUser` int(11) DEFAULT NULL,
  `EditTime` datetime DEFAULT NULL,
  `TopicId` varchar(100) NOT NULL,
  `Type` int(11) NOT NULL,
  PRIMARY KEY (`KeywordGroupingId`),
  UNIQUE KEY `keywordgrouping_Name_IDX` (`Name`,`Value`,`TopicId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `log`
--

CREATE TABLE IF NOT EXISTS `log` (
  `LogId` int(11) NOT NULL AUTO_INCREMENT,
  `UserId` varchar(200) NOT NULL,
  `NickName` varchar(200) DEFAULT NULL,
  `FunctionName` varchar(200) NOT NULL,
  `ActionName` varchar(200) NOT NULL,
  `ActionTime` datetime NOT NULL,
  `ActionReturn` text,
  `ActionObject` varchar(200) DEFAULT NULL,
  `AppName` varchar(200) NOT NULL,
  `UUId` int(11) NOT NULL,
  PRIMARY KEY (`LogId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `module`
--

CREATE TABLE IF NOT EXISTS `module` (
  `ModuleId` int(11) NOT NULL AUTO_INCREMENT,
  `ModuleCode` varchar(50) NOT NULL,
  `ModuleName` varchar(100) NOT NULL,
  `ParentCode` varchar(50) NOT NULL,
  `ModuleUrl` varchar(500) NOT NULL,
  `CreatedUserId` varchar(50) NOT NULL,
  `CreatedTime` datetime NOT NULL,
  `Status` int(11) DEFAULT '0' COMMENT '-1:删除; 0:停止; 1:启动',
  `EditUserId` varchar(50) NOT NULL,
  `UpdatedTime` datetime NOT NULL,
  PRIMARY KEY (`ModuleId`),
  KEY `module_ModuleCode_IDX` (`ModuleCode`,`ModuleName`,`ParentCode`,`Status`)
) ENGINE=InnoDB AUTO_INCREMENT=17 DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `module_privilege`
--

CREATE TABLE IF NOT EXISTS `module_privilege` (
  `MPId` int(11) NOT NULL AUTO_INCREMENT,
  `ModuleCode` varchar(50) NOT NULL,
  `PriCode` varchar(50) NOT NULL,
  `CreatedUserId` varchar(50) NOT NULL,
  `CreatedTime` datetime NOT NULL,
  PRIMARY KEY (`MPId`),
  KEY `module_privilege_ModuleCode_IDX` (`ModuleCode`,`PriCode`)
) ENGINE=InnoDB AUTO_INCREMENT=40 DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `news`
--

CREATE TABLE IF NOT EXISTS `news` (
  `NewsId` int(11) NOT NULL AUTO_INCREMENT,
  `Guidance` varchar(100) NOT NULL,
  `Viewpoint` varchar(100) NOT NULL,
  `CreatedUser` int(11) DEFAULT NULL,
  `CreatedTime` datetime DEFAULT NULL,
  `EditUser` int(11) DEFAULT NULL,
  `EditTime` datetime DEFAULT NULL,
  `Status` int(11) NOT NULL,
  `ImageName` varchar(100) NOT NULL,
  PRIMARY KEY (`NewsId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `notusertable`
--

CREATE TABLE IF NOT EXISTS `notusertable` (
  `UserID` int(11) NOT NULL,
  `NickName` varchar(200) DEFAULT NULL,
  PRIMARY KEY (`UserID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `perform`
--

CREATE TABLE IF NOT EXISTS `perform` (
  `Perform_Id` int(11) NOT NULL AUTO_INCREMENT,
  `Perform_Title` varchar(500) NOT NULL,
  `Picture_Url` varchar(255) NOT NULL,
  `Perform_Type` int(11) DEFAULT NULL,
  `Perform_Ers` varchar(200) NOT NULL,
  `Perform_Url` varchar(255) DEFAULT NULL,
  `Perform_Date` datetime NOT NULL,
  `Perform_Location` varchar(500) NOT NULL,
  `Status` int(11) NOT NULL DEFAULT '2',
  `CreatedTime` datetime DEFAULT NULL,
  `EditTime` datetime DEFAULT NULL,
  `Perform_Venue` varchar(500) NOT NULL,
  PRIMARY KEY (`Perform_Id`),
  UNIQUE KEY `Perform` (`Perform_Ers`,`Perform_Date`,`Perform_Location`)
) ENGINE=InnoDB AUTO_INCREMENT=53643 DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `privileges`
--

CREATE TABLE IF NOT EXISTS `privileges` (
  `PrivilegeId` int(11) NOT NULL AUTO_INCREMENT,
  `UserId` int(11) NOT NULL,
  `FunctionId` int(11) NOT NULL,
  `Status` int(11) NOT NULL DEFAULT '0',
  `CreatedUser` int(11) DEFAULT NULL,
  `CreatedTime` datetime DEFAULT NULL,
  `EditUser` int(11) DEFAULT NULL,
  `EditTime` datetime DEFAULT NULL,
  PRIMARY KEY (`PrivilegeId`)
) ENGINE=InnoDB AUTO_INCREMENT=6598 DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `privilege_list`
--

CREATE TABLE IF NOT EXISTS `privilege_list` (
  `privilege_id` int(11) NOT NULL,
  `privilege_name` varchar(32) NOT NULL,
  PRIMARY KEY (`privilege_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `privilege_setting`
--

CREATE TABLE IF NOT EXISTS `privilege_setting` (
  `PriId` int(11) NOT NULL AUTO_INCREMENT,
  `PriCode` varchar(50) NOT NULL,
  `PriName` varchar(100) NOT NULL,
  `CreatedUserId` varchar(50) NOT NULL,
  `CreatedTime` datetime NOT NULL,
  `Status` int(11) DEFAULT '0' COMMENT '-1:删除; 0:停止; 1:启动',
  `EditUserId` varchar(50) NOT NULL,
  `UpdatedTime` datetime NOT NULL,
  PRIMARY KEY (`PriId`),
  KEY `privileg_setting_PriCode_IDX` (`PriCode`,`PriName`,`Status`)
) ENGINE=InnoDB AUTO_INCREMENT=10 DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `publicprivatekey`
--

CREATE TABLE IF NOT EXISTS `publicprivatekey` (
  `PPKeyID` int(11) NOT NULL AUTO_INCREMENT,
  `Email` varchar(200) NOT NULL,
  `PublicKey` text NOT NULL,
  `PrivateKey` text NOT NULL,
  `Status` int(11) NOT NULL,
  `Ext` varchar(50) DEFAULT NULL,
  PRIMARY KEY (`PPKeyID`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `qalog`
--

CREATE TABLE IF NOT EXISTS `qalog` (
  `QId` int(11) NOT NULL,
  `ContentCount` int(11) NOT NULL,
  `QAType` varchar(50) NOT NULL,
  `UserId` int(11) NOT NULL,
  `UserAction` varchar(50) NOT NULL,
  `CreatedTime` datetime DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `question`
--

CREATE TABLE IF NOT EXISTS `question` (
  `q_id` int(4) NOT NULL AUTO_INCREMENT,
  `content` varchar(200) COLLATE utf8_unicode_ci NOT NULL,
  `content2` varchar(1000) COLLATE utf8_unicode_ci DEFAULT NULL,
  `content3` varchar(1000) COLLATE utf8_unicode_ci DEFAULT NULL,
  `content4` varchar(1000) COLLATE utf8_unicode_ci DEFAULT NULL,
  `content5` varchar(1000) COLLATE utf8_unicode_ci DEFAULT NULL,
  `content6` varchar(1000) COLLATE utf8_unicode_ci DEFAULT NULL,
  `content7` varchar(1000) COLLATE utf8_unicode_ci DEFAULT NULL,
  `content8` varchar(1000) COLLATE utf8_unicode_ci DEFAULT NULL,
  `content9` varchar(1000) COLLATE utf8_unicode_ci DEFAULT NULL,
  `content10` varchar(1000) COLLATE utf8_unicode_ci DEFAULT NULL,
  `topic_list` varchar(200) COLLATE utf8_unicode_ci DEFAULT NULL,
  `user` varchar(255) COLLATE utf8_unicode_ci NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `answer_count` smallint(5) DEFAULT NULL,
  `feature_words` text COLLATE utf8_unicode_ci,
  `modal` varchar(20) COLLATE utf8_unicode_ci DEFAULT NULL,
  `intention` varchar(20) COLLATE utf8_unicode_ci DEFAULT NULL,
  `mood` varchar(20) COLLATE utf8_unicode_ci DEFAULT NULL,
  `is_girl_friend` tinyint(1) DEFAULT NULL,
  `is_sutiable` tinyint(1) DEFAULT NULL,
  `rule` varchar(20) COLLATE utf8_unicode_ci DEFAULT NULL,
  `seg_word` text COLLATE utf8_unicode_ci,
  `feature_words_1` text COLLATE utf8_unicode_ci,
  `feature_words_2` text COLLATE utf8_unicode_ci,
  `is_modified` tinyint(1) DEFAULT '0',
  `modal_updated` varchar(50) COLLATE utf8_unicode_ci DEFAULT NULL,
  `mood1` varchar(100) COLLATE utf8_unicode_ci DEFAULT NULL,
  `emotion` text COLLATE utf8_unicode_ci,
  `topic` text COLLATE utf8_unicode_ci,
  `act` text COLLATE utf8_unicode_ci,
  `intent` text COLLATE utf8_unicode_ci,
  `qa_score` text COLLATE utf8_unicode_ci,
  `CUOutput` text COLLATE utf8_unicode_ci,
  `sentence_type` text COLLATE utf8_unicode_ci,
  `created_user` int(11) NOT NULL,
  `updated_user` int(11) NOT NULL,
  `updated_at` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00',
  PRIMARY KEY (`q_id`),
  UNIQUE KEY `question_PK` (`content`),
  KEY `index_name` (`content`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

-- --------------------------------------------------------

--
-- 資料表結構 `questionnaire`
--

CREATE TABLE IF NOT EXISTS `questionnaire` (
  `QId` varchar(50) NOT NULL,
  `UserID` varchar(50) NOT NULL,
  `QTId` varchar(50) NOT NULL,
  `QJson` text NOT NULL,
  `CreatedTime` datetime DEFAULT NULL,
  `Status` int(11) DEFAULT NULL,
  PRIMARY KEY (`QId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `realman`
--

CREATE TABLE IF NOT EXISTS `realman` (
  `RealmanID` varchar(200) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `realman_pair`
--

CREATE TABLE IF NOT EXISTS `realman_pair` (
  `RealmanID` varchar(200) DEFAULT NULL,
  `UserID` varchar(200) NOT NULL,
  `TimeOut` datetime DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `rebotprofile`
--

CREATE TABLE IF NOT EXISTS `rebotprofile` (
  `RFId` int(11) NOT NULL AUTO_INCREMENT,
  `strQ` varchar(200) DEFAULT NULL,
  `strA` varchar(1000) DEFAULT NULL,
  `Status` int(11) NOT NULL DEFAULT '1',
  PRIMARY KEY (`RFId`)
) ENGINE=InnoDB AUTO_INCREMENT=534 DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `rebotprofile_sentence_type`
--

CREATE TABLE IF NOT EXISTS `rebotprofile_sentence_type` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `q_id` int(11) DEFAULT NULL,
  `sentence_type` text,
  PRIMARY KEY (`id`),
  KEY `q_id` (`q_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `resumes`
--

CREATE TABLE IF NOT EXISTS `resumes` (
  `ResumeId` int(11) NOT NULL AUTO_INCREMENT,
  `Name` varchar(100) NOT NULL,
  `ResumeEmail` varchar(100) NOT NULL,
  `ResumeMobile` varchar(300) NOT NULL,
  `ResumePath` varchar(255) NOT NULL,
  `ResumeStatus` varchar(50) NOT NULL,
  `ResumeType` varchar(50) NOT NULL,
  `UploadTime` date NOT NULL,
  `Status` int(11) NOT NULL,
  `Comments` text NOT NULL,
  PRIMARY KEY (`ResumeId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `role`
--

CREATE TABLE IF NOT EXISTS `role` (
  `RoleId` int(11) NOT NULL AUTO_INCREMENT,
  `RoleCode` varchar(50) NOT NULL,
  `RoleName` varchar(100) NOT NULL,
  `EnterpriseId` varchar(50) NOT NULL,
  `CreatedUserId` varchar(50) NOT NULL,
  `CreatedTime` datetime NOT NULL,
  `Status` int(11) DEFAULT '0' COMMENT '-1:删除; 0:停止; 1:启动',
  `EditUserId` varchar(50) NOT NULL,
  `UpdatedTime` datetime NOT NULL,
  PRIMARY KEY (`RoleId`),
  KEY `role_RoleCode_IDX` (`RoleCode`,`RoleName`,`Status`,`EnterpriseId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `rolelog`
--

CREATE TABLE IF NOT EXISTS `rolelog` (
  `UserID` varchar(50) DEFAULT NULL,
  `role` int(11) NOT NULL,
  `CreatedTime` datetime DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `role_list`
--

CREATE TABLE IF NOT EXISTS `role_list` (
  `role_id` char(32) NOT NULL,
  `role_name` varchar(32) NOT NULL,
  `privilege` json NOT NULL,
  `enterprise_id` char(32) NOT NULL,
  PRIMARY KEY (`role_id`),
  KEY `enterprise_id_idx` (`enterprise_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `role_privilege`
--

CREATE TABLE IF NOT EXISTS `role_privilege` (
  `EnterpriseId` varchar(50) NOT NULL,
  `RoleCode` varchar(50) NOT NULL,
  `MPId` int(11) NOT NULL,
  `CreatedUserId` varchar(50) NOT NULL,
  `CreatedTime` datetime NOT NULL,
  PRIMARY KEY (`EnterpriseId`,`RoleCode`,`MPId`),
  KEY `role_privilege_RoleCode_IDX` (`EnterpriseId`,`RoleCode`,`MPId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `scenario`
--

CREATE TABLE IF NOT EXISTS `scenario` (
  `ScenarioId` int(11) NOT NULL AUTO_INCREMENT,
  `ScenarioName` varchar(100) CHARACTER SET utf8 NOT NULL,
  `CreatedTime` datetime NOT NULL,
  `Status` int(11) NOT NULL,
  PRIMARY KEY (`ScenarioId`),
  UNIQUE KEY `ScenarioName` (`ScenarioName`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- --------------------------------------------------------

--
-- 資料表結構 `scenariostate`
--

CREATE TABLE IF NOT EXISTS `scenariostate` (
  `StateId` int(11) NOT NULL,
  `EntryConditions` varchar(1000) CHARACTER SET utf8 NOT NULL,
  `EntryString` text CHARACTER SET utf8 NOT NULL,
  `OutputAnswer` text CHARACTER SET utf8 NOT NULL,
  `CollectAttributes` varchar(1000) CHARACTER SET utf8 NOT NULL,
  `Backfill` varchar(1000) CHARACTER SET utf8 NOT NULL,
  `PreviousState` varchar(255) CHARACTER SET utf8 NOT NULL,
  `NextState` varchar(255) CHARACTER SET utf8 NOT NULL,
  `DefaultNextState` int(11) NOT NULL,
  `DirectNextState` varchar(255) CHARACTER SET utf8 NOT NULL,
  `FunctionalScenario` varchar(100) CHARACTER SET utf8 NOT NULL,
  `ScenarioId` int(11) NOT NULL,
  `Status` int(11) NOT NULL,
  `CreatedTime` datetime NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- --------------------------------------------------------

--
-- 資料表結構 `sendlist`
--

CREATE TABLE IF NOT EXISTS `sendlist` (
  `SendId` int(11) NOT NULL AUTO_INCREMENT,
  `NickName` varchar(100) NOT NULL,
  `Phone` varchar(100) NOT NULL,
  `Message` varchar(200) NOT NULL,
  `SendType` int(11) NOT NULL,
  `SendTime` datetime NOT NULL,
  PRIMARY KEY (`SendId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `share_categories`
--

CREATE TABLE IF NOT EXISTS `share_categories` (
  `CategoryId` int(11) NOT NULL AUTO_INCREMENT,
  `CategoryName` varchar(200) NOT NULL,
  `ParentId` int(11) NOT NULL DEFAULT '0',
  `Status` int(11) DEFAULT '1' COMMENT '-1:删除; 0:停止; 1:启动',
  `CreatedUser` varchar(50) DEFAULT NULL,
  `CreatedTime` datetime DEFAULT NULL,
  `EditUser` varchar(50) DEFAULT NULL,
  `EditTime` datetime DEFAULT NULL,
  `CategoryPath` varchar(1000) DEFAULT NULL,
  PRIMARY KEY (`CategoryId`)
) ENGINE=InnoDB AUTO_INCREMENT=9 DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `share_intent`
--

CREATE TABLE IF NOT EXISTS `share_intent` (
  `IntentId` int(11) NOT NULL AUTO_INCREMENT,
  `IntentName` varchar(50) NOT NULL,
  `Status` int(11) DEFAULT '1' COMMENT '-1:删除; 0:停止; 1:启动',
  `CreatedUser` varchar(50) DEFAULT NULL,
  `CreatedTime` datetime DEFAULT NULL,
  `EditUser` varchar(50) DEFAULT NULL,
  `EditTime` datetime DEFAULT NULL,
  PRIMARY KEY (`IntentId`),
  KEY `user_count` (`IntentId`,`IntentName`)
) ENGINE=InnoDB AUTO_INCREMENT=125 DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `smsmessage`
--

CREATE TABLE IF NOT EXISTS `smsmessage` (
  `SMSMId` int(11) NOT NULL AUTO_INCREMENT,
  `UserId` int(11) NOT NULL,
  `Phone` text NOT NULL,
  `Message` varchar(200) NOT NULL,
  `SMSType` int(11) NOT NULL,
  `CreateTime` datetime NOT NULL,
  `UserNum` int(11) NOT NULL,
  `Status` int(11) NOT NULL,
  `SendTime` datetime DEFAULT NULL,
  `Subject` varchar(200) NOT NULL,
  PRIMARY KEY (`SMSMId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `taskengineapp`
--

CREATE TABLE IF NOT EXISTS `taskengineapp` (
  `pk` varchar(90) NOT NULL,
  `appID` varchar(50) NOT NULL,
  `scenarioID` varchar(50) NOT NULL,
  PRIMARY KEY (`pk`),
  KEY `appID` (`appID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `taskenginescenario`
--

CREATE TABLE IF NOT EXISTS `taskenginescenario` (
  `scenarioID` varchar(50) NOT NULL,
  `userID` varchar(50) NOT NULL,
  `content` mediumtext,
  `layout` mediumtext,
  `public` int(11) NOT NULL DEFAULT '0',
  `editing` tinyint(1) NOT NULL DEFAULT '0',
  `editingContent` mediumtext,
  `editingLayout` mediumtext,
  `updatetime` datetime NOT NULL DEFAULT '1970-01-01 00:00:00',
  `onoff` int(11) DEFAULT '1',
  PRIMARY KEY (`scenarioID`),
  KEY `userID` (`userID`),
  KEY `public` (`public`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `tasks`
--

CREATE TABLE IF NOT EXISTS `tasks` (
  `TaskID` int(11) NOT NULL AUTO_INCREMENT,
  `TaskType` int(11) DEFAULT NULL,
  `Description` varchar(1000) COLLATE utf8_unicode_ci DEFAULT NULL,
  `Category` int(11) DEFAULT NULL,
  `Priority` int(11) DEFAULT NULL,
  `Important` int(11) DEFAULT NULL,
  `Status` int(11) DEFAULT NULL,
  `Task_Category` varchar(50) COLLATE utf8_unicode_ci DEFAULT NULL,
  `Created_Date` datetime DEFAULT NULL,
  `Last_Modified_Date` datetime DEFAULT NULL,
  `Solved_Date` datetime DEFAULT NULL,
  `Closed_Date` datetime DEFAULT NULL,
  `Locked_Date` datetime DEFAULT NULL,
  `IssuedID` int(11) DEFAULT NULL,
  `SolvedID` int(11) DEFAULT NULL,
  `ClosedID` int(11) DEFAULT NULL,
  `Comments` text COLLATE utf8_unicode_ci,
  PRIMARY KEY (`TaskID`),
  KEY `TaskCategory` (`Category`),
  KEY `TaskPriorify` (`Priority`),
  KEY `TaskTask_Category` (`Task_Category`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

-- --------------------------------------------------------

--
-- 資料表結構 `topic`
--

CREATE TABLE IF NOT EXISTS `topic` (
  `TopicId` int(11) NOT NULL AUTO_INCREMENT,
  `TopicName` varchar(100) NOT NULL,
  `TopicCode` varchar(20) DEFAULT NULL,
  `ParentId` int(11) NOT NULL DEFAULT '0',
  `Status` int(11) NOT NULL DEFAULT '0',
  `CreatedUser` int(11) DEFAULT NULL,
  `CreatedTime` datetime DEFAULT NULL,
  `EditUser` int(11) DEFAULT NULL,
  `EditTime` datetime DEFAULT NULL,
  `TopicPath` varchar(2000) DEFAULT NULL,
  PRIMARY KEY (`TopicId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `toplist`
--

CREATE TABLE IF NOT EXISTS `toplist` (
  `NickName` varchar(100) DEFAULT NULL,
  `TopCount` int(11) DEFAULT NULL,
  `CreatedTime` datetime DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `uid_ip`
--

CREATE TABLE IF NOT EXISTS `uid_ip` (
  `UserId` varchar(50) DEFAULT NULL,
  `IP` varchar(100) DEFAULT NULL,
  `CreatedTime` datetime DEFAULT CURRENT_TIMESTAMP,
  KEY `uid_ip_index` (`UserId`,`IP`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `userchatbug`
--

CREATE TABLE IF NOT EXISTS `userchatbug` (
  `UUId` int(11) NOT NULL,
  `IP` varchar(50) NOT NULL,
  `AppId` varchar(50) NOT NULL,
  `CreatedTime` datetime DEFAULT NULL,
  `Msg` varchar(200) DEFAULT NULL,
  PRIMARY KEY (`UUId`,`IP`),
  KEY `chatbug_createdtime_status` (`AppId`,`CreatedTime`,`IP`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `userchatstatus`
--

CREATE TABLE IF NOT EXISTS `userchatstatus` (
  `UUId` int(11) NOT NULL,
  `UserId` varchar(50) DEFAULT NULL,
  `Text` varchar(1000) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `Answer` varchar(1000) DEFAULT NULL,
  `Status` int(11) NOT NULL,
  `CreatedTime` datetime DEFAULT NULL,
  `WeChatId` varchar(100) DEFAULT NULL,
  `Source` varchar(200) DEFAULT NULL,
  `Score` float DEFAULT NULL,
  `Module` varchar(200) DEFAULT NULL,
  `OldText` varchar(1000) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `IFormat` varchar(20) DEFAULT NULL,
  `OFormat` varchar(20) DEFAULT NULL,
  `FeedBack` varchar(100) DEFAULT NULL,
  `tag` varchar(20) DEFAULT NULL,
  `Owner` varchar(100) DEFAULT NULL,
  `NickName` varchar(100) DEFAULT NULL,
  `IP` varchar(50) NOT NULL,
  `AppId` varchar(50) NOT NULL,
  `SubModule` varchar(100) DEFAULT NULL,
  PRIMARY KEY (`IP`,`UUId`),
  KEY `chatstatus_createdtime_status` (`AppId`,`NickName`,`IP`),
  KEY `userchatstatus_UserId_IDX` (`UserId`,`Owner`),
  KEY `userchatstatus_CreatedTime_IDX` (`CreatedTime`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `userfeedback`
--

CREATE TABLE IF NOT EXISTS `userfeedback` (
  `UUId` int(11) NOT NULL,
  `Status` int(11) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `userlogin`
--

CREATE TABLE IF NOT EXISTS `userlogin` (
  `login_name` varchar(128) NOT NULL,
  `user_name` varchar(30) NOT NULL,
  `password` varchar(50) NOT NULL,
  `last_login_time` datetime NOT NULL,
  `status` int(11) NOT NULL DEFAULT '0'
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `usermessage`
--

CREATE TABLE IF NOT EXISTS `usermessage` (
  `MsgID` int(11) NOT NULL AUTO_INCREMENT,
  `UserID` varchar(50) NOT NULL,
  `Message` varchar(255) DEFAULT NULL,
  `OccurTime` datetime NOT NULL,
  `Status` int(11) NOT NULL,
  `MsgJson` text,
  `MsgInterval` int(11) NOT NULL DEFAULT '0',
  `Owner` varchar(50) DEFAULT NULL,
  `MsgType` varchar(50) DEFAULT NULL,
  `SendType` int(11) DEFAULT '0',
  PRIMARY KEY (`MsgID`),
  KEY `UserID` (`UserID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `users`
--

CREATE TABLE IF NOT EXISTS `users` (
  `UserId` int(11) NOT NULL AUTO_INCREMENT,
  `UserName` varchar(100) NOT NULL,
  `Email` varchar(100) NOT NULL,
  `Status` int(11) NOT NULL DEFAULT '1',
  `CanApprove` int(11) NOT NULL DEFAULT '0',
  `CreatedUser` int(11) DEFAULT NULL,
  `CreatedTime` datetime DEFAULT NULL,
  `EditUser` int(11) DEFAULT NULL,
  `EditTime` datetime DEFAULT NULL,
  `EmployeeId` varchar(200) DEFAULT NULL,
  `Password` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`UserId`),
  UNIQUE KEY `EmployeeId` (`EmployeeId`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `usertable`
--

CREATE TABLE IF NOT EXISTS `usertable` (
  `UserID` varchar(50) NOT NULL,
  `Pass` varchar(32) DEFAULT NULL,
  `Phone` varchar(20) CHARACTER SET utf8 COLLATE utf8_unicode_ci NOT NULL DEFAULT '',
  `3RD` varchar(20) DEFAULT '',
  `Account` varchar(40) DEFAULT '',
  `CreateTime` datetime NOT NULL,
  `LastLoginTime` datetime NOT NULL,
  `Status` int(11) NOT NULL,
  `BirthDay` varchar(50) DEFAULT NULL,
  `Sex` varchar(50) DEFAULT NULL,
  `NickName` varchar(200) DEFAULT NULL,
  `City` varchar(200) DEFAULT NULL,
  `Imei` varchar(200) DEFAULT NULL,
  `AndroidVersion` varchar(200) DEFAULT NULL,
  `PhoneModel` varchar(200) DEFAULT NULL,
  `Tag` text,
  `appVersion` varchar(20) DEFAULT NULL,
  `appSource` varchar(200) DEFAULT NULL,
  `Role` int(11) DEFAULT '1',
  `ChatNum` int(11) DEFAULT '0',
  `Constellation` varchar(1000) DEFAULT NULL,
  `AppId` varchar(50) DEFAULT NULL,
  `UUId` varchar(50) DEFAULT NULL,
  PRIMARY KEY (`UserID`),
  KEY `usertable_phone` (`Phone`),
  KEY `usertable_AppId_IDX` (`AppId`),
  KEY `usertable_appSource_IDX` (`appSource`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `user_list`
--

CREATE TABLE IF NOT EXISTS `user_list` (
  `user_id` char(32) NOT NULL,
  `user_name` varchar(32) NOT NULL,
  `user_type` tinyint(4) NOT NULL,
  `password` varchar(32) NOT NULL,
  `role_id` char(32) DEFAULT NULL,
  `email` varchar(256) DEFAULT NULL,
  `enterprise_id` char(32) NOT NULL,
  PRIMARY KEY (`user_id`),
  KEY `enterprise_id_idx` (`enterprise_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `user_role`
--

CREATE TABLE IF NOT EXISTS `user_role` (
  `UserId` varchar(50) NOT NULL,
  `RoleCode` varchar(50) NOT NULL,
  `EnterpriseId` varchar(50) NOT NULL,
  `CreatedTime` datetime NOT NULL,
  PRIMARY KEY (`UserId`,`RoleCode`,`EnterpriseId`),
  KEY `user_role_UserId_IDX` (`UserId`,`RoleCode`,`EnterpriseId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `vipshop_answer`
--

CREATE TABLE IF NOT EXISTS `vipshop_answer` (
  `Answer_Id` int(11) NOT NULL AUTO_INCREMENT,
  `Question_Id` int(11) NOT NULL,
  `Content` longtext COLLATE utf8_unicode_ci NOT NULL,
  `Answer_CMD` varchar(200) COLLATE utf8_unicode_ci DEFAULT NULL,
  `Begin_Time` datetime DEFAULT NULL,
  `End_Time` datetime DEFAULT NULL,
  `CreatedUser` varchar(50) COLLATE utf8_unicode_ci DEFAULT NULL,
  `CreatedTime` datetime DEFAULT NULL,
  `EditUser` varchar(50) COLLATE utf8_unicode_ci DEFAULT NULL,
  `EditTime` datetime DEFAULT NULL,
  `Status` int(11) DEFAULT '1',
  `Image_path` varchar(255) COLLATE utf8_unicode_ci DEFAULT NULL,
  `Not_Show_In_Relative_Q` tinyint(1) DEFAULT '0',
  `Content_String` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `Tags` varchar(1000) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `Answer_CMD_Msg` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`Answer_Id`),
  KEY `vipshop_answer_Question_Id_IDX` (`Question_Id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

-- --------------------------------------------------------

--
-- 資料表結構 `vipshop_answertag`
--

CREATE TABLE IF NOT EXISTS `vipshop_answertag` (
  `Answer_Id` int(11) NOT NULL,
  `Tag_Id` int(11) NOT NULL,
  `CreatedTime` datetime DEFAULT NULL,
  `Status` int(11) DEFAULT '1',
  PRIMARY KEY (`Answer_Id`,`Tag_Id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

-- --------------------------------------------------------

--
-- 資料表結構 `vipshop_categories`
--

CREATE TABLE IF NOT EXISTS `vipshop_categories` (
  `CategoryId` int(11) NOT NULL AUTO_INCREMENT,
  `CategoryName` varchar(100) NOT NULL,
  `ParentId` int(11) NOT NULL,
  `Status` int(11) NOT NULL DEFAULT '1',
  `CreatedUser` varchar(50) DEFAULT NULL,
  `CreatedTime` datetime DEFAULT NULL,
  `EditUser` varchar(50) DEFAULT NULL,
  `EditTime` datetime DEFAULT NULL,
  `level` int(11) NOT NULL DEFAULT '0',
  `ParentPath` varchar(200) NOT NULL,
  `SelfPath` varchar(200) NOT NULL,
  PRIMARY KEY (`CategoryId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `vipshop_dynamic_menu`
--

CREATE TABLE IF NOT EXISTS `vipshop_dynamic_menu` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `Answer_id` int(11) NOT NULL,
  `DynamicMenu` varchar(200) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- 資料表結構 `vipshop_onoff`
--

CREATE TABLE IF NOT EXISTS `vipshop_onoff` (
  `OnOff_Id` int(11) NOT NULL AUTO_INCREMENT,
  `OnOff_Code` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `OnOff_Name` varchar(200) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `OnOff_Status` int(11) DEFAULT NULL,
  `OnOff_Remark` text COLLATE utf8mb4_unicode_ci,
  `OnOff_Scenario` varchar(200) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `OnOff_NumType` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `OnOff_Num` int(11) DEFAULT '0',
  `OnOff_Msg` text COLLATE utf8mb4_unicode_ci,
  `OnOff_Flow` int(11) DEFAULT '0',
  `OnOff_WhiteList` text COLLATE utf8mb4_unicode_ci,
  `OnOff_BlackList` text COLLATE utf8mb4_unicode_ci,
  `UpdateTime` datetime DEFAULT NULL,
  PRIMARY KEY (`OnOff_Id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- 資料表結構 `vipshop_question`
--

CREATE TABLE IF NOT EXISTS `vipshop_question` (
  `Question_Id` int(11) NOT NULL AUTO_INCREMENT,
  `Content` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `CategoryId` int(11) DEFAULT NULL,
  `SQuestion_count` smallint(5) DEFAULT NULL,
  `CreatedUser` varchar(50) COLLATE utf8_unicode_ci DEFAULT NULL,
  `CreatedTime` datetime DEFAULT NULL,
  `EditUser` varchar(50) COLLATE utf8_unicode_ci DEFAULT NULL,
  `EditTime` datetime DEFAULT NULL,
  `Status` int(11) DEFAULT '1',
  PRIMARY KEY (`Question_Id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

-- --------------------------------------------------------

--
-- 資料表結構 `vipshop_related_question`
--

CREATE TABLE IF NOT EXISTS `vipshop_related_question` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `Answer_id` int(11) NOT NULL,
  `RelatedQuestion` varchar(200) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- 資料表結構 `vipshop_robot_setting`
--

CREATE TABLE IF NOT EXISTS `vipshop_robot_setting` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `content` text COLLATE utf8mb4_unicode_ci,
  `type` int(4) DEFAULT '1',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- 資料表結構 `vipshop_squestion`
--

CREATE TABLE IF NOT EXISTS `vipshop_squestion` (
  `SQ_Id` int(11) NOT NULL AUTO_INCREMENT,
  `Question_Id` int(11) NOT NULL,
  `Content` varchar(200) COLLATE utf8_unicode_ci NOT NULL,
  `CreatedUser` varchar(50) COLLATE utf8_unicode_ci DEFAULT NULL,
  `CreatedTime` datetime DEFAULT NULL,
  `EditUser` varchar(50) COLLATE utf8_unicode_ci DEFAULT NULL,
  `EditTime` datetime DEFAULT NULL,
  `Status` int(11) DEFAULT '1',
  PRIMARY KEY (`SQ_Id`),
  KEY `index_name` (`Content`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

-- --------------------------------------------------------

--
-- 資料表結構 `vipshop_tag`
--

CREATE TABLE IF NOT EXISTS `vipshop_tag` (
  `Tag_Id` int(11) NOT NULL AUTO_INCREMENT,
  `Tag_Name` varchar(100) COLLATE utf8_unicode_ci NOT NULL,
  `CreatedUser` varchar(50) COLLATE utf8_unicode_ci DEFAULT NULL,
  `CreatedTime` datetime DEFAULT NULL,
  `EditUser` varchar(50) COLLATE utf8_unicode_ci DEFAULT NULL,
  `EditTime` datetime DEFAULT NULL,
  `Status` int(11) DEFAULT '1',
  `Tag_Type` int(4) NOT NULL,
  PRIMARY KEY (`Tag_Id`),
  UNIQUE KEY `tag_PK` (`Tag_Name`),
  KEY `tag` (`Tag_Name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

-- --------------------------------------------------------

--
-- 資料表結構 `vipshop_tag_type`
--

CREATE TABLE IF NOT EXISTS `vipshop_tag_type` (
  `Type_id` int(4) NOT NULL AUTO_INCREMENT,
  `Type_name` varchar(100) COLLATE utf8_unicode_ci NOT NULL,
  PRIMARY KEY (`Type_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

-- --------------------------------------------------------

--
-- 資料表結構 `vipshop_tmp`
--

CREATE TABLE IF NOT EXISTS `vipshop_tmp` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `Content1` longtext COLLATE utf8_unicode_ci NOT NULL,
  `Content2` longtext COLLATE utf8_unicode_ci NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

-- --------------------------------------------------------

--
-- 資料表結構 `vip_answer`
--

CREATE TABLE IF NOT EXISTS `vip_answer` (
  `Answer_Id` int(11) NOT NULL AUTO_INCREMENT,
  `Question_Id` int(11) NOT NULL,
  `Content` longtext COLLATE utf8_unicode_ci NOT NULL,
  `Answer_CMD` varchar(200) COLLATE utf8_unicode_ci DEFAULT NULL,
  `Begin_Time` datetime DEFAULT NULL,
  `End_Time` datetime DEFAULT NULL,
  `CreatedUser` varchar(50) COLLATE utf8_unicode_ci DEFAULT NULL,
  `CreatedTime` datetime DEFAULT NULL,
  `EditUser` varchar(50) COLLATE utf8_unicode_ci DEFAULT NULL,
  `EditTime` datetime DEFAULT NULL,
  `Status` int(11) DEFAULT '1',
  `Image_path` varchar(255) COLLATE utf8_unicode_ci DEFAULT NULL,
  `Not_Show_In_Relative_Q` tinyint(1) DEFAULT '0',
  `Content_String` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `Tags` varchar(1000) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `Answer_CMD_Msg` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`Answer_Id`),
  KEY `vip_answer_Question_Id_IDX` (`Question_Id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

-- --------------------------------------------------------

--
-- 資料表結構 `vip_answertag`
--

CREATE TABLE IF NOT EXISTS `vip_answertag` (
  `Answer_Id` int(11) NOT NULL,
  `Tag_Id` int(11) NOT NULL,
  `CreatedTime` datetime DEFAULT NULL,
  `Status` int(11) DEFAULT '1',
  PRIMARY KEY (`Answer_Id`,`Tag_Id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

-- --------------------------------------------------------

--
-- 資料表結構 `vip_categories`
--

CREATE TABLE IF NOT EXISTS `vip_categories` (
  `CategoryId` int(11) NOT NULL AUTO_INCREMENT,
  `CategoryName` varchar(100) NOT NULL,
  `ParentId` int(11) NOT NULL,
  `Status` int(11) NOT NULL DEFAULT '1',
  `CreatedUser` varchar(50) DEFAULT NULL,
  `CreatedTime` datetime DEFAULT NULL,
  `EditUser` varchar(50) DEFAULT NULL,
  `EditTime` datetime DEFAULT NULL,
  `level` int(11) NOT NULL DEFAULT '0',
  `ParentPath` varchar(200) NOT NULL,
  `SelfPath` varchar(200) NOT NULL,
  PRIMARY KEY (`CategoryId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `vip_dynamic_menu`
--

CREATE TABLE IF NOT EXISTS `vip_dynamic_menu` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `Answer_id` int(11) NOT NULL,
  `DynamicMenu` varchar(200) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- 資料表結構 `vip_module`
--

CREATE TABLE IF NOT EXISTS `vip_module` (
  `ModuleId` int(11) NOT NULL AUTO_INCREMENT,
  `ModuleCode` varchar(50) NOT NULL,
  `ModuleName` varchar(100) NOT NULL,
  `ParentCode` varchar(50) NOT NULL,
  `ModuleUrl` varchar(500) NOT NULL,
  `CreatedUserId` varchar(50) NOT NULL,
  `CreatedTime` datetime NOT NULL,
  `Status` int(11) DEFAULT '0' COMMENT '-1:删除; 0:停止; 1:启动',
  `EditUserId` varchar(50) NOT NULL,
  `UpdatedTime` datetime NOT NULL,
  PRIMARY KEY (`ModuleId`),
  KEY `vip_module_ModuleCode_IDX` (`ModuleCode`,`ModuleName`,`ParentCode`,`Status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `vip_module_privilege`
--

CREATE TABLE IF NOT EXISTS `vip_module_privilege` (
  `MPId` int(11) NOT NULL AUTO_INCREMENT,
  `ModuleCode` varchar(50) NOT NULL,
  `PriCode` varchar(50) NOT NULL,
  `CreatedUserId` varchar(50) NOT NULL,
  `CreatedTime` datetime NOT NULL,
  PRIMARY KEY (`MPId`),
  KEY `vip_module_privilege_ModuleCode_IDX` (`ModuleCode`,`PriCode`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `vip_onoff`
--

CREATE TABLE IF NOT EXISTS `vip_onoff` (
  `OnOff_Id` int(11) NOT NULL AUTO_INCREMENT,
  `OnOff_Code` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `OnOff_Name` varchar(200) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `OnOff_Status` int(11) DEFAULT '0',
  `OnOff_Remark` text COLLATE utf8mb4_unicode_ci,
  `OnOff_Scenario` varchar(200) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `OnOff_NumType` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `OnOff_Num` int(11) DEFAULT '0',
  `OnOff_Msg` text COLLATE utf8mb4_unicode_ci,
  `OnOff_Flow` int(11) DEFAULT '0',
  `OnOff_WhiteList` text COLLATE utf8mb4_unicode_ci,
  `OnOff_BlackList` text COLLATE utf8mb4_unicode_ci,
  `UpdateTime` datetime DEFAULT NULL,
  PRIMARY KEY (`OnOff_Id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- 資料表結構 `vip_question`
--

CREATE TABLE IF NOT EXISTS `vip_question` (
  `Question_Id` int(11) NOT NULL AUTO_INCREMENT,
  `Content` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `CategoryId` int(11) DEFAULT NULL,
  `SQuestion_count` smallint(5) DEFAULT NULL,
  `CreatedUser` varchar(50) COLLATE utf8_unicode_ci DEFAULT NULL,
  `CreatedTime` datetime DEFAULT NULL,
  `EditUser` varchar(50) COLLATE utf8_unicode_ci DEFAULT NULL,
  `EditTime` datetime DEFAULT NULL,
  `Status` int(11) DEFAULT '1',
  PRIMARY KEY (`Question_Id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

-- --------------------------------------------------------

--
-- 資料表結構 `vip_related_question`
--

CREATE TABLE IF NOT EXISTS `vip_related_question` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `Answer_id` int(11) NOT NULL,
  `RelatedQuestion` varchar(200) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- 資料表結構 `vip_robot_setting`
--

CREATE TABLE IF NOT EXISTS `vip_robot_setting` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `content` text COLLATE utf8mb4_unicode_ci,
  `type` int(4) DEFAULT '1',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- 資料表結構 `vip_squestion`
--

CREATE TABLE IF NOT EXISTS `vip_squestion` (
  `SQ_Id` int(11) NOT NULL AUTO_INCREMENT,
  `Question_Id` int(11) NOT NULL,
  `Content` varchar(200) COLLATE utf8_unicode_ci NOT NULL,
  `CreatedUser` varchar(50) COLLATE utf8_unicode_ci DEFAULT NULL,
  `CreatedTime` datetime DEFAULT NULL,
  `EditUser` varchar(50) COLLATE utf8_unicode_ci DEFAULT NULL,
  `EditTime` datetime DEFAULT NULL,
  `Status` int(11) DEFAULT '1',
  PRIMARY KEY (`SQ_Id`),
  KEY `index_name` (`Content`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

-- --------------------------------------------------------

--
-- 資料表結構 `vip_tag`
--

CREATE TABLE IF NOT EXISTS `vip_tag` (
  `Tag_Id` int(11) NOT NULL AUTO_INCREMENT,
  `Tag_Name` varchar(100) COLLATE utf8_unicode_ci NOT NULL,
  `CreatedUser` varchar(50) COLLATE utf8_unicode_ci DEFAULT NULL,
  `CreatedTime` datetime DEFAULT NULL,
  `EditUser` varchar(50) COLLATE utf8_unicode_ci DEFAULT NULL,
  `EditTime` datetime DEFAULT NULL,
  `Status` int(11) DEFAULT '1',
  `Tag_Type` int(4) NOT NULL,
  PRIMARY KEY (`Tag_Id`),
  UNIQUE KEY `tag_PK` (`Tag_Name`),
  KEY `tag` (`Tag_Name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

-- --------------------------------------------------------

--
-- 資料表結構 `vip_tag_type`
--

CREATE TABLE IF NOT EXISTS `vip_tag_type` (
  `Type_id` int(4) NOT NULL AUTO_INCREMENT,
  `Type_name` varchar(100) COLLATE utf8_unicode_ci NOT NULL,
  PRIMARY KEY (`Type_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

-- --------------------------------------------------------

--
-- 資料表結構 `vip_tmp`
--

CREATE TABLE IF NOT EXISTS `vip_tmp` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `Content1` longtext COLLATE utf8_unicode_ci NOT NULL,
  `Content2` longtext COLLATE utf8_unicode_ci NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

-- --------------------------------------------------------

--
-- 資料表結構 `wechatimage`
--

CREATE TABLE IF NOT EXISTS `wechatimage` (
  `WImageID` int(11) NOT NULL AUTO_INCREMENT,
  `OpenID` varchar(200) NOT NULL,
  `smile` int(11) NOT NULL,
  `male` int(11) NOT NULL,
  `human` int(11) NOT NULL,
  `flag` int(11) NOT NULL,
  `Status` int(11) NOT NULL,
  `Type` int(11) NOT NULL,
  `CreateTime` datetime NOT NULL,
  `confidence` varchar(100) DEFAULT NULL,
  `sourceImg` int(11) DEFAULT NULL,
  PRIMARY KEY (`WImageID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `wechatpublicusers`
--

CREATE TABLE IF NOT EXISTS `wechatpublicusers` (
  `UserId` int(11) NOT NULL AUTO_INCREMENT,
  `AppId` varchar(50) NOT NULL,
  `OpenId` varchar(50) NOT NULL,
  `NickName` varchar(50) DEFAULT NULL,
  `CreatedTime` datetime DEFAULT NULL,
  `Status` int(11) NOT NULL DEFAULT '1',
  `service_type_info` varchar(50) DEFAULT NULL,
  PRIMARY KEY (`UserId`),
  UNIQUE KEY `wechatpublicusers_PK` (`OpenId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `wechatusers`
--

CREATE TABLE IF NOT EXISTS `wechatusers` (
  `OpenID` varchar(100) NOT NULL DEFAULT '',
  `NickName` varchar(100) DEFAULT NULL,
  `Gender` int(11) DEFAULT NULL,
  `SubscribeTime` datetime DEFAULT NULL,
  `Province` varchar(100) DEFAULT NULL,
  `Owner` varchar(100) DEFAULT NULL,
  `UUId` varchar(50) NOT NULL,
  `Status` int(11) NOT NULL,
  PRIMARY KEY (`UUId`),
  UNIQUE KEY `wechatusers_pk` (`OpenID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `wechatuserstest`
--

CREATE TABLE IF NOT EXISTS `wechatuserstest` (
  `OpenID` varchar(100) NOT NULL DEFAULT '',
  `NickName` varchar(100) DEFAULT NULL,
  `Gender` int(11) DEFAULT NULL,
  `SubscribeTime` datetime DEFAULT NULL,
  `Province` varchar(100) DEFAULT NULL,
  `Owner` varchar(100) DEFAULT NULL,
  PRIMARY KEY (`OpenID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `wechatvoice`
--

CREATE TABLE IF NOT EXISTS `wechatvoice` (
  `VoiceID` int(11) NOT NULL AUTO_INCREMENT,
  `OpenID` varchar(200) NOT NULL,
  `msg` varchar(200) DEFAULT NULL,
  `gender` int(11) DEFAULT NULL,
  `CreateTime` datetime NOT NULL,
  `Status` int(11) NOT NULL,
  `tag` varchar(20) DEFAULT NULL,
  PRIMARY KEY (`VoiceID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `workload`
--

CREATE TABLE IF NOT EXISTS `workload` (
  `WorkTabelID` int(11) NOT NULL,
  `InternID` int(11) NOT NULL,
  `Workload` float(10,2) NOT NULL,
  `CreatedTime` datetime NOT NULL,
  `Status` int(11) NOT NULL,
  `Name` varchar(50) DEFAULT NULL,
  KEY `workload_WorkTabelID_IDX` (`WorkTabelID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `workmessage`
--

CREATE TABLE IF NOT EXISTS `workmessage` (
  `WorkID` int(11) NOT NULL AUTO_INCREMENT,
  `Message` varchar(1000) DEFAULT NULL,
  `Price` float(9,3) NOT NULL,
  `Principal` varchar(50) NOT NULL,
  `Interns` varchar(500) NOT NULL,
  `CreatedUser` int(11) NOT NULL,
  `CreatedTime` datetime NOT NULL,
  `EditUser` int(11) NOT NULL,
  `EditTime` datetime NOT NULL,
  `Status` int(11) NOT NULL,
  `WorkName` varchar(200) NOT NULL,
  `Unit` varchar(50) DEFAULT NULL,
  PRIMARY KEY (`WorkID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `worktable`
--

CREATE TABLE IF NOT EXISTS `worktable` (
  `WorkTabelID` int(11) NOT NULL AUTO_INCREMENT,
  `WorkID` int(11) NOT NULL,
  `WorkTime` datetime NOT NULL,
  `Remark` varchar(1000) DEFAULT NULL,
  `CreatedUser` int(11) NOT NULL,
  `CreatedTime` datetime NOT NULL,
  `EditUser` int(11) NOT NULL,
  `EditTime` datetime NOT NULL,
  `Status` int(11) NOT NULL,
  PRIMARY KEY (`WorkTabelID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `youzu_mm_prize`
--

CREATE TABLE IF NOT EXISTS `youzu_mm_prize` (
  `MMPrizeId` int(11) NOT NULL AUTO_INCREMENT,
  `WeChatId` varchar(50) NOT NULL,
  `MMUserId` varchar(50) NOT NULL,
  `MMOsdkUserId` varchar(50) NOT NULL,
  `MMServerZone` varchar(50) DEFAULT NULL,
  `CreatedTime` datetime DEFAULT NULL,
  PRIMARY KEY (`MMPrizeId`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- --------------------------------------------------------

--
-- 資料表結構 `youzu_mm_user`
--

CREATE TABLE IF NOT EXISTS `youzu_mm_user` (
  `MMId` int(11) NOT NULL AUTO_INCREMENT,
  `WeChatId` varchar(50) NOT NULL,
  `MMUserId` varchar(50) NOT NULL,
  `MMAppId` varchar(50) NOT NULL,
  `MMOsdkUserId` varchar(50) NOT NULL,
  `MMNickName` varchar(50) DEFAULT NULL,
  `MMSex` int(11) DEFAULT NULL,
  `MMChannelId` varchar(50) NOT NULL,
  `MMIp` varchar(50) NOT NULL,
  `Status` int(11) NOT NULL,
  `CreatedTime` datetime DEFAULT NULL,
  `EditTime` datetime DEFAULT NULL,
  `MMServerZone` varchar(50) DEFAULT NULL,
  PRIMARY KEY (`MMId`),
  KEY `youzu_mm_user_id` (`MMUserId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `yyb_bot_log`
--

CREATE TABLE IF NOT EXISTS `yyb_bot_log` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `log_content` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `log_time` datetime DEFAULT CURRENT_TIMESTAMP,
  `log_reply` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

--
-- 已匯出資料表的限制(Constraint)
--

--
-- 資料表的 Constraints `rebotprofile_sentence_type`
--
ALTER TABLE `rebotprofile_sentence_type`
  ADD CONSTRAINT `rebotprofile_sentence_type_ibfk_1` FOREIGN KEY (`q_id`) REFERENCES `rebotprofile` (`RFId`) ON DELETE CASCADE ON UPDATE CASCADE;
COMMIT;

/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
