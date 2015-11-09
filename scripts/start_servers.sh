#!/bin/bash
# to be run from ..
#
# Start Servers
#
ApiDecryptionKey=${2}

function stop() {
    sudo pgrep -f clientapiserver | xargs sudo kill -9
    sudo pgrep -f backendserver | xargs kill -9
    sudo pgrep -f mtaserver | xargs kill -9
}


function start() {
    sudo goworkspace/bin/clientapiserver dartworkspace/build/web 443 > clientapi.log 2>&1 & 
    disown $!
    goworkspace/bin/mtaserver ${ApiDecryptionKey} > mtaserver.log 2>&1 &
    disown $!
    goworkspace/bin/backendserver > backend.log 2>&1 &
    disown $!
}

case ${1} in
    start)
	start
	;;
    stop)
	stop
	;;
    restart)
	stop
	start
	;;
	*)
	echo "Start script supports start, stop, and restart."
esac

