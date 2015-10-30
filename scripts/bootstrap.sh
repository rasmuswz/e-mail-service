#!/bin/bash
#
# Tool chain boot-strap script
#
#

OS=`uname -s`

PYTHON=`which python`
PIP=`which pip`
FABRIC=`which fab`

GREEN="\033[1;33m"
NORM="\033[0;1m"

function start_action() {
    echo -ne "${1}\t\t\t ... ";
}

function end_action() {
    [ "${1}" == "0" ] && 
    echo -e "[\033[33;1m OK \033[0;1m]" ||
    echo -e "[\033[31;1mFAIL\033[0;1m]"
}

function install_python() {
    start_action "Installing Python "
    if [ "${OS}" == "Darwin" ]; then
	[ -z `which brew` ] || brew install python
	[ -z `which port` ] || sudo port install python
    fi

    if [ "${OS}" == "FreeBSD" ]; then
	sudo pkg install python27
    fi

    if [ "${OS}" == "Linux" ]; then 
	APTGET=`which apt-get`
	if [ -z ${APTGET} ]; then
	    echo "Error: We only support Ubuntu Linux"
	    exit -1
	fi
	sudo apt-get install python
    fi

	    
    end_action 0
}

function install_pip() {
    start_action "Installing pip"
    sudo easy_install pip
    end_action 0
}

function install_fabric() {
    start_action "Installing fabric"
    sudo pip install fabric
    end_action 0
}


if [ -z ${PYTHON} ]; then
    install_python
fi

if [ -z ${PIP} ]; then
    install_pip
fi

if [ -z ${FABRIC} ]; then
    install_fabric
fi


echo "Done. Type \"fab install\" to install this package."
