#!/bin/bash

## Installing Golang
sudo apt-get update
sudo snap install go --classic
go version

# set up Go lang path #
echo GOPATH=$HOME/go
echo PATH=$PATH:/usr/local/go/bin:$GOPATH/bin >> ~/.profile
source ~/.profile

## Installing Docker and Docker Compose
sudo snap install docker
sudo chown ubuntu:ubuntu /var/run/docker.sock

sudo curl -L "https://github.com/docker/compose/releases/download/1.24.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
docker-compose --version

sudo apt install make
sudo apt-get install jq
sudo apt  install awscli
aws configure

curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.1/install.sh | bash
echo export NVM_DIR="$HOME/.nvm" >> ~/.profile
source ~/.profile

echo ''
echo 'All the dependencies are installed...'
echo ''