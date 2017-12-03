#!/bin/bash

# folder on home for server
folder="server"

# make server directory if it doesn't exist
mkdir -p ~/$folder

# build server
go build -o twitch-eos-thanks-server

# move server to $HOME/server
mv ./twitch-eos-thanks-server ~/$folder/twitch-eos-thanks-server

# copy service.sh
cp ./service.sh ~/$folder/service.sh

# chmod service to be execuable
chmod +x ~/$folder/service.sh

# check for config, and if not copy default
if [ ! -f ~/$folder/config.json ]; then
    cp ./config.default.json ~/$folder/config.json
fi

# complete
echo installation complete.