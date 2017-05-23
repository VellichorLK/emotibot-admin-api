package com.emotibot.rabbitmq;

import org.json.*;
/**
 * Hello world!
 *
 */


public class App extends RabbitMQServer
{

	
	public App(String queue, RabbitMQConnection con) {
		super(queue, con);
		// TODO Auto-generated constructor stub
	}
	
	//define your work in DoWork function
	@Override
	public String DoWork(String task){
		String res;
		String path,method,query,body;
		JSONObject obj = new JSONObject(task);
		
		method = obj.getString("method");
		path = obj.getString("path");
		query = obj.getString("query");
		
		//here notice only the owner knows the body format, parse body according your defined format
		try{
			body = obj.getString("body");
		}catch (JSONException e){
			body="not string type";
		}
		//---------------------------------------------------------
		
		System.out.println("path: " + path + " method: "+ method + " query: " + query + " body: " + body);
		
		res = "Done " + path + " " + method + " query: " + query + 
				" body: " + body + " from java " + System.getenv("HOSTNAME");

		return res;
	}
	
	public static void main( String[] args ){
		
		//get env to connect to rabbitmq server
    	String rabbitmq_host = System.getenv("RABBITMQ_HOST");
    	int rabbitmq_port =  Integer.valueOf(System.getenv("RABBITMQ_PORT"));
		
    	//define your task queue name to receive tasks
    	String queue_name = "java_task";
    	
    	
    	//create a new connection
    	RabbitMQConnection rc = new RabbitMQConnection(rabbitmq_host, rabbitmq_port);
    	rc.Connect();
    	
    	
    	//init channel to rabbitmq and start consuming the task
    	App app = new App(queue_name, rc);    	
    	if(app.InitChannel()){
    		app.StartConsuming();
    		
    		//loop forever
            while(true) {
                try {
                  Thread.sleep(1000);
                } catch (InterruptedException _ignore) {}
             }
            
    	}
    	
	}
	
}
