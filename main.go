package main

import (
        "github.com/fsouza/go-dockerclient"
	"github.com/coreos/go-omaha/omaha"
	"log"
	"encoding/xml"
	"net/http"
	"io/ioutil"
	"bytes"
	"strings"
	"time"
	"sync"
)

var CONFIG_FILE = "./config.json"
//var CONFIG_FILE = "/usr/share/NOS-update-client/config.json"

// Defaults to be set from config file
var DEBUG bool = true
var DOCKER_ENDPOINT string = "unix:///var/run/docker.sock"
var UPDATE_SERVER = "http://192.168.1.112:8080/v1/update"

var CONFIG UpdateClientConfig

var WAITGROUP sync.WaitGroup

func init() {
	//writeDebugConfig()
	// Set configs
	CONFIG = readConfig()
	if DEBUG {
		log.Printf("CONFIG : %#v", CONFIG)
	}
	DEBUG = CONFIG.Debug
	DOCKER_ENDPOINT = CONFIG.DockerEndpoint
	UPDATE_SERVER = CONFIG.UpdateServerURL

}

func main() {

        //dockerClient, err := docker.NewClient(DOCKER_ENDPOINT)
	//checkError("creating docker client", err)

	//printImages(dockerClient)
	//pullImage(dockerClient, "docker.jonfk.ca/opendaylight", "hydrogen")
	//findImageId(dockerClient, "docker.jonfk.ca/opendaylight:hydrogen")
	//startContainer(dockerClient, "docker.jonfk.ca/opendaylight:hydrogen")
	//findContainerId(dockerClient, "docker.jonfk.ca/opendaylight:hydrogen")
	//stopContainer(dockerClient, "docker.jonfk.ca/opendaylight:hydrogen")
	//removeContainer(dockerClient, "docker.jonfk.ca/opendaylight:hydrogen")

	requestUpdate()
	scheduleUpdateRequest(time.Duration(CONFIG.UpdateDelayInMinutes) * time.Minute)
	WAITGROUP.Wait()
}

func requestUpdate() {

	if DEBUG {
		log.Println("Requesting update")
	}

	log.Printf("CONFIG : %#v", CONFIG)

	osVersion := CONFIG.OSVersion
	osPlatform := CONFIG.OSPlatform
	osSp := CONFIG.OSSp
	osArch := CONFIG.OSArch

	appId := CONFIG.AppId
	appVersion := CONFIG.AppVersion
	appMachineID := CONFIG.AppMachineID

	orequest := omaha.NewRequest(osVersion, osPlatform, osSp, osArch)
	oapp := omaha.NewApp(appId)
	oapp.Version = appVersion
	oapp.MachineID = appMachineID
	oapp.AddUpdateCheck()
	orequest.Apps = append(orequest.Apps, oapp)

	marshalledORequest, err := xml.MarshalIndent(orequest, "", "  ")
	checkError("func requestUpdate: marshalling xml", err)

	// send request
	url := UPDATE_SERVER
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(marshalledORequest))
	req.Header.Set("Content-Type", "application/xml")

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
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
			log.Printf("Unmarshalled Body: %#v\n", oresponse)
		}
		checkError("func requestUpdate: unmarshalling response body", err)

		// React to response
		reactToOmahaResponse(&oresponse)

	} else {
		log.Println("Error occurred in func requestUpdate")
		log.Println("http response returned: ", resp.Status)
		log.Println("response Headers:", resp.Header)
		body, _ := ioutil.ReadAll(resp.Body)
		log.Println("response Body:", string(body))
	}
}

func reactToOmahaResponse(oresponse *omaha.Response) {
	if DEBUG {
		log.Println("Reacting to omaha response")
	}
        dockerClient, err := docker.NewClient(DOCKER_ENDPOINT)
	checkError("creating docker client", err)
	for _, orApp := range oresponse.Apps {
		// Write new version?
		newAppVersion := orApp.UpdateCheck.Manifest.Version
		for _, orPackage := range orApp.UpdateCheck.Manifest.Packages.Packages {
			newPackageName := orPackage.Name
			if DEBUG {
				log.Printf("Package Name: %v\n", newPackageName)
			}
			splitName := strings.Split(newPackageName, ":")
			dockerPackageName := splitName[0]
			dockerPackageTag := splitName[1]
			log.Printf("Docker Package Name received: %v, %v\n", dockerPackageName, dockerPackageTag)

			//ADD LOGIC stop containers, remove containers and add new container

			if CONFIG.DockerContainerId != "" {
				stopContainer(dockerClient, CONFIG.DockerImageName)
			}

			newDockerImageName := "docker.jonfk.ca/"+dockerPackageName

			pullImage(dockerClient, newDockerImageName, dockerPackageTag)
			newContainerId := startContainer(dockerClient, "docker.jonfk.ca/"+newPackageName)

			// Update CONFIG
			CONFIG.AppVersion = newAppVersion
			CONFIG.AppPackageName = newPackageName
			CONFIG.DockerImageName = newDockerImageName
			CONFIG.DockerContainerId = newContainerId
			writeConfig(CONFIG)
		}
	}
}

func scheduleUpdateRequest(delay time.Duration) chan struct{} {
	ticker := time.NewTicker(delay)
	quit := make(chan struct{})
	WAITGROUP.Add(1)

	go func() {
		for{
			select{
			case <- ticker.C:
				requestUpdate()
			case <- quit:
				ticker.Stop()
				WAITGROUP.Done()
				return
			}
		}
	}()
	return quit
}
