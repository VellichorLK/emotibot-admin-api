#!/bin/sh
DB=`printenv DB_HOST`
USER=`printenv DB_USER`
PWD=`printenv DB_PWD`
DBNAME=`printenv DB_NAME`

/usr/local/bin/transfer -p /usr/local/bin/tmp -db "$DB" -u "$USER" -pass "$PWD" -dbname "$DBNAME"