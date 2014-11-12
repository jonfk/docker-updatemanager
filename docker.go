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
		Registry: "docker.inocybe.com",
		Tag: tag,
	}
	auth := docker.AuthConfiguration{}
	err := client.PullImage(options, auth)
	checkError("Pulling image", err)
}

// Creates and starts a container on the image name in name argument
// Currently has specific ports opened for Opendaylight Helium
// returns the container id of container started
func startContainer(client *docker.Client, name string) string {
	if DEBUG {
		log.Printf("Starting Container using image: %v\n", name)
	}
	createOptions := docker.CreateContainerOptions {
		Config: &docker.Config{
			Image: name,
		},
	}
	container, err := client.CreateContainer(createOptions)
	checkError("startContainer func: creating container", err)

	// Configure port bindings or publish all ports to random host ports
	portBindings := make(map[string]string)
	portBindings["162"] = "162" // SNMP4SDN only when started as root
	portBindings["179"] = "179" // BGP
	portBindings["1088"] = "1088" // JMX access
	portBindings["1790"] = "1790" // BGP/PCEP
	portBindings["1830"] = "1830" // Netconf
	portBindings["2400"] = "2400" // OSGi console
	portBindings["2550"] = "2550" // ODL Clustering
	portBindings["2551"] = "2551" // ODL Clustering
	portBindings["2552"] = "2552" // ODL Clustering
	portBindings["4189"] = "4189" // PCEP
	portBindings["4342"] = "4342" // Lips Flow Mapping
	portBindings["5005"] = "5005" // JConsole
	portBindings["5666"] = "5666" // ODL Internal clustering RPC
	portBindings["6633"] = "6633" // OpenFlow
	portBindings["6640"] = "6640" // OVSDB
	portBindings["6653"] = "6653" // OpenFlow
	portBindings["7800"] = "7800" // ODL Clustering
	portBindings["8000"] = "8000" // Java debug access
	portBindings["8080"] = "8080" // OpenDaylight web portal
	portBindings["8101"] = "8101" // KarafSSH
	portBindings["8181"] = "8181" // MD-SAL RESTCONF and DLUX
	portBindings["8383"] = "8383" // Netconf
	portBindings["12001"] = "12001" // ODL Clustering

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
