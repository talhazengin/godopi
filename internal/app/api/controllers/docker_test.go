package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"godopi/internal/app/api/models"
	"godopi/internal/pkg/cache"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gotest.tools/v3/assert"
)

type mockCacheClient struct {
	MockGet func(ctx context.Context, key string) (string, error)
	MockSet func(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	MockDel func(ctx context.Context, keys ...string) error
}

func (mcc *mockCacheClient) Get(ctx context.Context, key string) (string, error) {
	return mcc.MockGet(ctx, key)
}
func (mcc *mockCacheClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return mcc.MockSet(ctx, key, value, expiration)
}
func (mcc *mockCacheClient) Del(ctx context.Context, keys ...string) error {
	return mcc.MockDel(ctx, keys...)
}

type mockDockerClient struct {
	MockGetAllContainersJson     func(c context.Context) (string, error)
	MockGetDetailedContainerJson func(c context.Context, containerId string) (string, error)
	MockCreateContainer          func(c context.Context, imageName string, containerName string) (string, error)
	MockDeleteContainer          func(c context.Context, containerId string) error
}

func (mdc *mockDockerClient) GetAllContainersJson(ctx context.Context) (string, error) {
	return mdc.MockGetAllContainersJson(ctx)
}
func (mdc *mockDockerClient) GetDetailedContainerJson(ctx context.Context, containerId string) (string, error) {
	return mdc.MockGetDetailedContainerJson(ctx, containerId)
}
func (mdc *mockDockerClient) CreateContainer(ctx context.Context, imageName string, containerName string) (string, error) {
	return mdc.MockCreateContainer(ctx, imageName, containerName)
}
func (mdc *mockDockerClient) DeleteContainer(ctx context.Context, containerId string) error {
	return mdc.MockDeleteContainer(ctx, containerId)
}

func TestGetAllContainersSuccessCaching(t *testing.T) {
	w := httptest.NewRecorder()
	c, e := gin.CreateTestContext(w)

	mockCacheClient := mockCacheClient{}
	mockCacheClient.MockGet = func(ctx context.Context, key string) (string, error) {
		return "cachedContainersMetadataTest", nil
	}

	dockerController := DockerController{dockerClient: &mockDockerClient{}, cacheClient: &mockCacheClient}

	e.GET("/", dockerController.GetAllContainers)
	c.Request, _ = http.NewRequestWithContext(c, http.MethodGet, "/", nil)
	e.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "cachedContainersMetadataTest", w.Body.String())
}

func TestGetAllContainersErrorDockerClient(t *testing.T) {
	w := httptest.NewRecorder()
	c, e := gin.CreateTestContext(w)

	mockCacheClient := mockCacheClient{}
	mockCacheClient.MockGet = func(ctx context.Context, key string) (string, error) {
		return "", cache.CacheNil
	}

	errorMessage := "could not get any container info"

	mockDockerClient := mockDockerClient{}
	mockDockerClient.MockGetAllContainersJson = func(c context.Context) (string, error) {
		return "", errors.New(errorMessage)
	}

	dockerController := DockerController{dockerClient: &mockDockerClient, cacheClient: &mockCacheClient}

	e.GET("/", dockerController.GetAllContainers)
	c.Request, _ = http.NewRequestWithContext(c, http.MethodGet, "/", nil)
	e.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, true, strings.Contains(w.Body.String(), errorMessage))
}

func TestGetDetailedContainerSuccessDockerClient(t *testing.T) {
	w := httptest.NewRecorder()
	c, e := gin.CreateTestContext(w)

	containerDetailedInfoTest := "containerDetailedInfoTest"

	mockDockerClient := mockDockerClient{}
	mockDockerClient.MockGetDetailedContainerJson = func(c context.Context, containerId string) (string, error) {
		return containerDetailedInfoTest, nil
	}

	dockerController := DockerController{dockerClient: &mockDockerClient, cacheClient: &mockCacheClient{}}

	e.GET("/:id", dockerController.GetDetailedContainer)

	containerId := "3423ASDF372FA7DF732"
	c.Request, _ = http.NewRequestWithContext(c, http.MethodGet, "/"+containerId, nil)
	e.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, containerDetailedInfoTest, w.Body.String())
}

