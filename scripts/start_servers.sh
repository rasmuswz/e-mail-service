#!/bin/bash
# to be run from ..
#
# Start Servers
#

function stop() {
    sudo pgrep -f clientapiserver | xargs sudo kill -9
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

