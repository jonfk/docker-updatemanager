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
	"os/exec"
)

//var CONFIG_FILE = "./config.json"
var CONFIG_FILE = "/etc/NOS-update-client/config.json"

// Defaults to be set from config file
var DEBUG bool = true
var DOCKER_ENDPOINT string = "unix:///var/run/docker.sock"
var UPDATE_SERVER = "http://update.inocybe.com:8080/v1/update"

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

	if CONFIG.AppMachineID == "" {
		out, err := exec.Command("uuidgen").Output()
		checkError("func init: generating uuid", err)

		CONFIG.AppMachineID = strings.Trim(string(out), "\n")
		writeConfig(CONFIG)
	}

	if CONFIG.OSVersion == "" || CONFIG.OSPlatform == "" {
		out, err := ioutil.ReadFile("/etc/os-release")
		checkError("func init: reading /etc/os-release", err)
		for _, line := range strings.Split(string(out), "\n") {
			sLine := strings.Split(line, "=")
			if sLine[0] == "NAME" {
				CONFIG.OSPlatform = strings.Trim(sLine[1], "\"")
			} else if sLine[0] == "VERSION" {
				CONFIG.OSVersion = strings.Trim(sLine[1], "\"")
			}
		}
		writeConfig(CONFIG)
	}

}

func main() {

        //dockerClient, err := docker.NewClient(DOCKER_ENDPOINT)
	//checkError("creating docker client", err)

	//printImages(dockerClient)
	//pullImage(dockerClient, "docker.inocybe.com/opendaylight", "helium")
	//findImageId(dockerClient, "docker.inocybe.com/opendaylight:hydrogen")
	//startContainerFromImage(dockerClient, "docker.jonfk.ca/opendaylight:hydrogen")
	//findContainerId(dockerClient, "docker.jonfk.ca/opendaylight:hydrogen")
	//stopContainerFromImage(dockerClient, "docker.jonfk.ca/opendaylight:hydrogen")
	//removeContainer(dockerClient, "docker.jonfk.ca/opendaylight:hydrogen")

	go handleKill()

	startupRun()
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
				stopContainerFromImage(dockerClient, CONFIG.DockerImageName)
				removeContainer(dockerClient, CONFIG.DockerImageName)
				CONFIG.DockerContainerId = ""
				CONFIG.DockerImageName = ""
			}

			newDockerImageName := "docker.inocybe.com/"+dockerPackageName

			pullImage(dockerClient, newDockerImageName, dockerPackageTag)
			newContainerId := startContainerFromImage(dockerClient, "docker.inocybe.com/"+newPackageName)

			// Update CONFIG
			CONFIG.AppVersion = newAppVersion
			CONFIG.AppPackageName = newPackageName
			CONFIG.DockerImageName = newDockerImageName
			CONFIG.DockerContainerId = newContainerId
			writeConfig(CONFIG)
		}
	}
}

// Runs on every start of the program
// starts up the docker container it's running if there is one
// else pull the opendaylight image and start the container from the image
func startupRun() {
        dockerClient, err := docker.NewClient(DOCKER_ENDPOINT)
	checkError("func startupRun: creating docker client", err)

	if CONFIG.DockerContainerId != "" {
		startContainer(dockerClient, CONFIG.DockerContainerId)
		return
	}

	dockerImageName := "docker.inocybe.com/opendaylight"
	dockerPackageTag := "helium"

	pullImage(dockerClient, dockerImageName, dockerPackageTag)
	newContainerId := startContainerFromImage(dockerClient, dockerImageName+":"+dockerPackageTag)

	// Update CONFIG
	CONFIG.DockerImageName = dockerImageName + ":" + dockerPackageTag
	CONFIG.DockerContainerId = newContainerId
	writeConfig(CONFIG)
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
