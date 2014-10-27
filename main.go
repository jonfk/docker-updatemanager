package main

import (
        "fmt"
        "github.com/fsouza/go-dockerclient"
	"github.com/coreos/go-omaha/omaha"
	"log"
	"errors"
	"encoding/xml"
	"net/http"
	"io/ioutil"
	"bytes"
)

var DEBUG bool = true

func main() {
        endpoint := "unix:///var/run/docker.sock"
        client, err := docker.NewClient(endpoint)
	checkError("creating docker client", err)

	printImages(client)
	//pullImage(client, "docker.jonfk.ca/opendaylight", "hydrogen")
	//findImageId(client, "docker.jonfk.ca/opendaylight:hydrogen")
	//startContainer(client, "docker.jonfk.ca/opendaylight:hydrogen")
	requestUpdate()
}

func printImages(client *docker.Client) {
        imgs, err := client.ListImages(false)
	checkError("list images", err)
        for _, img := range imgs {
                fmt.Println("ID: ", img.ID)
                fmt.Println("RepoTags: ", img.RepoTags)
                fmt.Println("Created: ", img.Created)
                fmt.Println("Size: ", img.Size)
                fmt.Println("VirtualSize: ", img.VirtualSize)
                fmt.Println("ParentId: ", img.ParentId)
		fmt.Println("")
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
	options := docker.PullImageOptions{
		Repository: repository,
		Registry: "docker.jonfk.ca",
		Tag: tag,
	}
	auth := docker.AuthConfiguration{}
	err := client.PullImage(options, auth)
	checkError("Pulling image", err)
}

func startContainer(client *docker.Client, name string) {
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
}

func requestUpdate() {
	osVersion := "testversion"
	osPlatform := "linux"
	osSp := ""
	osArch := "x64"

	appId := "aoeu"
	appVersion := "{1.2.3.4}"
	orequest := omaha.NewRequest(osVersion, osPlatform, osSp, osArch)
	//orequest.AddApp(appId, appVersion)
	oapp := omaha.NewApp(appId)
	oapp.Version = appVersion
	oapp.AddUpdateCheck()
	orequest.Apps = append(orequest.Apps, oapp)

	marshalledORequest, err := xml.MarshalIndent(orequest, "", "  ")
	checkError("func requestUpdate: marshalling xml", err)

	// send request
	url := ""
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(marshalledORequest))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	checkError("func requestUpdate: sending request", err)
	defer resp.Body.Close()

	if resp.Status == "200 OK" {
		body, err := ioutil.ReadAll(resp.Body)
		checkError("func requestUpdate: reading resp.Body", err)
		if DEBUG {
			log.Println("response Status:", resp.Status)
			log.Println("response Headers:", resp.Header)
			log.Println("response Body:", string(body))
		}

		var oresponse omaha.Response
		err = xml.Unmarshal(body, &oresponse)
		if DEBUG {
			log.Printf("Unmarshalled Body: %v\n", oresponse)
		}
	} else {
		log.Println("Error occurred in func requestUpdate")
		log.Println("http response returned: ", resp.Status)
		log.Println("response Headers:", resp.Header)
		body, _ := ioutil.ReadAll(resp.Body)
		log.Println("response Body:", string(body))
	}
}
