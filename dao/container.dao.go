package dao

import (
	"context"
	"fmt"
	"io"
	"time"

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
	prefix := configDao.ContainerPrefix
	if len(prefix) == 0 {
		prefix = "remote_terminal_default"
	}
	return fmt.Sprintf("%s-%s", prefix, ID)
}

func (c *ContainerDao) findByContainerID(containerID string) (*types.Container, error) {
	containers, err := c.cli.ContainerList(c.ctx, types.ContainerListOptions{
		All: true,
	})
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
	containers, err := c.cli.ContainerList(c.ctx, types.ContainerListOptions{
		All: true,
	})
	containerName := c.getContainerName(ID)
	if err != nil {
		return nil, err
	}
	for _, container := range containers {
		if len(container.Names) != 0 && container.Names[0] == fmt.Sprintf("/%s", containerName) {
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
		Cmd:          []string{CONTAINER_COMMAND},
		Image:        CONTAINER_IMAGE,
		Env:          []string{"TERM=xterm-256color"},
	}, nil, nil, nil, c.getContainerName(ID))
	if err != nil {
		return nil, err
	}
	container, err := c.findByContainerID(resp.ID)
	return container, err
}

func (c *ContainerDao) AttachAndWait(cont *types.Container, in io.Reader, out io.Writer, wsCloseChan chan interface{}, resizeChan chan [2]float64) error {
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
	go io.Copy(out, waiter.Reader)
	go io.Copy(waiter.Conn, in)

	statusCh, errCh := c.cli.ContainerWait(c.ctx, cont.ID, container.WaitConditionNotRunning)
	waiter.Conn.Write([]byte("\r"))
	for {
		select {
		case err := <-errCh:
			if err != nil {
				return err
			}
		case <-statusCh:
			return nil
		case <-wsCloseChan:
			return nil
		case size := <-resizeChan:
			c.Resize(cont, uint(size[0]), uint(size[1]))
		}
	}

}

func (c *ContainerDao) Resize(cont *types.Container, height uint, width uint) error {
	return c.cli.ContainerResize(c.ctx, cont.ID, types.ResizeOptions{
		Height: height,
		Width:  width,
	})
}

func (c *ContainerDao) Shutdown(cont *types.Container) error {
	timeout := time.Duration(1) * time.Second
	return c.cli.ContainerStop(c.ctx, cont.ID, &timeout)
}
