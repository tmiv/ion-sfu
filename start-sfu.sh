#! /usr/bin/bash

export EXTERNAL_IP=$(curl -s -H "Metadata-Flavor: Google" http://metadata/computeMetadata/v1/instance/network-interfaces/0/access-configs/0/external-ip)

if [ -z $EXTERNAL_IP ] 
    then
       export EXTERNAL_IP=127.0.0.1
fi

cat /srv/config.template.toml | sed "s/%%IP_ADDRESS_EXT%%/$EXTERNAL_IP/g" > /srv/config.toml
/srv/sfu -c /srv/config.toml

