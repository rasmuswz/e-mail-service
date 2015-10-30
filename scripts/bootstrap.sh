#!/bin/bash
#
# Tool chain boot-strap script
#
#

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
    end_action 0
}

function install_pip() {
    start_action "Installing pip"
    end_action 0
}

function install_fabric() {
    start_action "Installing fabric"
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


echo "Done."
