FROM golang:1.17-alpine as builder
WORKDIR /app
COPY . /app
RUN go mod download
RUN go build -o /remote-terminal

FROM golang:1.17-alpine
ENV GIN_MODE=release
ENV PROXY=""
ENV CONFIG=""
ENV DOCKER_HOST="unix:///var/run/docker.sock"

WORKDIR /app
COPY --from=builder /remote-terminal /app/remote-terminal
COPY ./template /app/template

VOLUME [ "/var/run/docker.sock" ]

EXPOSE 8000

CMD /app/remote-terminal --bind :8000 --proxy ${PROXY} --config ${CONFIG}