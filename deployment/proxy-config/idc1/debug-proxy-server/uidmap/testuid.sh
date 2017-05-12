#!/bin/bash

for i in $(cat uid.txt)
do
  echo -n $i " "
  curl "localhost:9000/?userid=$i" 2>/dev/null | cut -d" " -f5
done

