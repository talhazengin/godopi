package models

type Container struct {
	ImageName     string `json:"imageName" binding:"required"`
	ContainerName string `json:"containerName"`
}
