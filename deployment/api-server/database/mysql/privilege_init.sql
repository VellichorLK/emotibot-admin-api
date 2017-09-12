-- phpMyAdmin SQL Dump
-- version 4.7.2
-- https://www.phpmyadmin.net/
--
-- 主機: db
-- 產生時間： 2017 年 09 月 12 日 08:34
-- 伺服器版本: 8.0.2-dmr
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
-- 資料庫： `authentication`
--
CREATE DATABASE IF NOT EXISTS `authentication` DEFAULT CHARACTER SET utf8 COLLATE utf8_general_ci;
USE `authentication`;

-- --------------------------------------------------------

--
-- 資料表結構 `privilege_list`
--

CREATE TABLE IF NOT EXISTS `privilege_list` (
  `privilege_id` int(11) NOT NULL,
  `privilege_name` varchar(32) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

--
-- 資料表的匯出資料 `privilege_list`
--

INSERT INTO `privilege_list` (`privilege_id`, `privilege_name`) VALUES
(1, 'voiceCheck'),
(2, 'voiceUpload'),
(3, 'voiceQueue'),
(4, 'voiceReport'),
(5, 'systemConfig'),
(6, 'authConfig');

--
-- 已匯出資料表的索引
--

--
-- 資料表索引 `privilege_list`
--
ALTER TABLE `privilege_list`
  ADD PRIMARY KEY (`privilege_id`);
COMMIT;

/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
