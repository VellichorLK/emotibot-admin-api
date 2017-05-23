package com.emotibot.rabbitmq;

import java.io.IOException;

import com.rabbitmq.client.AMQP;
import com.rabbitmq.client.Channel;
import com.rabbitmq.client.Consumer;
import com.rabbitmq.client.DefaultConsumer;
import com.rabbitmq.client.Envelope;

public abstract class RabbitMQServer {
	private String task_queue;
	private RabbitMQConnection connection;
	private Channel ch;
	public RabbitMQServer(String queue, RabbitMQConnection con){
		task_queue = queue;
		connection = con;
		ch = null;
	}
	
	
	public boolean InitChannel(){
		boolean res = false;
		try{
			ch = connection.CreateChannel();
			ch.queueDeclare(task_queue, false, false, false, null);
			ch.basicQos(1);
			res = true;
		}catch (Exception e){
			e.printStackTrace();
			connection.OnDisconnect();
		}
		
		return res;
	}
	
	
	public Channel GetChannel(){
		return ch;
	}
	
	public void StartConsuming(){
		Consumer consumer = new DefaultConsumer(ch) {

            @Override
            public void handleDelivery(String consumerTag, Envelope envelope, AMQP.BasicProperties properties, byte[] body) throws IOException {
                AMQP.BasicProperties replyProps = new AMQP.BasicProperties
                        .Builder()
                        .correlationId(properties.getCorrelationId())
                        .build();

                String response = "";

                try {
                    String message = new String(body,"UTF-8");
                    response = DoWork(message);
                }
                catch (RuntimeException e){
                    System.out.println(" [.] " + e.toString());
                }
                finally {
                	
                	this.getChannel().basicPublish( "", properties.getReplyTo(), replyProps, response.getBytes("UTF-8"));

                	this.getChannel().basicAck(envelope.getDeliveryTag(), false);
                }
            }
        };            
        
        try {
        	ch.basicConsume(task_queue, false, consumer);
		} catch (IOException e) {
			// TODO Auto-generated catch block
			e.printStackTrace();
		}
	}
	
	public abstract String DoWork(String task);
	
	
}
