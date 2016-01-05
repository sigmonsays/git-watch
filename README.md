git-watch
=============================

watch a git repository for changes and execute update commands

features
=============================

- be generic as possible
- minimal dependencies
- intelligent restart
  - if install fails, we do not restart

install/Upgrade
=============================

    export GOPATH=$HOME/go
    go get github.com/sigmonsays/git-watch/git-watch/...

To upgrade

    go get -u github.com/sigmonsays/git-watch/git-watch/...

pre-compiled binaries    
=============================
See http://gobuild.io/github.com/sigmonsays/git-watch/git-watch
   
quickstart
=============================


Usage:

    Usage of git-watch:
      -check=30: git check interval
      -config="git-watch.yaml": git watch config

an example which rebuilds a docker container when upstream changes are detected. See
the examples directory for more options.

**git-watch.yaml**

    checkinterval: 5
    localbranch: master
    execcmd: make start
    updatecmd: make
    installcmd: make install

**Makefile**

    all:
       docker build -t api .

    install:
       # nothing to do

    start:
       docker stop api
       docker rm api
       docker run -i --rm --name api api


Then simply start git-watch in the git checkout directory and it'll begin monitoring it for changes. 

If the update command or install command fails, the process will not restart

configuration
=============================
