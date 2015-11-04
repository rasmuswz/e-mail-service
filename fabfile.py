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
# screen
#
# at least one hosts needs MySQL for the Storage component.
#

#
# Download Go Dependencies and go install mail.bitlab.dk
# Finally if successful wrap the source up in go_tag.tgz.
#
# param tag - the git hub tag
#
# Note! Go create platform specific binaries meaning that we need to
# deploy the source and rebuild it on the production environment.
#
def build_goworkspace(tag):
    if not exists("go_"+tag+".tgz"):
        with lcd("goworkspace"):
            with shell_env(GOPATH=os.path.realpath("goworkspace")):
                local("go get github.com/mailgun/mailgun-go");
                local("go install mail.bitlab.dk");
                local("tar cmvzf ../go_"+tag+".tgz --exclude .git ./src");
            
    

#
# Get dependencies and build the dart client UI
#
def build_dartworkspace(tag):
    if not exists("dart_"+tag+".tgz"):
        with lcd("dartworkspace"):
            local("pub get");
            local("pub build");
            local("tar cmvzf ../dart_"+tag+".tgz --exclude .git ./build");

#
# Make git-tag
#
def make_git_tag():
    return local("git rev-parse --short HEAD", capture=True).strip();

#
# Acquire the git-tag locally and make a directory on the remote
# server called mail.bitlab.dk_<git-tag>
#
def make_and_return_name_of_tagged_directory(tag):
    taggedDir="mail.bitlab.dk_"+tag
    run("mkdir -p "+taggedDir)
    return taggedDir
    
#
# Build (if necessary) and Send dart.tgz and go.tgz to the remote host
# and unpacking them in {taggedDir}.
#
def transfer_and_unpack_tarballs(taggedDir,tag):
    dartTarBall="dart_"+tag+".tgz"
    goTarBall="go_"+tag+".tgz"
    if not exists(taggedDir+"/"+dartTarBall):
        build_dartworkspace(tag)
        put(dartTarBall,taggedDir);
    if not exists(taggedDir+"/"+goTarBall):
        build_goworkspace(tag)
        put(goTarBall,taggedDir);
    run("mkdir -p "+taggedDir+"/dartworkspace");
    run("mkdir -p "+taggedDir+"/goworkspace");
    run("tar xmfz "+taggedDir+"/"+dartTarBall+ " -C "+taggedDir+"/dartworkspace");
    run("tar xmfz "+taggedDir+"/"+goTarBall+" -C "+taggedDir+"/goworkspace");

#
# Download GoSDK and unpack it properly on the remote server
#
def get_os_specific_GO_into(d):
    ostype=run("uname -s");
    if ostype.lower() == "linux":
        run("echo we are on linux box");
        run("wget --no-check-certificate  "+
            "https://storage.googleapis.com/golang/go1.5.1.linux-amd64.tar.gz")
        run("tar xfz go1.5.1.linux-amd64.tar.gz -C "+d)
    if ostype.lower() == "freebsd":
        run("echo we are on freebsd box");
        run("wget --no-check-certificate  "+
            "https://storage.googleapis.com/golang/go1.5.1.freebsd-amd64.tar.gz");
        run("tar xfz go1.5.1.freebsd-amd64.tar.gz -C"+d);
    if ostype.lower() == "darwin":
        run("echo we are on a darwin box, TODO(rwz): Not implemented yet");

#
# Install GoSDK on the remote and return the Path to Go-Tools
# executables directory, aka go/bin.
#
# Note! Go create platform specific binaries meaning that we need to
# deploy the source and rebuild it on the production environment.
#
def check_for_and_install_GOSDK_on_remote(taggedDir):
    d=run("pwd").strip();
    if not exists(d+"/go"):
        get_os_specific_GO_into(d)
    return d+"/go/bin"

#
#
#
def buildGoWorkspace(goBinDir,goWorkspaceDir):
    goPath=make_go_path(goWorkspaceDir)
    with cd(goWorkspaceDir):
        with shell_env(GOPATH=goPath,
                       GOROOT=goBinDir+"/.."):
            run("PATH=${PATH}:"+goBinDir+ " && go install mail.bitlab.dk");
            run("PATH=${PATH}:"+goBinDir+ " && go install mail.bitlab.dk/utilities/webserver");

def make_go_path(goWorkspaceDir):
    with cd(goWorkspaceDir):
        return run("pwd").strip();

    

def sync_with_git():
    local("git pull");
    local("git commit -am \"Deploying standby\" || true ");
    local("git pull");


def start_server(taggedDir):
    webServerExec=taggedDir+"/goworkspace/bin/webserver"
    webServerRoot=taggedDir+"/dartworkspace/build/web"
    webServerSession="webserver"
    mtaServerSession="mtaserver"

    sudo("screen -S "+webServerSession+" -X quit || true")
    run("screen -S "+mtaServerSession+" -X quit || true")

    sudo("screen -d -m -S "+webServerSession+" "+webServerExec+" "+webServerRoot)
    run("screen -d -m -S "+mtaServerSession+" "+taggedDir+"/goworkspace/bin/mail.bitlab.dk")

#
# Deploy the service to the mail.bitlab.dk servers.
#
@hosts(['ubuntu@mail1.bitlab.dk','rwz@mail0.bitlab.dk'])
def deploy():
    
    sync_with_git()

    run("mkdir -p deploy");

    with cd("deploy"):
        
        tag = make_git_tag()

        taggedDir = make_and_return_name_of_tagged_directory(tag)

        transfer_and_unpack_tarballs(taggedDir,tag)

        absGoBinDir = check_for_and_install_GOSDK_on_remote(taggedDir)

        buildGoWorkspace(absGoBinDir,taggedDir+"/goworkspace")
        
        start_server(taggedDir)

