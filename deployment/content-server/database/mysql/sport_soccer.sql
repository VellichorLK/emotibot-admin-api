-- phpMyAdmin SQL Dump
-- version 4.6.6
-- https://www.phpmyadmin.net/
--
-- 主機: db
-- 產生時間： 2017 年 03 月 29 日 06:07
-- 伺服器版本: 5.7.15
-- PHP 版本： 7.0.16

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
SET time_zone = "+00:00";


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;

--
-- 資料庫： `sport_soccer`
--
CREATE DATABASE IF NOT EXISTS `sport_soccer` DEFAULT CHARACTER SET utf8 COLLATE utf8_general_ci;
USE `sport_soccer`;

-- --------------------------------------------------------

--
-- 資料表結構 `crawled_live`
--

CREATE TABLE `crawled_live` (
  `id` int(11) UNSIGNED NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `crawled_statistic`
--

CREATE TABLE `crawled_statistic` (
  `id` int(11) UNSIGNED NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `json_preparation`
--

CREATE TABLE `json_preparation` (
  `match_id` int(11) NOT NULL,
  `team1_name` varchar(50) DEFAULT NULL,
  `team2_name` varchar(50) DEFAULT NULL,
  `team1_id` int(11) DEFAULT NULL,
  `team2_id` int(11) DEFAULT NULL,
  `time` datetime DEFAULT NULL,
  `json` text,
  `update_time` datetime DEFAULT NULL,
  `md5` varchar(50) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `leagues_info`
--

CREATE TABLE `leagues_info` (
  `id` int(11) UNSIGNED NOT NULL,
  `name` varchar(20) DEFAULT NULL,
  `current_league` varchar(20) DEFAULT NULL,
  `max_rnd` int(11) DEFAULT NULL,
  `cur_rnd` int(11) DEFAULT NULL,
  `sl_league_type` int(11) DEFAULT NULL,
  `fetch_time` datetime DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `live_match_event`
--

CREATE TABLE `live_match_event` (
  `id` int(11) UNSIGNED NOT NULL,
  `live_id` int(11) DEFAULT NULL,
  `event_id` int(11) DEFAULT NULL,
  `event_name` varchar(20) DEFAULT NULL,
  `q_id` int(11) DEFAULT NULL,
  `q_name` varchar(20) DEFAULT NULL,
  `event_type` int(11) DEFAULT NULL,
  `team_id` int(11) DEFAULT NULL,
  `team_name_cn` varchar(50) DEFAULT NULL,
  `player_id` int(11) DEFAULT NULL,
  `player_name_cn` varchar(50) DEFAULT NULL,
  `position_x` float DEFAULT NULL,
  `position_y` float DEFAULT NULL,
  `minute` int(11) DEFAULT NULL,
  `second` int(11) DEFAULT NULL,
  `goal_y` float DEFAULT NULL,
  `goal_z` float DEFAULT NULL,
  `pass_player_id` int(11) DEFAULT NULL,
  `pass_x` float DEFAULT NULL,
  `pass_y` float DEFAULT NULL,
  `period_id` int(11) DEFAULT NULL,
  `description` text,
  `fetch_time` datetime DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `live_match_info`
--

CREATE TABLE `live_match_info` (
  `id` int(11) UNSIGNED NOT NULL,
  `league_type` int(11) DEFAULT NULL,
  `league_type_name` varchar(50) DEFAULT NULL,
  `season` int(11) DEFAULT NULL,
  `round` int(11) DEFAULT NULL,
  `round_name` varchar(50) DEFAULT NULL,
  `group` varchar(10) DEFAULT NULL,
  `status` int(11) DEFAULT NULL,
  `match_status` int(11) DEFAULT NULL,
  `status_name` varchar(50) DEFAULT NULL,
  `team1_id` int(11) DEFAULT NULL,
  `team2_id` int(11) DEFAULT NULL,
  `score1` int(11) DEFAULT NULL,
  `score2` int(11) DEFAULT NULL,
  `if_hot` int(11) DEFAULT NULL,
  `live_mode` int(11) DEFAULT NULL,
  `match_type` int(11) DEFAULT NULL,
  `uc_id` int(11) DEFAULT NULL,
  `discipline` varchar(50) DEFAULT NULL,
  `discipline_name` varchar(50) DEFAULT NULL,
  `title` varchar(100) DEFAULT NULL,
  `time` datetime DEFAULT NULL,
  `team1_name` varchar(50) DEFAULT NULL,
  `team2_name` varchar(50) DEFAULT NULL,
  `live_url` varchar(100) DEFAULT NULL,
  `news_url` varchar(100) DEFAULT NULL,
  `match_city` varchar(50) DEFAULT NULL,
  `stadium` varchar(50) DEFAULT NULL,
  `opta_id` int(11) DEFAULT NULL,
  `odds_id` int(11) DEFAULT NULL,
  `odds_url` varchar(100) DEFAULT NULL,
  `fetch_time` datetime DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `live_teamformation`
--

CREATE TABLE `live_teamformation` (
  `id` int(11) UNSIGNED NOT NULL,
  `live_id` int(11) DEFAULT NULL,
  `player_id` int(11) DEFAULT NULL,
  `player_f_name` varchar(50) DEFAULT NULL,
  `player_l_name` varchar(50) DEFAULT NULL,
  `player_name_cn` varchar(50) DEFAULT NULL,
  `shirt_number` int(11) DEFAULT NULL,
  `sl_team_id` int(11) DEFAULT NULL,
  `team_id` int(11) DEFAULT NULL,
  `match_id` int(11) DEFAULT NULL,
  `up_time` int(11) DEFAULT NULL,
  `down_time` int(11) DEFAULT NULL,
  `pic` varchar(100) DEFAULT NULL,
  `position_id` int(11) DEFAULT NULL,
  `position_cn` varchar(20) DEFAULT NULL,
  `position_long_cn` varchar(20) DEFAULT NULL,
  `pos_xy` int(11) DEFAULT NULL,
  `status` varchar(20) DEFAULT NULL,
  `status_name` varchar(20) DEFAULT NULL,
  `fetch_time` datetime DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `match_info`
--

CREATE TABLE `match_info` (
  `id` int(11) UNSIGNED NOT NULL,
  `league_type` int(11) DEFAULT NULL,
  `league_type_name` varchar(50) DEFAULT NULL,
  `season` int(11) DEFAULT NULL,
  `round` int(11) DEFAULT NULL,
  `round_name` varchar(50) DEFAULT NULL,
  `group` varchar(10) DEFAULT NULL,
  `status` int(11) DEFAULT NULL,
  `match_status` int(11) DEFAULT NULL,
  `status_name` varchar(50) DEFAULT NULL,
  `team1_id` int(11) DEFAULT NULL,
  `team2_id` int(11) DEFAULT NULL,
  `score1` int(11) DEFAULT NULL,
  `score2` int(11) DEFAULT NULL,
  `if_hot` int(11) DEFAULT NULL,
  `live_mode` int(11) DEFAULT NULL,
  `discipline` varchar(50) DEFAULT NULL,
  `discipline_name` varchar(50) DEFAULT NULL,
  `title` varchar(100) DEFAULT NULL,
  `time` datetime DEFAULT NULL,
  `team1_name` varchar(50) DEFAULT NULL,
  `team2_name` varchar(50) DEFAULT NULL,
  `live_url` varchar(100) DEFAULT NULL,
  `news_url` varchar(100) DEFAULT NULL,
  `match_city` varchar(50) DEFAULT NULL,
  `stadium` varchar(50) DEFAULT NULL,
  `opta_id` int(11) DEFAULT NULL,
  `odds_id` int(11) DEFAULT NULL,
  `odds_url` varchar(100) DEFAULT NULL,
  `fetch_time` datetime DEFAULT NULL,
  `is_sent` int(11) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `match_nickname`
--

CREATE TABLE `match_nickname` (
  `id` int(11) UNSIGNED NOT NULL,
  `team1_name` varchar(50) DEFAULT NULL,
  `team2_name` varchar(50) DEFAULT NULL,
  `nickname` varchar(50) DEFAULT NULL,
  `type` varchar(20) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `player`
--

CREATE TABLE `player` (
  `id` int(11) UNSIGNED NOT NULL,
  `name` varchar(50) DEFAULT NULL,
  `name_cn` varchar(50) DEFAULT NULL,
  `team_id` int(11) DEFAULT NULL,
  `position_id` int(11) DEFAULT NULL,
  `position_name` varchar(20) DEFAULT NULL,
  `position_name_long` varchar(20) DEFAULT NULL,
  `shirt_number` int(11) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `player_nickname`
--

CREATE TABLE `player_nickname` (
  `id` int(11) UNSIGNED NOT NULL,
  `name` varchar(50) DEFAULT NULL,
  `nickname` varchar(50) DEFAULT NULL,
  `type` varchar(20) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `queue_live`
--

CREATE TABLE `queue_live` (
  `id` int(11) UNSIGNED NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `queue_statistic`
--

CREATE TABLE `queue_statistic` (
  `id` int(11) UNSIGNED NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `statistic_lineup`
--

CREATE TABLE `statistic_lineup` (
  `id` int(11) UNSIGNED NOT NULL,
  `player_id` int(11) DEFAULT NULL,
  `match_id` int(11) DEFAULT NULL,
  `team_id` int(11) DEFAULT NULL,
  `shirt_number` int(11) DEFAULT NULL,
  `position_id` int(11) DEFAULT NULL,
  `position_name` varchar(20) DEFAULT NULL,
  `position_name_long` varchar(20) DEFAULT NULL,
  `player_name_cn` varchar(50) DEFAULT NULL,
  `pos_xy` int(11) DEFAULT NULL,
  `fetch_time` datetime DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `statistic_lineup_event`
--

CREATE TABLE `statistic_lineup_event` (
  `id` int(11) UNSIGNED NOT NULL,
  `match_id` int(11) DEFAULT NULL,
  `player_id` int(11) DEFAULT NULL,
  `time` int(11) DEFAULT NULL,
  `event_code` int(11) DEFAULT NULL,
  `event_name` varchar(20) DEFAULT NULL,
  `fetch_time` datetime DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `statistic_player`
--

CREATE TABLE `statistic_player` (
  `id` int(11) UNSIGNED NOT NULL,
  `player_id` int(11) DEFAULT NULL,
  `player_name` varchar(50) DEFAULT NULL,
  `player_name_cn` varchar(50) DEFAULT NULL,
  `match_id` int(11) DEFAULT NULL,
  `team_id` int(11) DEFAULT NULL,
  `league_id` int(11) DEFAULT NULL,
  `position_id` int(11) DEFAULT NULL,
  `position_name` varchar(20) DEFAULT NULL,
  `position_name_long` varchar(20) DEFAULT NULL,
  `shirt_number` int(11) DEFAULT NULL,
  `total_scoring_att` int(11) DEFAULT NULL,
  `ontarget_scoring_att` int(11) DEFAULT NULL,
  `ontarget_scoring_percent` float DEFAULT NULL,
  `goals` int(11) DEFAULT NULL,
  `goal_assist` int(11) DEFAULT NULL,
  `total_att_assist` int(11) DEFAULT NULL,
  `total_contest` int(11) DEFAULT NULL,
  `won_contest` int(11) DEFAULT NULL,
  `fouls` int(11) DEFAULT NULL,
  `was_fouled` int(11) DEFAULT NULL,
  `total_clearance` int(11) DEFAULT NULL,
  `saves` int(11) DEFAULT NULL,
  `won_tackle` int(11) DEFAULT NULL,
  `total_tackle` int(11) DEFAULT NULL,
  `accurate_pass` int(11) DEFAULT NULL,
  `total_pass` int(11) DEFAULT NULL,
  `mins_played` int(11) DEFAULT NULL,
  `status` varchar(20) DEFAULT NULL,
  `status_name` varchar(20) DEFAULT NULL,
  `fetch_time` datetime DEFAULT NULL,
  `live_id` text
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `statistic_team`
--

CREATE TABLE `statistic_team` (
  `id` int(11) UNSIGNED NOT NULL,
  `team_id` int(11) DEFAULT NULL,
  `team_name` varchar(50) DEFAULT NULL,
  `match_id` int(11) DEFAULT NULL,
  `type_id` int(11) DEFAULT NULL,
  `type_name` varchar(20) DEFAULT NULL,
  `event` varchar(50) DEFAULT NULL,
  `value` float DEFAULT NULL,
  `is_home` int(11) DEFAULT NULL,
  `fetch_time` datetime DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `team`
--

CREATE TABLE `team` (
  `id` int(11) UNSIGNED NOT NULL,
  `name` varchar(50) DEFAULT NULL,
  `type_id` int(11) DEFAULT NULL,
  `sl_id` int(11) DEFAULT NULL,
  `team_id` int(11) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 資料表結構 `team_nickname`
--

CREATE TABLE `team_nickname` (
  `id` int(11) UNSIGNED NOT NULL,
  `name` varchar(50) DEFAULT NULL,
  `nickname` varchar(50) DEFAULT NULL,
  `type` varchar(20) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

--
-- 已匯出資料表的索引
--

--
-- 資料表索引 `crawled_live`
--
ALTER TABLE `crawled_live`
  ADD PRIMARY KEY (`id`),
  ADD KEY `ix_soccer_crawled_live_87ea5dfc8b8e384d` (`id`);

--
-- 資料表索引 `crawled_statistic`
--
ALTER TABLE `crawled_statistic`
  ADD PRIMARY KEY (`id`),
  ADD KEY `ix_soccer_crawled_statistic_87ea5dfc8b8e384d` (`id`),
  ADD KEY `ix_crawled_statistic_87ea5dfc8b8e384d` (`id`);

--
-- 資料表索引 `json_preparation`
--
ALTER TABLE `json_preparation`
  ADD PRIMARY KEY (`match_id`),
  ADD KEY `ix_json_preparation_92b6c3a6631dd5b2` (`match_id`);

--
-- 資料表索引 `leagues_info`
--
ALTER TABLE `leagues_info`
  ADD PRIMARY KEY (`id`),
  ADD KEY `ix_soccer_leagues_info_87ea5dfc8b8e384d` (`id`),
  ADD KEY `sl_league_type` (`sl_league_type`),
  ADD KEY `ix_leagues_info_87ea5dfc8b8e384d` (`id`);

--
-- 資料表索引 `live_match_event`
--
ALTER TABLE `live_match_event`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `event_key` (`live_id`,`event_id`),
  ADD KEY `ix_soccer_live_match_event_fd6ed872f2e58c67` (`live_id`,`event_id`),
  ADD KEY `live_id` (`live_id`),
  ADD KEY `event_type` (`event_type`),
  ADD KEY `team_id` (`team_id`),
  ADD KEY `player_id` (`player_id`);

--
-- 資料表索引 `live_match_info`
--
ALTER TABLE `live_match_info`
  ADD PRIMARY KEY (`id`),
  ADD KEY `ix_soccer_live_match_info_87ea5dfc8b8e384d` (`id`),
  ADD KEY `league_type` (`league_type`),
  ADD KEY `season` (`season`),
  ADD KEY `round` (`round`),
  ADD KEY `team1_id` (`team1_id`),
  ADD KEY `team2_id` (`team2_id`),
  ADD KEY `time` (`time`);

--
-- 資料表索引 `live_teamformation`
--
ALTER TABLE `live_teamformation`
  ADD PRIMARY KEY (`id`),
  ADD KEY `ix_soccer_live_teamformation_87ea5dfc8b8e384d` (`id`),
  ADD KEY `live_id` (`live_id`),
  ADD KEY `player_id` (`player_id`),
  ADD KEY `sl_team_id` (`sl_team_id`),
  ADD KEY `match_id` (`match_id`);

--
-- 資料表索引 `match_info`
--
ALTER TABLE `match_info`
  ADD PRIMARY KEY (`id`),
  ADD KEY `ix_soccer_match_info_87ea5dfc8b8e384d` (`id`),
  ADD KEY `league_type` (`league_type`),
  ADD KEY `season` (`season`),
  ADD KEY `round` (`round`),
  ADD KEY `team1_id` (`team1_id`),
  ADD KEY `team2_id` (`team2_id`),
  ADD KEY `time` (`time`),
  ADD KEY `ix_match_info_87ea5dfc8b8e384d` (`id`);

--
-- 資料表索引 `match_nickname`
--
ALTER TABLE `match_nickname`
  ADD PRIMARY KEY (`id`),
  ADD KEY `team1_name` (`team1_name`),
  ADD KEY `team2_name` (`team2_name`);

--
-- 資料表索引 `player`
--
ALTER TABLE `player`
  ADD PRIMARY KEY (`id`),
  ADD KEY `team_id` (`team_id`),
  ADD KEY `ix_soccer_player_87ea5dfc8b8e384d` (`id`),
  ADD KEY `ix_player_87ea5dfc8b8e384d` (`id`);

--
-- 資料表索引 `player_nickname`
--
ALTER TABLE `player_nickname`
  ADD PRIMARY KEY (`id`),
  ADD KEY `name` (`name`);

--
-- 資料表索引 `queue_live`
--
ALTER TABLE `queue_live`
  ADD PRIMARY KEY (`id`);

--
-- 資料表索引 `queue_statistic`
--
ALTER TABLE `queue_statistic`
  ADD PRIMARY KEY (`id`);

--
-- 資料表索引 `statistic_lineup`
--
ALTER TABLE `statistic_lineup`
  ADD PRIMARY KEY (`id`),
  ADD KEY `ix_soccer_statistic_lineup_c6c60b1f6b941a03` (`player_id`,`team_id`,`match_id`),
  ADD KEY `ix_statistic_lineup_c6c60b1f6b941a03` (`player_id`,`team_id`,`match_id`);

--
-- 資料表索引 `statistic_lineup_event`
--
ALTER TABLE `statistic_lineup_event`
  ADD PRIMARY KEY (`id`),
  ADD KEY `ix_soccer_statistic_lineup_event_87ea5dfc8b8e384d` (`id`),
  ADD KEY `match_id` (`match_id`),
  ADD KEY `player_id` (`player_id`),
  ADD KEY `event_code` (`event_code`),
  ADD KEY `ix_statistic_lineup_event_87ea5dfc8b8e384d` (`id`);

--
-- 資料表索引 `statistic_player`
--
ALTER TABLE `statistic_player`
  ADD PRIMARY KEY (`id`),
  ADD KEY `ix_soccer_statistic_player_5dc5c18e00031f14` (`match_id`,`player_id`),
  ADD KEY `player_name` (`player_name`),
  ADD KEY `match_id` (`match_id`),
  ADD KEY `team_id` (`team_id`),
  ADD KEY `league_id` (`league_id`),
  ADD KEY `ix_statistic_player_5dc5c18e00031f14` (`match_id`,`player_id`);

--
-- 資料表索引 `statistic_team`
--
ALTER TABLE `statistic_team`
  ADD PRIMARY KEY (`id`),
  ADD KEY `ix_soccer_statistic_team_13923a1f1953b0cd` (`match_id`,`team_id`,`event`),
  ADD KEY `team_id` (`team_id`),
  ADD KEY `match_id` (`match_id`),
  ADD KEY `type_id` (`type_id`),
  ADD KEY `ix_statistic_team_13923a1f1953b0cd` (`match_id`,`team_id`,`event`);

--
-- 資料表索引 `team`
--
ALTER TABLE `team`
  ADD PRIMARY KEY (`id`),
  ADD KEY `ix_soccer_team_da802293d689521c` (`id`,`sl_id`,`type_id`),
  ADD KEY `ix_soccer_team_9eca5fafa20f9103` (`name`,`sl_id`,`type_id`),
  ADD KEY `ix_soccer_team_6ae999552a0d2dca` (`name`),
  ADD KEY `ix_team_6ae999552a0d2dca` (`name`);

--
-- 資料表索引 `team_nickname`
--
ALTER TABLE `team_nickname`
  ADD PRIMARY KEY (`id`),
  ADD KEY `name` (`name`);

--
-- 在匯出的資料表使用 AUTO_INCREMENT
--

--
-- 使用資料表 AUTO_INCREMENT `crawled_live`
--
ALTER TABLE `crawled_live`
  MODIFY `id` int(11) UNSIGNED NOT NULL AUTO_INCREMENT;
--
-- 使用資料表 AUTO_INCREMENT `crawled_statistic`
--
ALTER TABLE `crawled_statistic`
  MODIFY `id` int(11) UNSIGNED NOT NULL AUTO_INCREMENT;
--
-- 使用資料表 AUTO_INCREMENT `leagues_info`
--
ALTER TABLE `leagues_info`
  MODIFY `id` int(11) UNSIGNED NOT NULL AUTO_INCREMENT;
--
-- 使用資料表 AUTO_INCREMENT `live_match_event`
--
ALTER TABLE `live_match_event`
  MODIFY `id` int(11) UNSIGNED NOT NULL AUTO_INCREMENT;
--
-- 使用資料表 AUTO_INCREMENT `live_match_info`
--
ALTER TABLE `live_match_info`
  MODIFY `id` int(11) UNSIGNED NOT NULL AUTO_INCREMENT;
--
-- 使用資料表 AUTO_INCREMENT `live_teamformation`
--
ALTER TABLE `live_teamformation`
  MODIFY `id` int(11) UNSIGNED NOT NULL AUTO_INCREMENT;
--
-- 使用資料表 AUTO_INCREMENT `match_info`
--
ALTER TABLE `match_info`
  MODIFY `id` int(11) UNSIGNED NOT NULL AUTO_INCREMENT;
--
-- 使用資料表 AUTO_INCREMENT `match_nickname`
--
ALTER TABLE `match_nickname`
  MODIFY `id` int(11) UNSIGNED NOT NULL AUTO_INCREMENT;
--
-- 使用資料表 AUTO_INCREMENT `player`
--
ALTER TABLE `player`
  MODIFY `id` int(11) UNSIGNED NOT NULL AUTO_INCREMENT;
--
-- 使用資料表 AUTO_INCREMENT `player_nickname`
--
ALTER TABLE `player_nickname`
  MODIFY `id` int(11) UNSIGNED NOT NULL AUTO_INCREMENT;
--
-- 使用資料表 AUTO_INCREMENT `queue_live`
--
ALTER TABLE `queue_live`
  MODIFY `id` int(11) UNSIGNED NOT NULL AUTO_INCREMENT;
--
-- 使用資料表 AUTO_INCREMENT `queue_statistic`
--
ALTER TABLE `queue_statistic`
  MODIFY `id` int(11) UNSIGNED NOT NULL AUTO_INCREMENT;
--
-- 使用資料表 AUTO_INCREMENT `statistic_lineup`
--
ALTER TABLE `statistic_lineup`
  MODIFY `id` int(11) UNSIGNED NOT NULL AUTO_INCREMENT;
--
-- 使用資料表 AUTO_INCREMENT `statistic_lineup_event`
--
ALTER TABLE `statistic_lineup_event`
  MODIFY `id` int(11) UNSIGNED NOT NULL AUTO_INCREMENT;
--
-- 使用資料表 AUTO_INCREMENT `statistic_player`
--
ALTER TABLE `statistic_player`
  MODIFY `id` int(11) UNSIGNED NOT NULL AUTO_INCREMENT;
--
-- 使用資料表 AUTO_INCREMENT `statistic_team`
--
ALTER TABLE `statistic_team`
  MODIFY `id` int(11) UNSIGNED NOT NULL AUTO_INCREMENT;
--
-- 使用資料表 AUTO_INCREMENT `team`
--
ALTER TABLE `team`
  MODIFY `id` int(11) UNSIGNED NOT NULL AUTO_INCREMENT;
--
-- 使用資料表 AUTO_INCREMENT `team_nickname`
--
ALTER TABLE `team_nickname`
  MODIFY `id` int(11) UNSIGNED NOT NULL AUTO_INCREMENT;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
