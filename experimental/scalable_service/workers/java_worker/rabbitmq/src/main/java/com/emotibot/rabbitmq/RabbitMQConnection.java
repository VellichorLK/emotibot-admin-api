package com.emotibot.rabbitmq;

import java.io.IOException;
import java.util.concurrent.Semaphore;

import com.rabbitmq.client.*;

public class RabbitMQConnection {
	private String rabbitmq_host;
	private int rabbitmq_port;
	
	private ConnectionFactory factory;
	private Connection connection;
	private Semaphore lock;
	
	
	public RabbitMQConnection(String host, int port){
		rabbitmq_host = host;
		rabbitmq_port = port;
		
		lock = new Semaphore(1);
		connection = null;
	}
	
	public void Connect(){
		
		factory = new ConnectionFactory();
		
		factory.setHost(rabbitmq_host);
		factory.setPort(rabbitmq_port);
		
		while(connection == null){
			try{
				connection = factory.newConnection();
			}catch (Exception e){
				System.out.println("Connect to rabbitMQ " + rabbitmq_host + ":" + rabbitmq_port + " failed!");
				Sleep(5000);
				e.printStackTrace();
			}
		}
		
	}
	
	public Channel CreateChannel(){
		Channel ch = null;
		try {
			ch = connection.createChannel();
		} catch (IOException e) {
			// TODO Auto-generated catch block
			e.printStackTrace();
			OnDisconnect();
		}
		
		return ch;
	}
	
	public void OnDisconnect(){
		Boolean get_lock = false;
		
		get_lock = lock.tryAcquire();
		
		if(get_lock){
			connection = null;
			Connect();
			lock.release();
		}
		
	}
	
	public void Close(){
		try {
			connection.close();
		} catch (IOException e) {
			// TODO Auto-generated catch block
			e.printStackTrace();
		}
	}
	
	private void Sleep(int ms){
		try {
			Thread.sleep(ms);
		} catch (InterruptedException e) {
			// TODO Auto-generated catch block
			e.printStackTrace();
		}
	}
	
}