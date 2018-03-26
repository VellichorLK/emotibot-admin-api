#!/bin/sh
url="$VIP_SERVER_MC_URL/manual_edit?app_id=vipshop&type=robot";
return=`curl "$url"`;
date=`date`;
echo "[CRON][$date] Rebuild robot profile: $return"