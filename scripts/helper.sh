#
#
# Author: Rasmus Winther Zakarias
#
# 
# 

#
# Term Colors
#
NORM="\033[0;0m"
GREEN="\033[32;32m"
RED="\033[0;31m"
BLUE="\033[0;34m"
YELLOW="\033[33;33m"

#
# Printing functions
#
function warning() {
    echo -e "[${RED}WARN${NORM}] ${1}"
}


function goodnews() {
    echo -e "[${GREEN}Good${NORM}] ${1}"
}

function died() {
    echo -e "[${RED}Died${NORM}] ${1}"
    exit -1
}

#
# $(check_exec_exists ${file}) == "0" if ${file} exists and
# is executable.
#
function check_exec_exists() {
    if [ -x ${1} ]; then
	echo "0"
    else
	echo "-1"
    fi
}

function action() {
    printf "%-50s " "${1}"
    R=$(${2})
    if [ "$?" == "0" ]; 
    then
	echo -e "\t[${GREEN} OK ${NORM}]"
	echo ${R}
    else
	echo -e "\t[${RED}FAIL${NORM}]"
	exit -1
    fi
}

function check_priv() {
    if [ "$(id -u)" == "0" ];
    then
	return 0
    else
	return 1
    fi
}

#
# Map implemented as env vars
#
put() {
    eval hash"$1"='$2';
}

get() {
    v="hash${1}" && eval echo '${'${v}'}'
}



