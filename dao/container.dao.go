package dao

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

const (
	CONTAINER_IMAGE   = "debian:11"
	CONTAINER_COMMAND = "/bin/bash"
)

type ContainerDao struct {
	ctx context.Context
	cli *client.Client
}

func NewContainerDao() (*ContainerDao, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &ContainerDao{
		ctx: context.Background(),
		cli: cli,
	}, nil
}

func (c *ContainerDao) getContainerName(ID string) string {
	configDao := NewConfigDaoMust()
	return fmt.Sprintf("%s-%s", configDao.ContainerPrefix, ID)
}

func (c *ContainerDao) findByContainerID(containerID string) (*types.Container, error) {
	containers, err := c.cli.ContainerList(c.ctx, types.ContainerListOptions{})
	if err != nil {
		return nil, err
	}
	for _, container := range containers {
		if container.ID == containerID {
			return &container, nil
		}
	}
	return nil, nil
}

func (c *ContainerDao) FindByID(ID string) (*types.Container, error) {
	containers, err := c.cli.ContainerList(c.ctx, types.ContainerListOptions{})
	containerName := c.getContainerName(ID)
	if err != nil {
		return nil, err
	}
	for _, container := range containers {
		if len(container.Names) != 0 && container.Names[0] == containerName {
			return &container, nil
		}
	}
	return nil, nil
}

func (c *ContainerDao) CreateByID(ID string, out io.Writer) (*types.Container, error) {
	reader, err := c.cli.ImagePull(c.ctx, CONTAINER_IMAGE, types.ImagePullOptions{})
	if err != nil {
		return nil, err
	}
	io.Copy(out, reader)
	resp, err := c.cli.ContainerCreate(c.ctx, &container.Config{
		AttachStderr: true,
		AttachStdin:  true,
		Tty:          true,
		AttachStdout: true,
		OpenStdin:    true,
		Cmd:          []string{"/bin/bash", "-"},
		Image:        "debian:11",
		Env:          []string{"TERM=xterm-256color"},
	}, nil, nil, nil, c.getContainerName(ID))
	if err != nil {
		return nil, err
	}
	container, err := c.findByContainerID(resp.ID)
	return container, err
}

func (c *ContainerDao) AttachAndWaitByID(cont *types.Container, conn net.Conn) error {
	// TODO attach stdin stdout stderr
	out := make(chan []byte)
	if err := c.cli.ContainerStart(c.ctx, cont.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}
	waiter, err := c.cli.ContainerAttach(c.ctx, cont.ID, types.ContainerAttachOptions{
		Stdin:  true,
		Stdout: true,
		Stderr: true,
		Stream: true,
	})
	if err != nil {
		return err
	}
	go io.Copy(conn, waiter.Reader)
	go func() {
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			out <- []byte(scanner.Text())
		}
	}()

	go func(w io.WriteCloser) {
		for {
			data, ok := <-out
			if !ok {
				w.Close()
				return
			}
			w.Write(data)
		}
	}(waiter.Conn)

	statusCh, errCh := c.cli.ContainerWait(c.ctx, cont.ID, container.WaitConditionNotRunning)

	select {
	case err := <-errCh:
		if err != nil {
			return err
		}
	case <-statusCh:
	}

	return nil
}
