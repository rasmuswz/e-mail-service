#!/bin/bash
# to be run from ..
#
# Start Servers
#

#sudo screen -S "clientapi" -X quit

function stop() {
    sudo kill -9 `ps aux | grep clientapiserver | colrm 1 6 | colrm 6` >/dev/null 2>&1
    screen -S "backend" -X quit
    screen -S "mtaserver" -X quit 
}


function start() {
    sudo screen -dmS "clientapi" goworkspace/bin/clientapiserver dartworkspace/build/web 443
    screen -dmS "backend" goworkspace/bin/backendserver
    screen -dmS "mtaserver" goworkspace/bin/mtaserver
}

case ${1} in
    start)
	start
	;;
    stop)
	stop
	;;

    *)
	stop
	start;
	;;
esac

