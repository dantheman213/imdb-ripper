FROM ubuntu:20.04 as staging

WORKDIR /tmp
RUN apt-get update

# Install Google Chrome
RUN apt-get install -y curl libappindicator1 fonts-liberation
RUN curl -o chrome.deb https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb
RUN apt install -y /tmp/chrome.deb

FROM golang:1.16 as workspace

WORKDIR /go/src/app
COPY . .

RUN make deps
RUN make

# Bundle app
FROM staging as release
COPY --from=workspace /go/src/app/bin/imdb-ripper /usr/bin/imdb-ripper
RUN chmod +x /usr/bin/imdb-ripper

ENTRYPOINT ["/usr/bin/imdb-ripper"]
