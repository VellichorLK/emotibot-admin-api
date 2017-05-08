# Development environment for golang
FROM golang:1.8
MAINTAINER WeiMin Yao <wmyaoyao@gmail.com>

# Install protobuf v3.2.0
RUN apt-get update && \
    apt-get install -y zip
RUN mkdir -p /tmp/proto && \
    cd /tmp/proto && \
    wget https://github.com/google/protobuf/releases/download/v3.2.0/protoc-3.2.0-linux-x86_64.zip && \
    unzip protoc-3.2.0-linux-x86_64.zip && \
    mv ./bin/protoc /usr/local/go/bin && \
    rm -rf /tmp/proto

# We have go and protobuf installed.
# The following steps are used to add another user to the docker so that we
# won't mess up with the file permissions mounted into the docker.

ENV PATH /usr/local/bin:/usr/local/sbin:/usr/sbin:/usr/bin:/sbin:/bin:$PATH
ENV PATH /opt/caffe/.build_release/tools:$PATH
ENV GOPATH /go
# Rename the variables below if you want
ENV PROJROOT /go/src/emotibot.com/emotigo
ENV DEVUSER devuser
ENV DEVHOME /home/$DEVUSER

# new user as sudoers without password
# for mac user boot2docker vm "docker" user uid is 1000
# for linux user, you must change 1000 to your own uid
RUN \
  useradd -d $DEVHOME -p $(openssl passwd -crypt julie520) -u 1000 -m -s /bin/bash $DEVUSER && \
  usermod -a -G sudo $DEVUSER && \
  chown -R $DEVUSER $DEVHOME && \
  echo "$DEVUSER ALL=NOPASSWD: ALL" >> /etc/sudoers

# add build-in scripts
ADD ./script /script

# Generate a /etc/profile.d so that env variables are still there when
# a new bash starts.
# Also force to switch to the WORKDIR
RUN \
  echo "#!/bin/bash" >> /etc/profile.d/profile.sh && \
  echo "sleep 1" >> /etc/profile.d/profile.sh && \
  echo "echo 'Initialize the dev environment.'" >> /etc/profile.d/profile.sh && \
  echo "export GOPATH=$GOPATH" >> /etc/profile.d/profile.sh && \
  echo "export PATH=$PATH" >> /etc/profile.d/profile.sh && \
  echo "cd $PROJROOT" >> /etc/profile.d/profile.sh

RUN \
 chmod 755 /script/*

# Define working directory. Actually, this won't work since we'll start a new shell
# for $DEVUSER. The cd is done/hardcoded in profile.sh instand.
WORKDIR $PROJROOT

# entrypoint
ENTRYPOINT ["/script/bootstrap.sh"]
