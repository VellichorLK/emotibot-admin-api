FROM logstash:5

RUN apt-get update && apt-get install -y vim

RUN logstash-plugin install logstash-output-jdbc

COPY config-dir/* /config-dir/

# CMD ["-f", "/config-dir/logstash.conf"]
