git-watch
=============================

watch a git repository for changes and execute update commands

features
=============================

- minimal dependancies


quickstart
=============================

an example which rebuilds a docker container when upstream changes are detected

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
       ocker rm api
       ocker run -i --rm --name api api


Then simply start git-watch in the git checkout directory and it'll begin monitoring it for changes. 

If the update command or install command fails, the process will not restart

configuration
=============================
