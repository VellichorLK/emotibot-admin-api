db = db.getSiblingDB('admin');
cursor = db.system.users.find({'user':'root'});

if(cursor.count()==0){
	
	// create root user (root user can operate all database)
	db.createUser({user:'root', pwd:'Emotibot1', roles:['root']});

	// create admin user (admin user can only operate admin db, like create new user, delete user, etc)
	db.createUser({user:'admin', pwd:'Emotibot1', roles:['dbOwner']});
	
	// create readWrite user for log db
	db.createUser({user:'logger', pwd:'Emotibot1', roles:[{role:'readWrite', db:'userlog'}]});
	
	// create index for logs
	db = db.getSiblingDB("userlog")
	db.createCollection("rawlogs")
	db.rawlogs.createIndex({uuid:-1})
	
}
else{
	printjson({'user':'does exit'});
}