FROM golang

ADD . /go/src/orangetheory-api
WORKDIR /go/src/orangetheory-api

RUN git config --global url.'https://c6ab1ff345ba7a004ef2bc99fc577329d6c4ce0b:x-oauth-basic@github.com/'.insteadOf 'https://github.com/'

WORKDIR /go/src/orangetheory-api

COPY . ./

# Go dep!
RUN go get -u github.com/golang/dep/...
RUN dep ensure

RUN go build

CMD './orangetheory-api'