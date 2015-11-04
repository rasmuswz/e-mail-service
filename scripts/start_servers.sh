#!/bin/bash
# to be run from ..
#
# Start Servers
#

sudo screen -S "clientapi" -X quit 
sudo screen -dmS "clientapi" goworkspace/bin/clientapiserver dartworkspace/build/web 443

screen -S "backend" -X quit 
screen -dmS "backend" goworkspace/bin/backendserver

screen -S "mtaserver" -X quit 
screen -dmS "mtaserver" goworkspace/bin/mtaserver

