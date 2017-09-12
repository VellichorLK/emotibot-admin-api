-- --------------------------------------------------------

--
-- Table structure for table `ecovacs_module`
--
CREATE DATABASE IF NOT EXISTS `authentication` DEFAULT CHARACTER SET utf8 COLLATE utf8_general_ci;
USE `authentication`;

CREATE TABLE IF NOT EXISTS `ecovacs_module` (
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
  KEY `ecovacs_module_ModuleCode_IDX` (`ModuleCode`,`ModuleName`,`ParentCode`,`Status`)
) ENGINE=InnoDB AUTO_INCREMENT=10 DEFAULT CHARSET=utf8;

--
-- Dumping data for table `ecovacs_module`
--

INSERT INTO `ecovacs_module` (`ModuleId`, `ModuleCode`, `ModuleName`, `ParentCode`, `ModuleUrl`, `CreatedUserId`, `CreatedTime`, `Status`, `EditUserId`, `UpdatedTime`) VALUES
(1, 'voiceCheck', '音頻檢查', '0', '', '01C2DB528B60E5A498781452FCB509E6C', '2017-04-12 10:38:00', 1, '01C2DB528B60E5A498781452FCB509E6C', '2017-04-12 10:38:00'),
(2, 'voiceQueue', '分析隊列', '0', '', '01C2DB528B60E5A498781452FCB509E6C', '2017-04-12 10:38:00', 1, '01C2DB528B60E5A498781452FCB509E6C', '2017-04-12 10:38:00'),
(3, 'voiceReport', '統計報表', '0', '', '01C2DB528B60E5A498781452FCB509E6C', '2017-04-12 10:38:00', 1, '01C2DB528B60E5A498781452FCB509E6C', '2017-04-12 10:38:00'),
(4, 'systemConfig', '系統設置', '0', '', '01C2DB528B60E5A498781452FCB509E6C', '2017-04-12 10:38:00', 1, '01C2DB528B60E5A498781452FCB509E6C', '2017-04-12 10:38:00');

-- --------------------------------------------------------
