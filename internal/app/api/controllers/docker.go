package controllers

import (
	"godopi/internal/app/api/models"
	"godopi/internal/app/configs"
	"godopi/internal/pkg/cache"
	"godopi/internal/pkg/docker"
	"net/http"
	"time"

	. "godopi/internal/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const ALL_CONTAINERS string = "ALL_CONTAINERS"

type DockerController struct {
	dockerClient docker.DockerClient
	cacheClient  cache.CacheClient
}

func NewDockerController() DockerController {
	Logger().Info("Constructing new docker controller..")

	redisAddress := configs.Config().GetString(configs.REDIS_ADDRESS)

	return DockerController{dockerClient: docker.NewDockerClient(), cacheClient: cache.NewCacheClient(redisAddress)}
}

// GetAllContainers godoc
// @Summary Gets all the running containers
// @Tags    Docker
// @Accept  json
// @Produce json
// @Success 200 {string} Status
// @Router  /docker/containers [get]
func (dc DockerController) GetAllContainers(ctx *gin.Context) {
	containersJson, err := dc.cacheClient.Get(ctx.Request.Context(), ALL_CONTAINERS)

	if err == cache.CacheNil {
		Logger().Info("Key does not exist in the cache storage", zap.String("Key", ALL_CONTAINERS))
	} else if err != nil {
		err = errors.Wrapf(err, "there is an error while getting the value from the cache storage. Key:%s", ALL_CONTAINERS)
		Logger().Error(err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"Message": "Error retrieving containers!", "Error": err.Error()})
		ctx.Abort()
		return
	} else {
		ctx.String(http.StatusOK, containersJson)
		return
	}

	containersJson, err = dc.dockerClient.GetAllContainersJson(ctx.Request.Context())

	if err != nil {
		err = errors.Wrap(err, "there is an error while getting the containers info")
		Logger().Error(err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"Message": "Error retrieving containers!", "Error": err.Error()})
		ctx.Abort()
		return
	}

	err = dc.cacheClient.Set(ctx.Request.Context(), ALL_CONTAINERS, containersJson, time.Minute)

	if err != nil {
		err = errors.Wrapf(err, "there is an error while setting a value into the cache storage. Value:%s", containersJson)
		Logger().Error(err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"Message": "Error retrieving containers!", "Error": err.Error()})
		ctx.Abort()
		return
	}

	ctx.String(http.StatusOK, containersJson)
}

// GetDetailedContainer godoc
// @Summary Gets detail for a container
// @Tags 	Docker
// @Accept 	json
// @Produce json
// @Param   id path string true "Container ID"
// @Success 200 {string} Status
// @Router 	/docker/containers/{id} [get]
func (dc DockerController) GetDetailedContainer(ctx *gin.Context) {
	containerId := ctx.Param("id")
	containerDetailJson, err := dc.dockerClient.GetDetailedContainerJson(ctx.Request.Context(), containerId)

	if err != nil {
		err = errors.Wrapf(err, "there is an error while getting detailed container info. ContainerId:%s", containerId)
		Logger().Error(err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"Message": "Error retrieving container detail!", "Error": err.Error()})
		ctx.Abort()
		return
	}

	ctx.String(http.StatusOK, containerDetailJson)
}

// CreateContainer godoc
// @Summary Creates a container by the given parameters
// @Tags 	Docker
// @Accept  json
// @Produce json
// @Param   Container body models.Container true "Create Container"
// @Success 201 {string} Status
// @Router  /docker/containers [post]
func (dc DockerController) CreateContainer(ctx *gin.Context) {
	var newContainer models.Container
	if err := ctx.BindJSON(&newContainer); err != nil {
		err = errors.Wrap(err, "there is an error while validating parameters of container")
		Logger().Error(err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"Message": "Error creating container!", "Error": err.Error()})
		ctx.Abort()
		return
	}

	containerId, err := dc.dockerClient.CreateContainer(ctx.Request.Context(), newContainer.ImageName, newContainer.ContainerName)

	if err != nil {
		err = errors.Wrap(err, "there is an error while creating container")
		Logger().Error(err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"Message": "Error creating container!", "Error": err.Error()})
		ctx.Abort()
		return
	}

	if newContainer.ContainerName != "" {
		ctx.JSON(http.StatusCreated, gin.H{"Success": "Container " + newContainer.ContainerName + " created from the image " + newContainer.ImageName + " with id: " + containerId})
	} else {
		ctx.JSON(http.StatusCreated, gin.H{"Success": "Container created from the image " + newContainer.ImageName + " with id: " + containerId})
	}
}

// DeleteContainer godoc
// @Summary Deletes a container
// @Tags 	Docker
// @Accept  json
// @Produce json
// @Param   id path string true "Container ID"
// @Success 200 {string} Status
// @Router  /docker/containers/{id} [delete]
func (dc DockerController) DeleteContainer(ctx *gin.Context) {
	containerId := ctx.Param("id")
	err := dc.dockerClient.DeleteContainer(ctx.Request.Context(), containerId)

	if err != nil {
		err = errors.Wrapf(err, "there is an error while deleting container. ContainerId:%s", containerId)
		Logger().Error(err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"Message": "Error deleting container!", "Error": err.Error()})
		ctx.Abort()
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"Success": "Container# " + containerId + " deleted"})
}
