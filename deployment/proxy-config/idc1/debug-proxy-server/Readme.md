# Debugging Server
The debug server provides a way to find out the uid mapping.
To run the server:
```
# Note: prepare the docker images, you just need to run this script ONCE.
./build.sh

# Start all servers
./run.sh

# Test the server
curl localhost:9000?userid=5566

# Kill all servers
./kill.sh
```
