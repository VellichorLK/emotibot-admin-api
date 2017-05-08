CNAME=logging-mongo
DOCKER_IMG=mongo:3.2.8

MONGO_DATA_PATH=$1
HOST_PORT=$2

if [ ! -d "${MONGO_DATA_PATH}" ]; then
	echo "$MONGO_DATA_PATH does not exist"
	echo "please make sure the path is created and mounted to mongo container"
	exit 1
fi

# set users
cmd="
docker exec -it $CNAME mongo /tmp/init_users.js
"
eval $cmd


# restart mongo with authentication enable
docker rm -f -v $CNAME

cmd="\
docker run -d \
  --restart="always" \
  --name $CNAME \
  -p $HOST_PORT:27017 \
  -v $MONGO_DATA_PATH/db:/data/db \
  -v $MONGO_DATA_PATH/configdb:/data/configdb \
  $DOCKER_IMG \
  mongod --auth
"

echo $cmd
eval $cmd
