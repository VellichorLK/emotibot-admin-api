#!/bin/sh
while read line
do
  eval echo $line >> .env
done < $1

#詞庫導入需要
./files_init.sh

./admin-api .env