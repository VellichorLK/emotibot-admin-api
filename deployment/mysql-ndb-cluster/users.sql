CREATE USER 'deployer'@'%' IDENTIFIED BY 'password';
GRANT ALL PRIVILEGES ON *.* TO 'deployer'@'%' WITH GRANT OPTION;