#!/bin/bash
#
# Tool chain boot-strap script
#
# Author: Rasmus Winther Zakarias
#
# Description:
#
# This script boot-straps the build process downloading the required
# SDKs and tools. For this boot-strap processing we assume the
# following least requirements to be met:
#
#  - bash-3.2.57 or later
#  - OSX: port/brew 
#  - Linux: apt-get
#  - Optional: color terminal (make output nicer)
#

#
# Load helper
#
SCRIPT_DIR=$(dirname ${BASH_SOURCE[0]})
source ${SCRIPT_DIR}/helper.sh

#
# Result from checks are mapped here
#
put "CHKPRIV" "check_priv"

#
# Check bash version
#
function check_bash() {
    if [[ "${BASH_VERSION}" < "3.2" ]]; then
	warning "Bash version is less than 3.2"
    else
	goodnews "Bash version checks out"
    fi
}

#
# On Linux we require apt-get
#
function check_apt_get() {
    APTGET=`which apt-get`
    if [ $(check_exec_exists ${APTGET}) == "0" ];
    then
	goodnews "${APTGET} was found" ||
	put "INSTALLER" "sudo ${APTGET} install python"
    else
      died "required tool ${APTGET} on this platform wasn't found"
    fi
}

#
# On FreeBSD we require pkg
#
function check_pkg() {
    PKG=`which pkg`
    if [ $(check_exec_exists ${PKG}) == "0" ]; 
    then
	goodnews "${PKG} was found"
	put "INSTALLER" "sudo pkg install python"
    else
	died "required tool ${PKG} on this platform wasn't found"
    fi
}

#
# On OSX we require either Brew or MacPorts
#
function check_brew_or_port() {
    BREW=`which brew`
    PORT=`which port`
    if [ $(check_exec_exists ${BREW}) == "0" ];
    then
	goodnews "${BREW} was found"
	put "CHKPRIV" "true"
	put "INSTALLER" "brew install python"
	return
    fi

    if [ $(check_exec_exists ${PORT}) == "0" ] ;
    then
	goodnews "${PORT} was found"
	put "INSTALLER" "sudo port port install python"
    fi

    if [ -z $(get "INSTALLER") ]; 
    then
	died "${BREW} nor ${PORT} were found please install one of them"
    fi
}

#
# Check install tool based on OSTYPE
#
function check_install_tool() {
    case "${OSTYPE}" in
	linux*)
	    check_apt_get ;;
	freebsd*)
	    check_pkg ;;
	darwin*)
	    check_brew_or_port ;;
	*)
	    died "OS type is not recognized unable to determine package manager"
    esac
}

#
# Check whether we need to install python
#
function check_python() {
    PYTHON=`which python`
    if [ -x ${PYTHON} ];
    then
	  PYVER=$(${PYTHON} -V 2>&1)
	  if [[ ${PYVER} < "Python 2.7" ]] &&
	     [[ ${PYVER} > "Python 2.8" ]];
	  then
	      warn "We prefer Python 2.7 version ${PYVER} was found".
	  else
	      goodnews "${PYTHON} in version ${PYVER} was found"
	  fi
    else
	  return 1
    fi
      return 0
}

#
# Run installer command
#
function install_python() {
    $($(get "INSTALLER"))
}

function check_fabric() {
    FAB=`which fab2`
    if [[ -x ${FAB} ]]; 
    then
	goodnews "Fabric is installed"
	return 0
    fi
    return 1
}

function check_pip() {
    PIP=`which pip`
    if [[ -x ${PIP} ]]; 
    then 
	goodnews "Pip is installed"
	put "PIP" "${PIP}"
	return 0
    fi
    return 1
}

function download() {
    echo $(get "PYTHON")
    echo "import urllib; urllib.urlretrieve('${1}','${2}');" | python
}

function install_pip() {
    download "https://bootstrap.pypa.io/get-pip.py"
    if [ "$?" != "0" ];
    then
	return -1;
    fi
    cat "get-pip.py" | $(get "PYTHON")
    if [ "$?" != "0" ];
    then
	return -1;
    fi
    return 0
}

