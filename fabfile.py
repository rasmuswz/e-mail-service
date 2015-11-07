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
# kill
# pgrep
#
# at least one hosts needs MySQL for the Storage component.
#

#
# Download Go Dependencies and go install our three services
# BackEnd, MTA and ClientApi.
#
# if successful wrap the source up in a tarball ready for deployment.
#
# param tag - the git hub tag
#
# Note! Go create platform specific binaries meaning that we need to
# deploy the source and rebuild it on the production environment.
#
buildCmdPrefix="go install mail.bitlab.dk/";
def build_goworkspace(tag):
    path=os.environ["PATH"];
    path=path+":"+os.path.realpath("thirdparty/go/bin")
    path=path+":"+os.path.realpath("thirdparty/dart-sdk/bin");


    if not exists("go_"+tag+".tgz"):
        with lcd("goworkspace"):
            with shell_env(GOPATH=os.path.realpath("goworkspace"),
                           PATH=path):

                local("go get github.com/mailgun/mailgun-go");
                local("go get github.com/aws/aws-sdk-go/service/ses");
                local(buildCmdPrefix+"backend/backendserver");
                local(buildCmdPrefix+"clientapi/clientapiserver");
                local(buildCmdPrefix+"mtacontainer/mtaserver");
                local("tar cmvzf ../go_"+tag+".tgz --exclude .git ./src");
            
def build_remote_goworkspace(goBinDir,goWorkspaceDir):
    goPath=make_go_path(goWorkspaceDir)
    with cd(goWorkspaceDir):
        with shell_env(GOPATH=goPath,
                       GOROOT=goBinDir+"/.."):
            setGoPathPrefix="PATH=${PATH}:"+goBinDir+" && ";
            run(setGoPathPrefix+buildCmdPrefix+"backend/backendserver");
            run(setGoPathPrefix+buildCmdPrefix+"clientapi/clientapiserver");
            run(setGoPathPrefix+buildCmdPrefix+"mtacontainer/mtaserver");

    

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
# Deploy self signed certificate for mail.bitlab.dk 
# with its private key.
#
def decrypt_pack_and_send_certificate(taggedDir,tag):
    certFile="cert_"+tag+".tgz"
    if not exists(taggedDir+"cert.pem"):
        local("openssl rsa -in protectedkey.pem -out key.pem"); # Decrypt key
        local("tar cmvzf cert_"+tag+".tgz cert.pem key.pem scripts");
        put(certFile,taggedDir)
        run("tar xmfz "+taggedDir+"/"+certFile+" -C "+taggedDir);

            

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


def make_go_path(goWorkspaceDir):
    with cd(goWorkspaceDir):
        return run("pwd").strip();

def sync_with_git():
    local("git pull");
    local("git commit -am \"Deploying standby\" || true ");
    local("git push");

def start_service_cmd(sesName,exe,root,port,logFile):
    return "screen -dmS "+sesName+" sh -c '"+exe+" "+root+" "+port+" >"+logFile+" 2>&1'";


def restart_named_screen_session(taggedDir,dosudo,cmd,name):
    quitCmd="screen -S "+name+" -X quit || true" 
    sesName=name;
    logFile=taggedDir+"/"+name+".log"
#    startCmd="screen -dmS "+sesName+" sh -c '"+exe+" "+root+" "+port+" >"+logFile+" 2>&1'"
    startCmd="screen -dmS "+sesName+" sh -c '"+cmd+" >"+logFile+" 2>&1'"
    if dosudo:
        sudo(quitCmd)
        sudo(startCmd)
    else:
        run(quitCmd)
        run(startCmd)


def start_clientapi_server(taggedDir):
    clientApiSrvExe="goworkspace/bin/clientapiserver"
    docRoot=taggedDir+"/dartworkspace/build/web";
    apiPort="443";
    exe=taggedDir+"/"+clientApiSrvExe;
    cmd=exe +" " + docRoot + " " + apiPort + " "
    restart_named_screen_session(taggedDir,True,cmd,"ClientApi")

def start_backend_server(taggedDir):
    backendSrvExe="goworkspace/bin/backendserver"
    restart_named_screen_session(taggedDir,False,backendSrvExe,"Backend");

def start_mta_server(taggedDir):
    mtaSrvExe="goworkspace/bin/mtaserver"
    restart_named_screen_session(taggedDir,False,mtaSrvExe,"MTAServer");

def start_servers(taggedDir):
    with cd(taggedDir):
        run("scripts/start_servers.sh");

def write_tag_in_file(filename,tag, destination):
    f = open(filename,"w");
    f.write(tag);
    f.close();
    if (destination != None):
        put(filename,destination);

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

        build_remote_goworkspace(absGoBinDir,taggedDir+"/goworkspace")
        
        decrypt_pack_and_send_certificate(taggedDir,tag)

        write_tag_in_file("dartworkspace/build/web/version.txt", tag, taggedDir+"/dartworkspace/build/web/version.txt");
        write_tag_in_file("dartworkspace/web/version.txt",tag,None);

        start_servers(taggedDir)

        print("Version "+tag+" has been deployed");


@hosts(['ubuntu@mail1.bitlab.dk','rwz@mail0.bitlab.dk'])
def demo():

    with cd("deploy"):

        tag = make_git_tag();

        taggedDir = make_and_return_name_of_tagged_directory(tag);

        write_tag_in_file("dartworkspace/build/web/version.txt", tag, taggedDir+"/dartworkspace/build/web");

