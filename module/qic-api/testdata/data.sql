LOCK TABLES `Group` WRITE;
INSERT INTO `Group` (`app_id`, `is_delete`, `group_name`, `enterprise`, `description`, `create_time`, `update_time`, `is_enable`, `limit_speed`, `limit_silence`, `type`)
VALUES
	(1,0,'testing','123456789','this is an integration test data',0,0,1,0,0,0),
	(2,0,'testing2','123456789','this is another integration test data',0,0,1,0,0,1);

UNLOCK TABLES;
