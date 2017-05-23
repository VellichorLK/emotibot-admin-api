#!/usr/bin/python
# -*- coding:utf-8 -*-

import pika
import uuid
import time
import sys
import threading
import traceback
import os
import json

ll = True


class PikaConection():
    def __init__(self, host, port):
        self.host = host
        self.port = port
        self.con = None
        self.lock = threading.Lock()
        self.connect()

    def connect(self):

        while self.con is None:
            try:
                self.con = pika.BlockingConnection(pika.ConnectionParameters(host=self.host, port=self.port))
            except:
                print "Trying to connect to rabbitMQ " + self.host + ":" + str(self.port) + " failed"
                time.sleep(3)

    def close(self):
        self.con.close()

    def on_disconnect(self):
        get_lock = self.lock.acquire(False)
        if get_lock:
            if self.test_connection():
                self.con = None
                self.connect()

            self.lock.release()

    def process_event(self, timeout):
        self.con.process_data_events(time_limit=timeout)

    def get_channel(self):
        return self.con.channel()

    def test_connection(self):
        res = True
        try:
            ch = self.con.channel()
            ch.close()
            res = False
        except:
            traceback.print_exc()
            pass
        return res

class PikaServer():

    def __init__(self, con, queue_name, do_func):
        self.con = con
        self.queue_name = queue_name
        self.do_func = do_func

    def init_channel(self):
        init_succ = True
        try:
            self.ch = self.con.get_channel()
            #create input queue
            result = self.ch.queue_declare(queue=self.queue_name)

            self.ch.basic_qos(prefetch_count=1)
            self.ch.basic_consume(self.on_request, queue=self.queue_name)
        except:
            print "pika server init failed!"
            traceback.print_exc()
            init_succ = False
        return init_succ

    def start_consume(self):

        while True:
            try:
                self.ch.start_consuming()
            except KeyboardInterrupt:
                break
            except:
                print "consuming exception!"
                traceback.print_exc()
                self.con.on_disconnect()
                self.init_channel()

    def on_request(self, ch, method, props, body):

        response = self.do_func(body)

        try:
            ch.basic_publish(exchange='',
                             routing_key=props.reply_to,
                             properties=pika.BasicProperties(correlation_id= \
                                                                 props.correlation_id),
                             body=response)
            ch.basic_ack(delivery_tag=method.delivery_tag)
        except:
            traceback.print_exc()
            self.con.on_disconnect()
            self.init_channel()

def do_something_func(task):
    #default task is json format
    taskjson = json.loads(task)
    response={}
    response['result'] = "Done from python " + os.environ['HOSTNAME']
    response['path'] = taskjson['path']
    response['method'] = taskjson['method']
    response['query'] = taskjson['query']

    try:
        body = taskjson['body']
        try:
            response['body'] = json.loads(body)
        except:
            response['body'] = body
    except:
        pass

    return json.dumps(response, ensure_ascii=False)

if __name__ == "__main__":
    input_que = "python_task"
    rabbitmq_ip = os.environ['RABBITMQ_HOST']
    rabbitmq_port = int(os.environ['RABBITMQ_PORT'])

    pika_con = PikaConection(rabbitmq_ip, rabbitmq_port )


    pika_server = PikaServer(pika_con,input_que,do_something_func)
    if pika_server.init_channel():
        pika_server.start_consume()
    else:
        print 'pika server init failed'

    pika_con.close()