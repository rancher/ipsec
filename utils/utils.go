package utils

import (
	"github.com/rancher/go-rancher-metadata/metadata"
)

// IsContainerConsideredRunning function is used to test if the container is in any of
// the states that are considered running.
func IsContainerConsideredRunning(aContainer metadata.Container) bool {
	return (aContainer.State == "running" || aContainer.State == "starting" || aContainer.State == "stopping")
}
