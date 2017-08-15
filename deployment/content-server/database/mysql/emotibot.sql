CREATE DATABASE IF NOT EXISTS `emotibot` DEFAULT CHARACTER SET utf8 COLLATE utf8_general_ci;
USE `emotibot`;

DROP TABLE IF EXISTS `vipshop_module_privilege`;
CREATE TABLE `vipshop_module_privilege` (
  `MPId` int(11) NOT NULL AUTO_INCREMENT PRIMARY KEY,
  `ModuleCode` varchar(50) NOT NULL,
  `PriCode` varchar(50) NOT NULL,
  `CreatedUserId` varchar(50) NOT NULL,
  `CreatedTime` datetime NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

ALTER TABLE `vipshop_module_privilege` ADD KEY `vipshop_module_privilege_ModuleCode_IDX` (`ModuleCode`,`PriCode`);

INSERT INTO `vipshop_module_privilege` (`ModuleCode`, `PriCode`, `CreatedUserId`, `CreatedTime`) VALUES
('dailyReport', 'view', '01C2DB528B60E5A498781452FCB509E6C', '2017-04-12 11:17:28'),
('statsAnalysis', 'view', '01C2DB528B60E5A498781452FCB509E6C', '2017-04-12 11:17:28'),
('qna', 'view', '01C2DB528B60E5A498781452FCB509E6C', '2017-04-12 11:17:28'),
('qna', 'edit', '01C2DB528B60E5A498781452FCB509E6C', '2017-04-12 11:17:28'),
('qna', 'new', '01C2DB528B60E5A498781452FCB509E6C', '2017-04-12 11:17:28'),
('qna', 'delete', '01C2DB528B60E5A498781452FCB509E6C', '2017-04-12 11:17:28'),
('qna', 'import', '01C2DB528B60E5A498781452FCB509E6C', '2017-04-12 11:17:28'),
('qna', 'export', '01C2DB528B60E5A498781452FCB509E6C', '2017-04-12 11:17:28'),
('chatVIP', 'view', '01C2DB528B60E5A498781452FCB509E6C', '2017-04-12 11:17:28'),
('wordbank', 'view', '01C2DB528B60E5A498781452FCB509E6C', '2017-04-12 11:17:28'),
('wordbank', 'import', '01C2DB528B60E5A498781452FCB509E6C', '2017-04-12 11:17:28'),
('wordbank', 'export', '01C2DB528B60E5A498781452FCB509E6C', '2017-04-12 11:17:28'),
('robotprofile', 'view', '01C2DB528B60E5A498781452FCB509E6C', '2017-04-12 11:17:28'),
('robotprofile', 'edit', '01C2DB528B60E5A498781452FCB509E6C', '2017-04-12 11:17:28'),
('functionswitch', 'view', '01C2DB528B60E5A498781452FCB509E6C', '2017-04-12 11:17:28'),
('functionswitch', 'edit', '01C2DB528B60E5A498781452FCB509E6C', '2017-04-12 11:17:28'),
('botmessage', 'view', '01C2DB528B60E5A498781452FCB509E6C', '2017-04-12 11:17:28'),
('botmessage', 'edit', '01C2DB528B60E5A498781452FCB509E6C', '2017-04-12 11:17:28'),
('switchList', 'view', '01C2DB528B60E5A498781452FCB509E6C', '2017-04-12 11:17:28'),
('switchList', 'edit', '01C2DB528B60E5A498781452FCB509E6C', '2017-04-12 11:17:28');

DROP TABLE IF EXISTS `privilege_setting`;
CREATE TABLE `privilege_setting` (
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

INSERT INTO `privilege_setting` (`PriId`, `PriCode`, `PriName`, `CreatedUserId`, `CreatedTime`, `Status`, `EditUserId`, `UpdatedTime`) VALUES
(3, 'view', '查看', '01C2DB528B60E5A498781452FCB509E6C', '2017-04-12 10:41:01', 0, '01C2DB528B60E5A498781452FCB509E6C', '2017-04-12 10:41:01'),
(4, 'edit', '修改', '01C2DB528B60E5A498781452FCB509E6C', '2017-04-12 10:41:01', 0, '01C2DB528B60E5A498781452FCB509E6C', '2017-04-12 10:41:01'),
(5, 'delete', '删除', '01C2DB528B60E5A498781452FCB509E6C', '2017-04-12 10:41:01', 0, '01C2DB528B60E5A498781452FCB509E6C', '2017-04-12 10:41:01'),
(6, 'new', '新增', '01C2DB528B60E5A498781452FCB509E6C', '2017-04-12 10:41:01', 0, '01C2DB528B60E5A498781452FCB509E6C', '2017-04-12 10:41:01'),
(7, 'import', '导入', '01C2DB528B60E5A498781452FCB509E6C', '2017-04-12 10:41:01', 0, '01C2DB528B60E5A498781452FCB509E6C', '2017-04-12 10:41:01'),
(8, 'export', '导出', '01C2DB528B60E5A498781452FCB509E6C', '2017-04-12 10:41:01', 0, '01C2DB528B60E5A498781452FCB509E6C', '2017-04-12 10:41:01'),
(9, 'rollback', '恢复', '01C2DB528B60E5A498781452FCB509E6C', '2017-04-12 10:41:01', 0, '01C2DB528B60E5A498781452FCB509E6C', '2017-04-12 10:41:01');

DROP TABLE IF EXISTS `role_privilege`;
CREATE TABLE `role_privilege` (
  `EnterpriseId` varchar(50) NOT NULL,
  `RoleCode` varchar(50) NOT NULL,
  `MPId` int(11) NOT NULL,
  `CreatedUserId` varchar(50) NOT NULL,
  `CreatedTime` datetime NOT NULL,
  PRIMARY KEY (`EnterpriseId`,`RoleCode`,`MPId`),
  KEY `role_privilege_RoleCode_IDX` (`EnterpriseId`,`RoleCode`,`MPId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

DROP TABLE IF EXISTS `user_role`;
CREATE TABLE `user_role` (
  `UserId` varchar(50) NOT NULL,
  `RoleCode` varchar(50) NOT NULL,
  `EnterpriseId` varchar(50) NOT NULL,
  `CreatedTime` datetime NOT NULL,
  PRIMARY KEY (`UserId`,`RoleCode`,`EnterpriseId`),
  KEY `user_role_UserId_IDX` (`UserId`,`RoleCode`,`EnterpriseId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

DROP TABLE IF EXISTS `role`;
CREATE TABLE `role` (
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

DROP TABLE IF EXISTS `vipshop_module`;
CREATE TABLE `vipshop_module` (
  `ModuleId` int(11) NOT NULL AUTO_INCREMENT PRIMARY KEY,
  `ModuleCode` varchar(50) NOT NULL,
  `ModuleName` varchar(100) NOT NULL,
  `ParentCode` varchar(50) NOT NULL,
  `ModuleUrl` varchar(500) NOT NULL,
  `CreatedUserId` varchar(50) NOT NULL,
  `CreatedTime` datetime NOT NULL,
  `Status` int(11) DEFAULT '0' COMMENT '-1:删除; 0:停止; 1:启动',
  `EditUserId` varchar(50) NOT NULL,
  `UpdatedTime` datetime NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

ALTER TABLE `vipshop_module` ADD KEY `vipshop_module_ModuleCode_IDX` (`ModuleCode`,`ModuleName`,`ParentCode`,`Status`);

INSERT INTO `vipshop_module` (`ModuleCode`, `ModuleName`, `ParentCode`, `ModuleUrl`, `CreatedUserId`, `CreatedTime`, `Status`, `EditUserId`, `UpdatedTime`) VALUES
('dailyReport', '日志管理', '0', '', '01C2DB528B60E5A498781452FCB509E6C', '2017-04-12 10:38:00', 1, '01C2DB528B60E5A498781452FCB509E6C', '2017-04-12 10:38:00'),
('statsAnalysis', '统计分析管理', '0', '', '01C2DB528B60E5A498781452FCB509E6C', '2017-04-12 10:38:00', 1, '01C2DB528B60E5A498781452FCB509E6C', '2017-04-12 10:38:00'),
('qna', '问答库', '0', '', '01C2DB528B60E5A498781452FCB509E6C', '2017-04-12 10:38:00', 1, '01C2DB528B60E5A498781452FCB509E6C', '2017-04-12 10:38:00'),
('chatVIP', '对话测试', '0', '', '01C2DB528B60E5A498781452FCB509E6C', '2017-04-12 10:38:00', 1, '01C2DB528B60E5A498781452FCB509E6C', '2017-04-12 10:38:00'),
('wordbank', '词库管理', '0', '', '01C2DB528B60E5A498781452FCB509E6C', '2017-04-12 10:38:00', 1, '01C2DB528B60E5A498781452FCB509E6C', '2017-04-12 10:38:00'),
('robotprofile', '形象设置', '0', '', '01C2DB528B60E5A498781452FCB509E6C', '2017-04-12 10:38:00', 1, '01C2DB528B60E5A498781452FCB509E6C', '2017-04-12 10:38:00'),
('functionswitch', '技能设置', '0', '', '01C2DB528B60E5A498781452FCB509E6C', '2017-04-12 10:38:00', 1, '01C2DB528B60E5A498781452FCB509E6C', '2017-04-12 10:38:00'),
('switchList', '开关管理', '0', '', '01C2DB528B60E5A498781452FCB509E6C', '2017-04-12 10:38:00', 1, '01C2DB528B60E5A498781452FCB509E6C', '2017-04-12 10:38:00'),
('botmessage', '话术设置', '0', '', '01C2DB528B60E5A498781452FCB509E6C', '2017-04-12 10:38:00', 1, '01C2DB528B60E5A498781452FCB509E6C', '2017-04-12 10:38:00');

/*
    BF用户注册表。
*/

DROP TABLE IF EXISTS `api_user`;
CREATE TABLE `api_user` (
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

INSERT INTO api_user(UserId, Phone, Email, CreatedTime, Password, NickName, Gender, Type, Status, UpdatedTime, Remark, Owner)
  VALUES('012CB30AC63689B9E59FAA54E5F9C9D29', '1350000000', 'vip', now(), '24c1c867f8772d2fae7ff572013de0bc', '唯品会', 1, 1, 1, now(), ' ', 'vip');

/*
    BF 企业账号注册表。
*/
DROP TABLE IF EXISTS `enterprise`;
CREATE TABLE `enterprise` (
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

INSERT INTO enterprise(EnterpriseId, Account, CreatedTime, Password, NickName, Location, PeopleNumber, Industry, LinkName, LinkPhone, LinkEmail, Status, UpdatedTime, Remark, UserId)
  VALUES('0D01010E11C0BF5277A705DE36AEE3FBB', 'vipadmin', now(), '24c1c867f8772d2fae7ff572013de0bc', 'vipadmin', '', 1, 'shop', 'shop', '', '', 1, now(), '', '012CB30AC63689B9E59FAA54E5F9C9D29');

/*
    BF 企业账号和用户的关系表。
*/
DROP TABLE IF EXISTS `enterprise_user`;
CREATE TABLE `enterprise_user` (
  `EnterpriseId` varchar(50) NOT NULL,
  `UserId` varchar(50) NOT NULL,
  `CreatedTime` datetime NOT NULL,
  `Status` int(11) DEFAULT '0' COMMENT '-1:解除绑定; 0:绑定要请; 1:管理员(admin); 2:普通用户',
  `UpdatedTime` datetime NOT NULL,
  `EnterpriseRemark` varchar(1000) DEFAULT NULL,
  `UserRemark` varchar(1000) DEFAULT NULL,
  PRIMARY KEY (`EnterpriseId`,`UserId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

INSERT INTO enterprise_user(EnterpriseId, UserId, CreatedTime, Status, UpdatedTime, EnterpriseRemark, UserRemark)
  VALUES('0D01010E11C0BF5277A705DE36AEE3FBB', '012CB30AC63689B9E59FAA54E5F9C9D29', now(), 1, now(), '', '');


DROP TABLE IF EXISTS `api_preduct`;
CREATE TABLE `api_preduct` (
  `PreductId` int(11) NOT NULL AUTO_INCREMENT,
  `PreductName` varchar(200) DEFAULT NULL,
  `PreductRemark` varchar(1000) DEFAULT NULL,
  `CreatedUser` int(11) NOT NULL,
  `CreatedTime` datetime NOT NULL,
  `PreductVersion` int(11) NOT NULL,
  `Status` int(11) NOT NULL,
  PRIMARY KEY (`PreductId`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8;
INSERT INTO api_preduct(PreductName, PreductRemark, CreatedUser, CreatedTime, PreductVersion, Status)
  VALUES('vip_robot', 'vip_robot', 1, now(), 1, 1);

/*
    appid 创建查询表。
*/
DROP TABLE IF EXISTS `api_userkey`;
CREATE TABLE `api_userkey` (
  `UserId` varchar(50) NOT NULL,
  `Count` int(11) NOT NULL,
  `Version` int(11) NOT NULL,
  `CreatedTime` datetime NOT NULL,
  `PreductName` varchar(255) NOT NULL,
  `ApiKey` varchar(50) NOT NULL,
  `Status` int(11) NOT NULL,
  `MaxCount` int(11) DEFAULT '500',
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

INSERT INTO api_userkey(UserId, Count, Version, CreatedTime, PreductName, ApiKey, Status, NickName, CommonFunctionIds, AreaIds, Type)
  VALUES('0D01010E11C0BF5277A705DE36AEE3FBB', 99999999, 1, now(), 'vip_robot', 'vipshop', 1, 'vipshop', ',7,', '1,2,3,4', 2);


/*
  标准答案表
*/
DROP TABLE IF EXISTS `vipshop_answer`;
CREATE TABLE `vipshop_answer` (
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


/*
  标准答案与维度信息的mapping表
*/
DROP TABLE IF EXISTS `vipshop_answertag`;
CREATE TABLE `vipshop_answertag` (
  `Answer_Id` int(11) NOT NULL,
  `Tag_Id` int(11) NOT NULL,
  `CreatedTime` datetime DEFAULT NULL,
  `Status` int(11) DEFAULT '1',
  PRIMARY KEY (`Answer_Id`,`Tag_Id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;


/*
  分类信息表
*/
DROP TABLE IF EXISTS `vipshop_categories`;
CREATE TABLE `vipshop_categories` (
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

INSERT INTO `vipshop_categories`(`CategoryId`, `CategoryName`, `ParentId`, `Status`, `level`, `ParentPath`, `SelfPath`)
  VALUES(-1, '暂无分类', 0, 1, 1, '/', '/暂无分类/');


/*
  指定动态菜单表
*/
DROP TABLE IF EXISTS `vipshop_dynamic_menu`;
CREATE TABLE `vipshop_dynamic_menu` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `Answer_id` int(11) NOT NULL,
  `DynamicMenu` varchar(200) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


DROP TABLE IF EXISTS `vipshop_onoff`;
CREATE TABLE `vipshop_onoff` (
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


/*
  标准问题表
*/
DROP TABLE IF EXISTS `vipshop_question`;
CREATE TABLE `vipshop_question` (
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


/*
  指定相关问表
*/
DROP TABLE IF EXISTS `vipshop_related_question`;
CREATE TABLE `vipshop_related_question` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `Answer_id` int(11) NOT NULL,
  `RelatedQuestion` varchar(200) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


/*
  相似问题表
*/
DROP TABLE IF EXISTS `vipshop_squestion`;
CREATE TABLE `vipshop_squestion` (
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



/*
  答案的维度信息表
*/
DROP TABLE IF EXISTS `vipshop_tag`;
CREATE TABLE `vipshop_tag` (
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


INSERT INTO `vipshop_tag` (`Tag_Id`, `Tag_Name`, `CreatedUser`, `CreatedTime`, `EditUser`, `EditTime`, `Status`, `Tag_Type`) VALUES
(1, '#weixin#', NULL, NULL, NULL, NULL, 1, 1),
(2, '#app#', NULL, NULL, NULL, NULL, 1, 1),
(3, '#web#', NULL, NULL, NULL, NULL, 1, 1),
(4, '#特卖会APP#', NULL, NULL, NULL, NULL, 1, 2),
(5, '#PC端#', NULL, NULL, NULL, NULL, 1, 2),
(7, '#女#', NULL, NULL, NULL, NULL, 1, 3),
(8, '#男#', NULL, NULL, NULL, NULL, 1, 3),
(9, '#70s#', NULL, NULL, NULL, NULL, 1, 4),
(10, '#80s#', NULL, NULL, NULL, NULL, 1, 4),
(11, '#85s#', NULL, NULL, NULL, NULL, 1, 4),
(12, '#90s#', NULL, NULL, NULL, NULL, 1, 4),
(13, '#准新客#', NULL, NULL, NULL, NULL, 1, 5),
(14, '#非准新客#', NULL, NULL, NULL, NULL, 1, 5),
(17, '#WAP端#', NULL, NULL, NULL, NULL, 1, 2),
(18, '#微信公众号#', NULL, NULL, NULL, NULL, 1, 2),
(19, '#QQ公众号#', NULL, NULL, NULL, NULL, 1, 2),
(20, '#乐蜂APP#', NULL, NULL, NULL, NULL, 1, 2),
(23, '#花海仓#', NULL, NULL, NULL, NULL, 1, 2),
(24, '#特卖会app准新客#', NULL, NULL, NULL, NULL, 1, 2),
(25, '#ios#', NULL, NULL, NULL, NULL, 1, 1),
(27, '#母婴APP#', NULL, NULL, NULL, NULL, 1, 2);

/*
    用于对唯品会-问答定制：维度所属类别。
*/
DROP TABLE IF EXISTS `vipshop_tag_type`;
CREATE TABLE `vipshop_tag_type` (
  `Type_id` int(4) NOT NULL AUTO_INCREMENT,
  `Type_name` varchar(100) COLLATE utf8_unicode_ci NOT NULL,
  PRIMARY KEY (`Type_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

INSERT INTO `vipshop_tag_type` (`Type_id`, `Type_name`) VALUES
(1, '平台'),
(2, '品牌'),
(3, '性别'),
(4, '年龄段'),
(5, '购买爱好');


/*
  记录用户行为及状态的表： 包括问答定制，下载，词库上传。 初始化状态为running，失败为fail，成功为success；
  如果running时，docker被重启， 会自动重新处理该操作
*/
DROP TABLE IF EXISTS `process_status`;
CREATE TABLE IF NOT EXISTS `process_status` (
`id` int(18) NOT NULL AUTO_INCREMENT,
`app_id` varchar(50) NOT NULL,
`module` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
`start_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
`end_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
`status` varchar(10) NOT NULL,
`message` text COLLATE utf8mb4_unicode_ci,
`entity_file_name` varchar(255),
PRIMARY KEY (`id`),
KEY `IDX_app_id` (`app_id`),
KEY `IDX_app_module` (`app_id`, `module`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


/*
  话术用到的表
*/
DROP TABLE IF EXISTS `vipshop_robot_setting`;
CREATE TABLE `vipshop_robot_setting` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `content` text COLLATE utf8mb4_unicode_ci,
  `type` int(4) DEFAULT '1',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;




/*
  如果没有建表权限， 以下两个表也需要厨初始化
*/
CREATE TABLE IF NOT EXISTS `vipshop_robotquestion` (
`q_id` int(10) NOT NULL AUTO_INCREMENT,
`content` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
`user` varchar(255) CHARACTER SET utf8 COLLATE utf8_unicode_ci NOT NULL DEFAULT 'vipshop',
`created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
`answer_count` smallint(5) DEFAULT 0,
`content2` varchar(255) COLLATE utf8mb4_unicode_ci,
`content3` varchar(255) COLLATE utf8mb4_unicode_ci,
`content4` varchar(255) COLLATE utf8mb4_unicode_ci,
`content5` varchar(255) COLLATE utf8mb4_unicode_ci,
`content6` varchar(255) COLLATE utf8mb4_unicode_ci,
`content7` varchar(255) COLLATE utf8mb4_unicode_ci,
`content8` varchar(255) COLLATE utf8mb4_unicode_ci,
`content9` varchar(255) COLLATE utf8mb4_unicode_ci,
`content10` varchar(255) COLLATE utf8mb4_unicode_ci,
`status` int(2) DEFAULT 0,
`rownum` int(15) DEFAULT 0,
PRIMARY KEY (`q_id`),
KEY `content` (`content`),
KEY `IDX_q_id` (`q_id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;



CREATE TABLE IF NOT EXISTS `vipshop_robotanswer` (
`a_id` int(4) NOT NULL AUTO_INCREMENT,
`parent_q_id` int(10) NOT NULL,
`content` text COLLATE utf8mb4_unicode_ci  NOT NULL,
`user` varchar(255) CHARACTER SET utf8 COLLATE utf8_unicode_ci NOT NULL DEFAULT 'vipshop',
`created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
PRIMARY KEY (`a_id`),
KEY `IDX_a_id` (`a_id`),
KEY `answer_parent_q_id` (`parent_q_id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

--
-- Table structure for table `wechatusers`
--

DROP TABLE IF EXISTS `wechatusers`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `wechatusers` (
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
/*!40101 SET character_set_client = @saved_cs_client */;
