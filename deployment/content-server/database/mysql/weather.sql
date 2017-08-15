-- phpMyAdmin SQL Dump
-- version 4.6.6
-- https://www.phpmyadmin.net/
--
-- 主機: db
-- 產生時間： 2017 年 04 月 13 日 06:11
-- 伺服器版本: 5.7.17
-- PHP 版本： 7.0.16

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
SET time_zone = "+00:00";


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;

--
-- 資料庫： `weather`
--
CREATE DATABASE IF NOT EXISTS `weather` DEFAULT CHARACTER SET utf8 COLLATE utf8_general_ci;
USE `weather`;

-- --------------------------------------------------------

--
-- 資料表結構 `city_list`
--

CREATE TABLE IF NOT EXISTS `city_list` (
  `city_id` text NOT NULL,
  `city_name` text NOT NULL,
  `province` text NOT NULL,
  `country` text NOT NULL,
  `lat` float NOT NULL,
  `lon` float NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `forecast`
--

CREATE TABLE IF NOT EXISTS `forecast` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `city_id` text,
  `city_name` text,
  `date` text,
  `fetch_time` text,
  `generation` int(11) DEFAULT NULL,
  `day_sky` text,
  `day_temperature` float DEFAULT NULL,
  `day_wind_direction` varchar(10) DEFAULT NULL,
  `day_wind_level` int(11) DEFAULT NULL,
  `day_sun_type` text,
  `day_sun_hour` int(11) DEFAULT NULL,
  `day_sun_minute` int(11) DEFAULT NULL,
  `day_air_level_cn` varchar(20) DEFAULT NULL,
  `day_air_level` int(11) DEFAULT NULL,
  `night_sky` text,
  `night_temperature` float DEFAULT NULL,
  `night_wind_direction` varchar(10) DEFAULT NULL,
  `night_wind_level` int(11) DEFAULT NULL,
  `night_sun_type` text,
  `night_sun_hour` int(11) DEFAULT NULL,
  `night_sun_minute` int(11) DEFAULT NULL,
  `min_temperature` float DEFAULT NULL,
  `max_temperature` float DEFAULT NULL,
  `now_hour` int(11) DEFAULT NULL,
  `now_aqi` int(11) DEFAULT NULL,
  `now_temperature` float DEFAULT NULL,
  `now_wind_direction` varchar(10) DEFAULT NULL,
  `now_wind_level` int(11) DEFAULT NULL,
  `sugg_uv_name` varchar(50) DEFAULT NULL,
  `sugg_uv_value` varchar(20) DEFAULT NULL,
  `sugg_uv_text` text,
  `sugg_flu_name` varchar(50) DEFAULT NULL,
  `sugg_flu_value` varchar(20) DEFAULT NULL,
  `sugg_flu_text` text,
  `sugg_car_washing_name` varchar(50) DEFAULT NULL,
  `sugg_car_washing_value` varchar(20) DEFAULT NULL,
  `sugg_car_washing_text` text,
  `sugg_cloth_name` varchar(50) DEFAULT NULL,
  `sugg_cloth_value` varchar(20) DEFAULT NULL,
  `sugg_cloth_text` text,
  `sugg_driving_name` varchar(50) DEFAULT NULL,
  `sugg_driving_value` varchar(20) DEFAULT NULL,
  `sugg_driving_text` text,
  `sugg_air_name` varchar(50) DEFAULT NULL,
  `sugg_air_value` varchar(20) DEFAULT NULL,
  `sugg_air_text` text,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `realtime`
--

CREATE TABLE IF NOT EXISTS `realtime` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `city_id` text,
  `city_name` text,
  `date` text,
  `fetch_time` text,
  `generation` int(11) DEFAULT NULL,
  `day_sky` text,
  `day_temperature` float DEFAULT NULL,
  `day_wind_direction` varchar(10) DEFAULT NULL,
  `day_wind_level` int(11) DEFAULT NULL,
  `day_sun_type` text,
  `day_sun_hour` int(11) DEFAULT NULL,
  `day_sun_minute` int(11) DEFAULT NULL,
  `day_air_level_cn` varchar(20) DEFAULT NULL,
  `day_air_level` int(11) DEFAULT NULL,
  `night_sky` text,
  `night_temperature` float DEFAULT NULL,
  `night_wind_direction` varchar(10) DEFAULT NULL,
  `night_wind_level` int(11) DEFAULT NULL,
  `night_sun_type` text,
  `night_sun_hour` int(11) DEFAULT NULL,
  `night_sun_minute` int(11) DEFAULT NULL,
  `min_temperature` float DEFAULT NULL,
  `max_temperature` float DEFAULT NULL,
  `now_hour` int(11) DEFAULT NULL,
  `now_aqi` int(11) DEFAULT NULL,
  `now_temperature` float DEFAULT NULL,
  `now_wind_direction` varchar(10) DEFAULT NULL,
  `now_wind_level` int(11) DEFAULT NULL,
  `sugg_uv_name` varchar(50) DEFAULT NULL,
  `sugg_uv_value` varchar(20) DEFAULT NULL,
  `sugg_uv_text` text,
  `sugg_flu_name` varchar(50) DEFAULT NULL,
  `sugg_flu_value` varchar(20) DEFAULT NULL,
  `sugg_flu_text` text,
  `sugg_car_washing_name` varchar(50) DEFAULT NULL,
  `sugg_car_washing_value` varchar(20) DEFAULT NULL,
  `sugg_car_washing_text` text,
  `sugg_cloth_name` varchar(50) DEFAULT NULL,
  `sugg_cloth_value` varchar(20) DEFAULT NULL,
  `sugg_cloth_text` text,
  `sugg_driving_name` varchar(50) DEFAULT NULL,
  `sugg_driving_value` varchar(20) DEFAULT NULL,
  `sugg_driving_text` text,
  `sugg_air_name` varchar(50) DEFAULT NULL,
  `sugg_air_value` varchar(20) DEFAULT NULL,
  `sugg_air_text` text,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
