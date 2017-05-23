# Base image for java run environment
FROM maven:3-jdk-8-alpine

# mkdir
RUN mkdir -p /usr/src/app/rabbitmq

ADD rabbitmq/pom.xml /usr/src/app/rabbitmq/pom.xml


RUN cd /usr/src/app/rabbitmq && mvn dependency:go-offline dependency:copy-dependencies -DoutputDirectory=/usr/src/app

ADD . /usr/src/app

RUN cd /usr/src/app/rabbitmq && mvn package


#mvn jar to the same folder with nlu data
ENV RABBITMQ_VER 0.0.1-SNAPSHOT
RUN mv /usr/src/app/rabbitmq/target/rabbitmq-$RABBITMQ_VER.jar /usr/src/app

WORKDIR /usr/src/app

# Entrypoint that reads the shell ENV variables as config flags
ENTRYPOINT ["/usr/src/app/entrypoint.sh"]