func TestGetDetailedContainerErrorDockerClient(t *testing.T) {
	w := httptest.NewRecorder()
	c, e := gin.CreateTestContext(w)

	errorMessage := "could not get detailed container info"

	mockDockerClient := mockDockerClient{}
	mockDockerClient.MockGetDetailedContainerJson = func(c context.Context, containerId string) (string, error) {
		return "", errors.New(errorMessage)
	}

	dockerController := DockerController{dockerClient: &mockDockerClient, cacheClient: &mockCacheClient{}}

	e.GET("/:id", dockerController.GetDetailedContainer)

	containerId := "3423ASDF372FA7DF732"
	c.Request, _ = http.NewRequestWithContext(c, http.MethodGet, "/"+containerId, nil)
	e.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, true, strings.Contains(w.Body.String(), errorMessage))
}

func TestCreateContainerSuccessDockerClient(t *testing.T) {
	w := httptest.NewRecorder()
	c, e := gin.CreateTestContext(w)

	createdContainerId := "3423ASDF372FA7DF732"

	mockDockerClient := mockDockerClient{}
	mockDockerClient.MockCreateContainer = func(c context.Context, imageName, containerName string) (string, error) {
		return createdContainerId, nil
	}

	dockerController := DockerController{dockerClient: &mockDockerClient, cacheClient: &mockCacheClient{}}

	e.POST("/", dockerController.CreateContainer)

	container := models.Container{
		ImageName:     "testImageName",
		ContainerName: "testContainerName",
	}

	data, _ := json.Marshal(container)
	c.Request, _ = http.NewRequestWithContext(c, http.MethodPost, "/", bytes.NewBuffer(data))
	e.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, true, strings.Contains(w.Body.String(), createdContainerId))
	assert.Equal(t, true, strings.Contains(w.Body.String(), container.ContainerName))
	assert.Equal(t, true, strings.Contains(w.Body.String(), container.ImageName))
}

func TestCreateContainerErrorDockerClient(t *testing.T) {
	w := httptest.NewRecorder()
	c, e := gin.CreateTestContext(w)

	errorMessage := "could not create container"

	mockDockerClient := mockDockerClient{}
	mockDockerClient.MockCreateContainer = func(c context.Context, imageName, containerName string) (string, error) {
		return "", errors.New(errorMessage)
	}

	dockerController := DockerController{dockerClient: &mockDockerClient, cacheClient: &mockCacheClient{}}

	e.POST("/", dockerController.CreateContainer)

	container := models.Container{
		ImageName:     "testImageName",
		ContainerName: "testContainerName",
	}

	data, _ := json.Marshal(container)
	c.Request, _ = http.NewRequestWithContext(c, http.MethodPost, "/", bytes.NewBuffer(data))
	e.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, true, strings.Contains(w.Body.String(), errorMessage))
}

func TestDeleteContainerSuccessDockerClient(t *testing.T) {
	w := httptest.NewRecorder()
	c, e := gin.CreateTestContext(w)

	deletedContainerId := "3423ASDF372FA7DF732"

	mockDockerClient := mockDockerClient{}
	mockDockerClient.MockDeleteContainer = func(c context.Context, containerId string) error {
		return nil
	}

	dockerController := DockerController{dockerClient: &mockDockerClient, cacheClient: &mockCacheClient{}}

	e.DELETE("/:id", dockerController.DeleteContainer)

	c.Request, _ = http.NewRequestWithContext(c, http.MethodDelete, "/"+deletedContainerId, nil)
	e.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, true, strings.Contains(w.Body.String(), deletedContainerId))
}

func TestDeleteContainerErrorDockerClient(t *testing.T) {
	w := httptest.NewRecorder()
	c, e := gin.CreateTestContext(w)

	deletedContainerId := "3423ASDF372FA7DF732"
	errorMessage := "could not delete the container"

	mockDockerClient := mockDockerClient{}
	mockDockerClient.MockDeleteContainer = func(c context.Context, containerId string) error {
		return errors.New(errorMessage)
	}

	dockerController := DockerController{dockerClient: &mockDockerClient, cacheClient: &mockCacheClient{}}

	e.DELETE("/:id", dockerController.DeleteContainer)

	c.Request, _ = http.NewRequestWithContext(c, http.MethodDelete, "/"+deletedContainerId, nil)
	e.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, true, strings.Contains(w.Body.String(), deletedContainerId))
	assert.Equal(t, true, strings.Contains(w.Body.String(), errorMessage))
}
