#
# Tooling 
#
# Author: Rasmus Winther Zakarias
#
# This file is used to deploy e-mail-service
# on a linux/OSX box after it has been built.
#
from fabric.api import *

#
# deployment
#
env.hosts = ['ubuntu@mail1.bitlab.dk','rwz@mail0.bitlab.dk']

#
#
#
def deploy():
    with cd ("build"):
        put("release.tgz","~");
        run("tar xfz release.tgz");
        run("pub serve");
