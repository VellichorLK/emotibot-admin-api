FROM python:2.7

RUN pip install pika

RUN mkdir -p /usr/src/app

COPY python_worker /usr/src/app

CMD ["python", "/usr/src/app/pika_server.py"]
