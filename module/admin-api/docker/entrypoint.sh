#!/bin/sh
while read line
do
  eval echo $line >> .env
done < $1

echo "0 * * * * sh `pwd`/profile_rebuild.sh" >> crontab.list
crontab crontab.list

./files_init.sh

./admin-api .env