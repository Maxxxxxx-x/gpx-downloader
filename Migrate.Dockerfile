FROM golang:1.23.3

WORKDIR /app
COPY . .
RUN make tidy

RUN export GOARCH=amd64
RUN export GOOS=linux

RUN make build/migrate

CMD ["/bin/bash", "/tmp/bin/migrate"]