function install_fabric() {
    $(get "PIP") install fabric
    if [ "$?" == "0" ];
    then 
	return 0;
    else
	return -1;
    fi
}

function install_golang_sdk() {
    mkdir -p ${SCRIPT_DIR}/../thirdparty


    if [[ "${OSTYPE}" =~ "linux" ]]; then
	    if [[ ! -f "go.tgz" ]]; then
	      echo "Downloading";
	      download "https://storage.googleapis.com/golang/go1.5.1.linux-amd64.tar.gz" "go.tgz"
	    fi
	    tar xfz go.tgz -C ${SCRIPT_DIR}/../thirdparty
	fi

	if [[ "${OSTYPE}" == "FreeBSD" ]]; then
	    if [[ ! -f "go.tgz" ]]; then
	      download "https://storage.googleapis.com/golang/go1.5.1.freebsd-amd64.tar.gz" "go.tgz"
	    fi
	    tar xfz go.tgz -C ${SCRIPT_DIR}/../thirdparty
	fi

	if [[ "${OSTYPE}" =~ "darwin" ]]; then
	    if [[ ! -f "go.pkg" ]]; then
	      download "https://storage.googleapis.com/golang/go1.5.1.darwin-amd64.pkg" "go.pkg"
	    fi
	    sudo `which installer` -pkg go.pkg -target /
	 fi

    if [[ ! -f go.zip ]] && [[ ! -f go.pkg ]]; then
	    died "Didn't have go-lang sdk download for OSTYPE ${OSTYPE}"
	    return 1;
	 else
	    return 0;
    fi
}

function install_dart_sdk() {

    mkdir -p ${SCRIPT_DIR}/../thirdparty

	if [[ "${OSTYPE}" =~ "linux" ]]; then
  	    if [[ ! -f "dart.zip" ]]; then
	      download "https://storage.googleapis.com/dart-archive/channels/stable/release/1.12.2/sdk/dartsdk-linux-x64-release.zip" dart.zip
	    fi
	    unzip dart.zip -d ${SCRIPT_DIR}/../thirdparty
	fi

	if [[ "${OSTYPE}" == "FreeBSD" ]]; then
	    echo "Sorry the Dart-SDK port for FreeBSD is experimental see the FreeBSD 11 forum"
	fi

	if [[ "${OSTYPE}" =~ "darwin" ]]; then
	    if [[ ! -f "dart.zip" ]]; then
	      download "https://storage.googleapis.com/dart-archive/channels/stable/release/1.12.2/sdk/dartsdk-macos-x64-release.zip" "dart.zip"
	    fi
	    unzip dart.zip -d ${SCRIPT_DIR}/../thirdparty
	fi

    if [[ ! -f dart.zip ]]; then
	    died "Didn't have dart-lang sdk download for OSTYPE ${OSTYPE}"
    fi

    if [ "$?" == "0" ];
    then
	return 0;
    fi
    return 1;

}

function check_go_sdk() {
    GO=`which go`
    if [[ -x ${GO} ]];
    then
	return 0;
    fi
    return 1
}

function check_dart_sdk() {
    DART=`which dart`
    if [[ -x ${DART} ]];
    then
	return 0;
    fi
    return 1
}

function main() {
    check_bash
    check_python
    if [ "$?" == "1" ]; then
	check_install_tool
	action "Checking priviledges" $(get "CHKPRIV")
	action "Installing Python using $(get "INSTALLER") " install_python
    else
	  put "PYTHON" "`which python`"
    fi 

    check_fabric
    if [ "$?" == "1" ]; then
	check_pip
	if [ "$?" == "1" ]; 
	then
	    action "Installing pip" install_pip
	fi
	action "Installing Fabric: " install_fabric
    fi

    check_go_sdk
    if [ "$?" == "1" ]; then
	action "Installing Google Go" install_golang_sdk
    fi

    check_dart_sdk
    if [ "$?" == "1" ]; then
	   action "Installing Dart SDK" install_dart_sdk
    fi

    goodnews "Everything seems to be in order."
    echo -e "\n\nYou might want to add `realpath ${SCRIPT_DIR}/../thirdparty`/**/bin"
    echo "to your path."
}

install_dart_sdk
install_golang_sdk
