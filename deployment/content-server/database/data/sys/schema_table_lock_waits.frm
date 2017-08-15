TYPE=VIEW
query=select `g`.`OBJECT_SCHEMA` AS `object_schema`,`g`.`OBJECT_NAME` AS `object_name`,`pt`.`THREAD_ID` AS `waiting_thread_id`,`pt`.`PROCESSLIST_ID` AS `waiting_pid`,`sys`.`ps_thread_account`(`p`.`OWNER_THREAD_ID`) AS `waiting_account`,`p`.`LOCK_TYPE` AS `waiting_lock_type`,`p`.`LOCK_DURATION` AS `waiting_lock_duration`,`sys`.`format_statement`(`pt`.`PROCESSLIST_INFO`) AS `waiting_query`,`pt`.`PROCESSLIST_TIME` AS `waiting_query_secs`,`ps`.`ROWS_AFFECTED` AS `waiting_query_rows_affected`,`ps`.`ROWS_EXAMINED` AS `waiting_query_rows_examined`,`gt`.`THREAD_ID` AS `blocking_thread_id`,`gt`.`PROCESSLIST_ID` AS `blocking_pid`,`sys`.`ps_thread_account`(`g`.`OWNER_THREAD_ID`) AS `blocking_account`,`g`.`LOCK_TYPE` AS `blocking_lock_type`,`g`.`LOCK_DURATION` AS `blocking_lock_duration`,concat(\'KILL QUERY \',`gt`.`PROCESSLIST_ID`) AS `sql_kill_blocking_query`,concat(\'KILL \',`gt`.`PROCESSLIST_ID`) AS `sql_kill_blocking_connection` from (((((`performance_schema`.`metadata_locks` `g` join `performance_schema`.`metadata_locks` `p` on(((`g`.`OBJECT_TYPE` = `p`.`OBJECT_TYPE`) and (`g`.`OBJECT_SCHEMA` = `p`.`OBJECT_SCHEMA`) and (`g`.`OBJECT_NAME` = `p`.`OBJECT_NAME`) and (`g`.`LOCK_STATUS` = \'GRANTED\') and (`p`.`LOCK_STATUS` = \'PENDING\')))) join `performance_schema`.`threads` `gt` on((`g`.`OWNER_THREAD_ID` = `gt`.`THREAD_ID`))) join `performance_schema`.`threads` `pt` on((`p`.`OWNER_THREAD_ID` = `pt`.`THREAD_ID`))) left join `performance_schema`.`events_statements_current` `gs` on((`g`.`OWNER_THREAD_ID` = `gs`.`THREAD_ID`))) left join `performance_schema`.`events_statements_current` `ps` on((`p`.`OWNER_THREAD_ID` = `ps`.`THREAD_ID`))) where (`g`.`OBJECT_TYPE` = \'TABLE\')
md5=18ba2eedcb19b3b07c83b640d9960eff
updatable=0
algorithm=1
definer_user=mysql.sys
definer_host=localhost
suid=0
with_check_option=0
timestamp=2017-06-15 05:45:34
create-version=1
source=SELECT g.object_schema AS object_schema, g.object_name AS object_name, pt.thread_id AS waiting_thread_id, pt.processlist_id AS waiting_pid, sys.ps_thread_account(p.owner_thread_id) AS waiting_account, p.lock_type AS waiting_lock_type, p.lock_duration AS waiting_lock_duration, sys.format_statement(pt.processlist_info) AS waiting_query, pt.processlist_time AS waiting_query_secs, ps.rows_affected AS waiting_query_rows_affected, ps.rows_examined AS waiting_query_rows_examined, gt.thread_id AS blocking_thread_id, gt.processlist_id AS blocking_pid, sys.ps_thread_account(g.owner_thread_id) AS blocking_account, g.lock_type AS blocking_lock_type, g.lock_duration AS blocking_lock_duration, CONCAT(\'KILL QUERY \', gt.processlist_id) AS sql_kill_blocking_query, CONCAT(\'KILL \', gt.processlist_id) AS sql_kill_blocking_connection FROM performance_schema.metadata_locks g INNER JOIN performance_schema.metadata_locks p  ON g.object_type = p.object_type AND g.object_schema = p.object_schema AND g.object_name = p.object_name AND g.lock_status = \'GRANTED\' AND p.lock_status = \'PENDING\' INNER JOIN performance_schema.threads gt ON g.owner_thread_id = gt.thread_id INNER JOIN performance_schema.threads pt ON p.owner_thread_id = pt.thread_id LEFT JOIN performance_schema.events_statements_current gs ON g.owner_thread_id = gs.thread_id LEFT JOIN performance_schema.events_statements_current ps ON p.owner_thread_id = ps.thread_id WHERE g.object_type = \'TABLE\'
client_cs_name=utf8
connection_cl_name=utf8_general_ci
view_body_utf8=select `g`.`OBJECT_SCHEMA` AS `object_schema`,`g`.`OBJECT_NAME` AS `object_name`,`pt`.`THREAD_ID` AS `waiting_thread_id`,`pt`.`PROCESSLIST_ID` AS `waiting_pid`,`sys`.`ps_thread_account`(`p`.`OWNER_THREAD_ID`) AS `waiting_account`,`p`.`LOCK_TYPE` AS `waiting_lock_type`,`p`.`LOCK_DURATION` AS `waiting_lock_duration`,`sys`.`format_statement`(`pt`.`PROCESSLIST_INFO`) AS `waiting_query`,`pt`.`PROCESSLIST_TIME` AS `waiting_query_secs`,`ps`.`ROWS_AFFECTED` AS `waiting_query_rows_affected`,`ps`.`ROWS_EXAMINED` AS `waiting_query_rows_examined`,`gt`.`THREAD_ID` AS `blocking_thread_id`,`gt`.`PROCESSLIST_ID` AS `blocking_pid`,`sys`.`ps_thread_account`(`g`.`OWNER_THREAD_ID`) AS `blocking_account`,`g`.`LOCK_TYPE` AS `blocking_lock_type`,`g`.`LOCK_DURATION` AS `blocking_lock_duration`,concat(\'KILL QUERY \',`gt`.`PROCESSLIST_ID`) AS `sql_kill_blocking_query`,concat(\'KILL \',`gt`.`PROCESSLIST_ID`) AS `sql_kill_blocking_connection` from (((((`performance_schema`.`metadata_locks` `g` join `performance_schema`.`metadata_locks` `p` on(((`g`.`OBJECT_TYPE` = `p`.`OBJECT_TYPE`) and (`g`.`OBJECT_SCHEMA` = `p`.`OBJECT_SCHEMA`) and (`g`.`OBJECT_NAME` = `p`.`OBJECT_NAME`) and (`g`.`LOCK_STATUS` = \'GRANTED\') and (`p`.`LOCK_STATUS` = \'PENDING\')))) join `performance_schema`.`threads` `gt` on((`g`.`OWNER_THREAD_ID` = `gt`.`THREAD_ID`))) join `performance_schema`.`threads` `pt` on((`p`.`OWNER_THREAD_ID` = `pt`.`THREAD_ID`))) left join `performance_schema`.`events_statements_current` `gs` on((`g`.`OWNER_THREAD_ID` = `gs`.`THREAD_ID`))) left join `performance_schema`.`events_statements_current` `ps` on((`p`.`OWNER_THREAD_ID` = `ps`.`THREAD_ID`))) where (`g`.`OBJECT_TYPE` = \'TABLE\')
