
#!/bin/bash

echo "[CB-MCIS-Monitoring Dependent Component: Start to Delete CB-MCIS-Monitoring Dependent Component]"

echo "[CB-MCIS-Monitoring Dependent Component: UnInstall sysbench]"
sudo apt-get purge -y update
sudo apt-get purge -y install sysbench

echo "[CB-MCIS-Monitoring Dependent Component: UnInstall Ping]"
sudo apt-get purge -y iputils-ping

echo "[CB-MCIS-Monitoring Dependent Component: UnInstall debconf-utils]"
sudo apt-get purge -y debconf-utils
sudo debconf-set-selections <<< 'mysql-server mysql-server/root_password password psetri1234ak'
sudo debconf-set-selections <<< 'mysql-server mysql-server/root_password_again password psetri1234ak'

echo "[CB-MCIS-Monitoring Dependent Component: UnInstall MySQL]"
sudo DEBIAN_FRONTEND=noninteractive apt-get -y purge mysql-server

echo "[CB-MCIS-Monitoring Dependent Component: Generate dump tables for evaluation]"

mysql -u root -ppsetri1234ak -e "CREATE DATABASE sysbench;"
mysql -u root -ppsetri1234ak -e "CREATE USER 'sysbench'@'localhost' IDENTIFIED BY 'psetri1234ak';"
mysql -u root -ppsetri1234ak -e "GRANT ALL PRIVILEGES ON *.* TO 'sysbench'@'localhost' IDENTIFIED  BY 'psetri1234ak';"

echo "[CB-MCIS-Monitoring Dependent Component: Deletion is done]"








