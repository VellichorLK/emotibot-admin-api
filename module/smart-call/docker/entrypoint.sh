#!/bin/sh
while read line
do
  eval echo $line >> .env
done < $1


./admin-api .env