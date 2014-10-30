package main

import (
        "github.com/fsouza/go-dockerclient"
	"log"
	"errors"
)

func printImages(client *docker.Client) {
        imgs, err := client.ListImages(false)
	checkError("list images", err)
        for _, img := range imgs {
                log.Println("ID: ", img.ID)
                log.Println("RepoTags: ", img.RepoTags)
                log.Println("Created: ", img.Created)
                log.Println("Size: ", img.Size)
                log.Println("VirtualSize: ", img.VirtualSize)
                log.Println("ParentId: ", img.ParentId)
		log.Println("")
        }
}

func findImageId(client *docker.Client, name string) (string, error) {
        imgs, err := client.ListImages(false)
	checkError("find image id for : " + name, err)
        for _, img := range imgs {
		for _, repoTag := range img.RepoTags {
			if repoTag == name {
				log.Printf("Found image %v with"+
					"tag %v\n", img.ID, name)
				if DEBUG {
					log.Println("Details: ")
					log.Println("ID: ", img.ID)
					log.Println("RepoTags: ", img.RepoTags)
					log.Println("Created: ", img.Created)
					log.Println("Size: ", img.Size)
					log.Println("VirtualSize: ", img.VirtualSize)
					log.Println("ParentId: ", img.ParentId)
					log.Println("")
				}
				return img.ID, nil
			}
		}
        }
	return "", errors.New("Cannot find image with tag: " + name)
}

func pullImage(client *docker.Client, repository, tag string) {
	if DEBUG {
		log.Printf("Pulling docker image: %v:%v", repository, tag)
	}
	options := docker.PullImageOptions{
		Repository: repository,
		Registry: "docker.jonfk.ca",
		Tag: tag,
	}
	auth := docker.AuthConfiguration{}
	err := client.PullImage(options, auth)
	checkError("Pulling image", err)
}

// Creates and starts a container on the image name in name argument
// returns the container id of container started
func startContainer(client *docker.Client, name string) string {
	if DEBUG {
		log.Printf("Starting Container using image: %v\n", name)
	}
	createOptions := docker.CreateContainerOptions {
		Name: "opendaylight",
		Config: &docker.Config{
			Image: name,
		},
	}
	container, err := client.CreateContainer(createOptions)
	checkError("startContainer func: creating container", err)

	// Configure port bindings or publish all ports to random host ports
	portBindings := make(map[string]string)
	portBindings["1088"] = "1088"
	portBindings["1830"] = "1830"
	portBindings["2400"] = "2400"
	portBindings["4342"] = "4342"
	portBindings["5666"] = "5666"
	portBindings["6633"] = "6633"
	portBindings["7800"] = "7800"
	portBindings["8000"] = "8000"
	portBindings["8080"] = "8080"
	portBindings["8383"] = "8383"
	portBindings["12001"] = "12001"

	hostConfig := docker.HostConfig{
		PortBindings: generatePortBindings(portBindings),
		PublishAllPorts: false,
	}
	client.StartContainer(container.ID, &hostConfig)
	log.Println("started container : " + container.ID)
	return container.ID
}

func findContainerId(client *docker.Client, imageName string) (string, error) {
	if DEBUG {
		log.Printf("Finding container ID for image: %v\n", imageName)
	}
	listContainerOptions := docker.ListContainersOptions{
		All: true,
	}
	containers, err:= client.ListContainers(listContainerOptions)
	checkError("func findContainerId: client.ListContainers", err)

	for _, container := range containers {
		if container.Image == imageName {
			if DEBUG {
				log.Println("ID: ", container.ID)
				log.Println("Image: ", container.Image)
				log.Println("Command: ", container.Command)
				log.Println("Created: ", container.Created)
				log.Println("Status: ", container.Status)
				log.Println("Ports: ", container.Ports)
				log.Println("SizeRw: ", container.SizeRw)
				log.Println("SizeRootFs: ", container.SizeRootFs)
				log.Println("Names: ", container.Names)
				log.Println("")
			}
			return container.ID, nil
		}
	}
	return "", errors.New("Cannot find container with image: " + imageName)
}

// Stops container started with image imageName
func stopContainer(client *docker.Client, imageName string) {
	if DEBUG {
		log.Printf("Stopping container with image name: \n", imageName)
	}
	containerId, err := findContainerId(client, imageName)
	checkError("func stopContainer: findContainerId for image: " + imageName, err)

	client.StopContainer(containerId, 0)
}

func removeContainer(client *docker.Client, imageName string) {
	if DEBUG {
		log.Printf("Removing container with image name: \n", imageName)
	}
	containerId, err := findContainerId(client, imageName)
	checkError("func removeContainer: findContainerId for image: " + imageName, err)

	removeContainerOptions := docker.RemoveContainerOptions{
		ID: containerId,
		RemoveVolumes: false,
		Force: false,
	}
	err = client.RemoveContainer(removeContainerOptions)
	checkError("func removeContainer: client.RemoveContainer", err)
}
