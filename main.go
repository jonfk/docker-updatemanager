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
)

var CONFIG_FILE = "./config.json"

var DEBUG bool = true
var DOCKER_ENDPOINT string = "unix:///var/run/docker.sock"
var UPDATE_SERVER = "http://192.168.1.112:8080/v1/update"

var CONFIG UpdateClientConfig

func main() {
	// Set configs
	CONFIG := readConfig()
	log.Printf("CONFIG : %#v", CONFIG)

        dockerClient, err := docker.NewClient(DOCKER_ENDPOINT)
	checkError("creating docker client", err)

	printImages(dockerClient)
	//pullImage(dockerClient, "docker.jonfk.ca/opendaylight", "hydrogen")
	//findImageId(dockerClient, "docker.jonfk.ca/opendaylight:hydrogen")
	//startContainer(dockerClient, "docker.jonfk.ca/opendaylight:hydrogen")
	//requestUpdate(dockerClient)
	//findContainerId(dockerClient, "docker.jonfk.ca/opendaylight:hydrogen")
	//stopContainer(dockerClient, "docker.jonfk.ca/opendaylight:hydrogen")
	//removeContainer(dockerClient, "docker.jonfk.ca/opendaylight:hydrogen")

	/*
	config := UpdateClientConfig{
		// Update client configs
		UpdateServerURL:"http://192.168.1.112:8080/v1/update",
		DockerEndpoint: "unix:///var/run/docker.sock",
		Debug: true,
		// app configs
		OSVersion: "MyOSVersion",
		OSPlatform: "coreos",
		OSSp: "",
		AppId: "aoeu",
		AppVersion: "{1.2.3.4}",
		AppPackageName:"opendaylight:hydrogen",
	}
*/
	//writeConfig(config)
}

func requestUpdate(dockerClient *docker.Client) {
	osVersion := "testversion"
	osPlatform := "coreos"
	osSp := ""
	osArch := "x86_64"

	appId := "aoeu"
	appVersion := "{1.2.3.4}"
	appMachineID := "MyMachineJonfk"

	orequest := omaha.NewRequest(osVersion, osPlatform, osSp, osArch)
	//orequest.AddApp(appId, appVersion)
	oapp := omaha.NewApp(appId)
	oapp.Version = appVersion
	oapp.MachineID = appMachineID
	oapp.AddUpdateCheck()
	orequest.Apps = append(orequest.Apps, oapp)

	marshalledORequest, err := xml.MarshalIndent(orequest, "", "  ")
	checkError("func requestUpdate: marshalling xml", err)

	// send request
	//url := "http://requestb.in/18tdpae1"
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
        dockerClient, err := docker.NewClient(DOCKER_ENDPOINT)
	checkError("creating docker client", err)
	for _, orApp := range oresponse.Apps {
		//newAppVersion := orApp.UpdateCheck.Manifest.Version
		for _, orPackage := range orApp.UpdateCheck.Manifest.Packages.Packages {
			newPackageName := orPackage.Name
			if DEBUG {
				log.Printf("Package Name: %v\n", newPackageName)
			}
			splitName := strings.Split(newPackageName, ":")
			dockerPackageName := splitName[0]
			dockerPackageTag := splitName[1]
			log.Printf("Docker Package Name received: %v, %v\n", dockerPackageName, dockerPackageTag)

			pullImage(dockerClient, "docker.jonfk.ca/"+dockerPackageName, dockerPackageTag)
			startContainer(dockerClient, "docker.jonfk.ca/"+newPackageName)
		}
	}
}
