package docker

import (
	"context"
	"encoding/json"
	. "godopi/internal/pkg/logger"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type DockerClient interface {
	GetAllContainersJson(ctx context.Context) (string, error)
	GetDetailedContainerJson(ctx context.Context, containerId string) (string, error)
	CreateContainer(ctx context.Context, imageName string, containerName string) (string, error)
	DeleteContainer(ctx context.Context, containerId string) error
}

type dockerClient struct {
	client *client.Client
}

func NewDockerClient() DockerClient {
	Logger().Info("Constructing new docker client..")

	client, err := client.NewClientWithOpts(client.FromEnv)

	if err != nil {
		Logger().Fatal("Encountered an error while initializing the Docker engine client! Error:", zap.Error(err))
	}

	return dockerClient{client: client}
}

func (dc dockerClient) GetAllContainersJson(ctx context.Context) (string, error) {
	containers, err := dc.client.ContainerList(ctx, types.ContainerListOptions{})

	if err != nil {
		return "", errors.Wrap(err, "there is an error while requesting container list through docker client")
	}

	byteData, err := json.MarshalIndent(containers, "", "")

	if err != nil {
		return "", errors.Wrap(err, "there is an error while marshalling container list into json")
	}

	return string(byteData), nil
}

func (dc dockerClient) GetDetailedContainerJson(ctx context.Context, containerId string) (string, error) {
	containerDetail, err := dc.client.ContainerInspect(ctx, containerId)

	if err != nil {
		return "", errors.Wrapf(err, "there is an error while requesting container inspect through docker client. ContainerId:%s", containerId)
	}

	byteData, err := json.MarshalIndent(containerDetail, "", "")

	if err != nil {
		return "", errors.Wrap(err, "there is an error while marshalling container detail into json")
	}

	return string(byteData), nil
}

func (dc dockerClient) CreateContainer(ctx context.Context, imageName string, containerName string) (string, error) {
	ioReadCloser, err := dc.client.ImagePull(ctx, imageName, types.ImagePullOptions{})

	if err != nil {
		return "", errors.Wrapf(err, "there is an error while pulling the image. ImageName:%s", imageName)
	}

	defer ioReadCloser.Close()

	_, err = io.Copy(os.Stdout, ioReadCloser)

	if err != nil {
		return "", errors.Wrapf(err, "there is an error while copying the image data. ImageName:%s", imageName)
	}

	containerConfig := &container.Config{
		Image: imageName,
	}

	container, err := dc.client.ContainerCreate(ctx, containerConfig, nil, nil, nil, containerName)

	if err != nil {
		return "", errors.Wrapf(err, "there is an error while requesting container create through docker client. ImageName:%s", imageName)
	}

	if err = dc.client.ContainerStart(ctx, container.ID, types.ContainerStartOptions{}); err != nil {
		return "", errors.Wrapf(err, "there is an error while requesting container start through docker client. ContainerId:%s", container.ID)
	}

	return container.ID, nil
}

func (dc dockerClient) DeleteContainer(ctx context.Context, containerId string) error {
	err := dc.client.ContainerRemove(ctx, containerId, types.ContainerRemoveOptions{RemoveVolumes: false, RemoveLinks: false, Force: true})

	if err != nil {
		return errors.Wrapf(err, "there is an error while requesting container delete through docker client. ContainerId:%s", containerId)
	}

	return nil
}
