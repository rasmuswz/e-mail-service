from fabric.api import local

def git_status():
    """- This target is equivalent to git status."""
    local("git status .");

def git_commit(msg="Lazy commit no message given"):
    """- Commit taking argument msg, e.g. fab git_commit:msg=\"Hey cool stuff added\"."""
    local("git commit -am \"" + msg + "\"");

