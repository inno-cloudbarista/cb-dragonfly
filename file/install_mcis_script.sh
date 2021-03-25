
#!/bin/bash

echo "[CB-MCIS-Monitoring Dependent Component: Start to prepare a VM evaluation]"

echo "[CB-MCIS-Monitoring Dependent Component: Install sysbench]"
sudo apt-get -y update
sudo apt-get -y install sysbench

echo "[CB-MCIS-Monitoring Dependent Component: Install Ping]"
sudo apt-get -y install iputils-ping

echo "[CB-MCIS-Monitoring Dependent Component: Install debconf-utils]"
sudo apt-get -y install debconf-utils
sudo debconf-set-selections <<< 'mysql-server mysql-server/root_password password psetri1234ak'
sudo debconf-set-selections <<< 'mysql-server mysql-server/root_password_again password psetri1234ak'

echo "[CB-MCIS-Monitoring Dependent Component: Install MySQL]"
sudo DEBIAN_FRONTEND=noninteractive apt-get -y install mysql-server

echo "[CB-MCIS-Monitoring Dependent Component: Generate dump tables for evaluation]"

mysql -u root -ppsetri1234ak -e "CREATE DATABASE sysbench;"
mysql -u root -ppsetri1234ak -e "CREATE USER 'sysbench'@'localhost' IDENTIFIED BY 'psetri1234ak';"
mysql -u root -ppsetri1234ak -e "GRANT ALL PRIVILEGES ON *.* TO 'sysbench'@'localhost' IDENTIFIED  BY 'psetri1234ak';"

echo "[CB-MCIS-Monitoring Dependent Component: Preparation is done]"





