FROM golang:1.23.3

WORKDIR /app
COPY . .
RUN make tidy

RUN export GOARCH=amd64
RUN export GOOS=linux

RUN make build/prod

CMD ["/bin/bash", "/tmp/bin/gpx-downloader"]
