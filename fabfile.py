#
# Tooling 
#
# Author: Rasmus Winther Zakarias
#
# This file is used to deploy e-mail-service
# on a linux/OSX box after it has been built.
#
from fabric.contrib.project import rsync_project
from fabric.contrib.files import exists
from fabric.api import *
import os
import subprocess


#
# Dependencies installed on deployed hosts:
#
# rsync
# bash
# uname -s
# mkdir
#

#
# Prepare:
# 
# Download go dependencies 
#
def build_goworkspace():
    with lcd("goworkspace"):
        with shell_env(GOPATH=os.path.realpath("goworkspace")):
            local("go get github.com/mailgun/mailgun-go");
            local("go install mail.bitlab.dk");
            local("tar cvzf ../go.tgz --exclude .git ./src");

    

#
# Get dependencies and build the dart client UI
#
def build_dartworkspace():
    with lcd("dartworkspace"):
        local("pub get");
        local("pub build");
        local("tar cvzf ../dart.tgz --exclude .git ./build");

#
# 
#
@hosts(['ubuntu@mail1.bitlab.dk','rwz@mail0.bitlab.dk'])
def deploy():
    run("mkdir -p deploy");
    with cd("deploy"):
        taggedDir="mail.bitlab.dk_"+local("git rev-parse --short HEAD", capture=True).strip();
        run("mkdir -p "+taggedDir);
        if not exists(taggedDir+"/dart.tgz"):
            put("dart.tgz",taggedDir);
        if not exists(taggedDir+"/go.tgz"):
            put("go.tgz",taggedDir);
        run("mkdir -p "+taggedDir+"/dartworkspace");
        run("mkdir -p "+taggedDir+"/goworkspace");
        run("tar xfz "+taggedDir+"/dart.tgz -C "+taggedDir+"/dartworkspace");
        run("tar xfz "+taggedDir+"/go.tgz -C "+taggedDir+"/goworkspace");
        

            
        basePath=run("pwd");
        if not exists('go'):
            ostype=run("uname -s");
            if "linux" in ostype or "Linux" in ostype:
                if not exists("go1.5.1.freebsd-amd64.tar.gz"):
                    run("wget --no-check-certificate  "+
                        "https://storage.googleapis.com/golang/go1.5.1.linux-amd64.tar.gz");
                run("tar xfz go1.5.1.linux-amd64.tar.gz");
            if "FreeBSD" in ostype:
                if not exists("go1.5.1.freebsd-amd64.tar.gz"):
                    run("wget --no-check-certificate  "+
                        "https://storage.googleapis.com/golang/go1.5.1.freebsd-amd64.tar.gz");
                run("tar xfz go1.5.1.freebsd-amd64.tar.gz");
            if "darwin" in ostype:
                run("wget --no-check-certificate"+
                    " https://storage.googleapis.com/golang/go1.5.1.darwin-amd64.tar.gz");
                run("tar xfz go1.5.1.darwin-amd64.tar.gz");
        with shell_env(PATH="${PATH}:"+basePath+"/go/bin",
                       GOROOT=basePath+"/go",
                       GOPATH=basePath+"/"+taggedDir+"/goworkspace"):
            print("${PATH}:"+basePath+"/go/bin");
            with cd(taggedDir+"/goworkspace"):
                run("PATH=${PATH}:"+basePath+"/go/bin && go install mail.bitlab.dk");
                
    

