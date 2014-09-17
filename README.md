git-watch
=============================

watch a git repository for changes and execute update commands

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
       pwd
       docker build -t api .

    install:
       docker stop api
       ocker rm api
       ocker run -i --rm --name api api



