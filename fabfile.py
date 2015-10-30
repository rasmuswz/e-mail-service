from fabric.api import local

def status():
    """- This target is equivalent to git status."""
    local("git status .");

def commit(msg="Lazy commit no message given"):
    """- Commit taking argument msg, e.g. fab commit:msg=\"Hey cool stuff added\"."""
    local("git commit -am \"" + msg + "\"");

def add(file=""):
    """- Git add a file"""
    local("git add "+file);

def pull():
    """- Git pull """
    local("git pull");

def push():
    """- Git push changes to the cloud"""
    local("git push");


def install():
    """- Install e-mail-services web application on this machine"""
    print("Todo(rwz): Implement me");

def deploy():
    """- Deploy this e-mail-service to another remote machine"""
    print("Todo(rwz): Implement me");
