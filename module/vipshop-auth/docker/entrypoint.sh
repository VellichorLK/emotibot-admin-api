#!/bin/sh
while read line
do
  eval echo $line >> .env
done < $1

./vipshop-auth .env