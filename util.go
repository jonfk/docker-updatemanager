package main

import (
	"log"
        "github.com/fsouza/go-dockerclient"
)

func checkError(context string, err error) {
	if (err != nil) {
		log.Printf("Error occurred in : %v \n", context)
		log.Fatal(err)
	}
}

// Helper function to generate the port bindings with proper types
// Useful for general use case of binding a container port to a host port on
// default interface
func generatePortBindings(portBindings map[string]string) map[docker.Port][]docker.PortBinding {
	apiPortBindings := make(map[docker.Port][]docker.PortBinding)
	for key, value := range portBindings {
		apiPortBindings[(docker.Port)(key + "/tcp")] = []docker.PortBinding{docker.PortBinding{HostPort: value}}
	}
	return apiPortBindings

}
