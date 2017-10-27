#!/bin/sh
while read line
do
  eval echo $line >> .env
done < $1

./files_init.sh

./vipshop-admin .env