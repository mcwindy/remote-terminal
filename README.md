# remote-terminal

remote terminal

![example](https://raw.githubusercontent.com/ChenKS12138/remote-terminal/main/docs/example.png)

## How To Run

```shell
docker pull chenks/remote-terminal:latest
docker run -d --name local-remote-terminal --restart always --add-host=host.docker.internal:host-gateway -e GIN_MODE=release -e PROXY=http://host.docker.internal:1087 -e CONFIG=https://raw.githubusercontent.com/ChenKS12138/remote-terminal/main/example.config.yml -p 8000:8000 -v /var/run/docker.sock:/var/run/docker.sock chenks/remote-terminal:latest
```

## Configuration

see https://github.com/ChenKS12138/remote-terminal/blob/main/example.config.yml
