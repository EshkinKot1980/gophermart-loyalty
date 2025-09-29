FROM golang:1.24

RUN wget http://github.com/golang-migrate/migrate/releases/latest/download/migrate.linux-amd64.deb &&\
    dpkg -i migrate.linux-amd64.deb &&\
    rm -f migrate.linux-amd64.deb

WORKDIR /usr/local/src