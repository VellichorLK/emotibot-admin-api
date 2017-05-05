-- phpMyAdmin SQL Dump
-- version 4.7.0
-- https://www.phpmyadmin.net/
--
-- 主機: db
-- 產生時間： 2017 年 05 月 03 日 10:36
-- 伺服器版本: 5.7.17
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
-- 資料庫： `backend_log`
--
CREATE DATABASE IF NOT EXISTS `backend_log` DEFAULT CHARACTER SET utf8 COLLATE utf8_general_ci;
USE `backend_log`;

-- --------------------------------------------------------

--
-- 資料表結構 `chat_record`
--

CREATE TABLE `chat_record` (
  `idchatlog` int(11) NOT NULL,
  `user_id` char(64) NOT NULL,
  `user_Q` text NOT NULL,
  `std_Q` char(255) DEFAULT NULL,
  `answer` mediumtext NOT NULL,
  `module` char(32) NOT NULL,
  `emotion` char(32) DEFAULT NULL,
  `created_time` char(32) NOT NULL,
  `score` int(11) DEFAULT NULL,
  `custom_info` text,
  `host` char(32) NOT NULL,
  `unique_id` char(16) NOT NULL,
  `note` text
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

--
-- 已匯出資料表的索引
--

--
-- 資料表索引 `chat_record`
--
ALTER TABLE `chat_record`
  ADD PRIMARY KEY (`idchatlog`);

--
-- 在匯出的資料表使用 AUTO_INCREMENT
--

--
-- 使用資料表 AUTO_INCREMENT `chat_record`
--
ALTER TABLE `chat_record`
  MODIFY `idchatlog` int(11) NOT NULL AUTO_INCREMENT;COMMIT;

/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
