package main

import (
	"log"
        "github.com/fsouza/go-dockerclient"
	"io/ioutil"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"
)

type UpdateClientConfig struct {
	// Update client configs
	UpdateServerURL string `json`
	DockerEndpoint string `json`
	UpdateDelayInMinutes int `json`
	Debug bool `json`

	// app configs
	OSVersion string `json`
	OSPlatform string `json`
	OSSp string `json`
	OSArch string `json`
	AppId string `json`
	AppMachineID string `json`
	// Will be updated on update check response
	AppVersion string `json`
	AppPackageName string `json`

	// docker configs (runtime)
	DockerImageName string `json`
	DockerContainerId string `json`

}

// Handles SIGTERM and SIGINT signal
// Tries to exit gracefully and stops all running docker containers
func handleKill() {

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Block until a signal is received.
	s := <-c
	log.Println("Received signal:", s)

	// Exit Logic
        dockerClient, err := docker.NewClient(DOCKER_ENDPOINT)
	checkError("handleKill func: creating docker client", err)
	stopContainerFromImage(dockerClient, CONFIG.DockerImageName)

	code := 0
	if DEBUG {
		log.Printf("Exiting with error code: %v\n", code)
	}
	os.Exit(code)
}

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

func readConfig() UpdateClientConfig {
	if DEBUG {
		log.Println("Reading config file from: ", CONFIG_FILE)
	}
	data, err := ioutil.ReadFile(CONFIG_FILE)
	checkError("func readConfig: ioutil.Readfile: " + CONFIG_FILE, err)

	var config UpdateClientConfig
	err = json.Unmarshal(data, &config)
	if DEBUG {
		log.Printf("Config: %#v\n", config)
	}
	return config
}

func writeConfig(config UpdateClientConfig) {
	if DEBUG {
		log.Println("Writing config file to: ", CONFIG_FILE)
	}
	data, err := json.MarshalIndent(config, "", "  ")
	if DEBUG {
		log.Printf("Config: %#v\n", config)
	}
	checkError("func writeConfig: json.MarshalIndent: ", err)
	ioutil.WriteFile(CONFIG_FILE, data, 0664)
}

// DEBUG functions

func writeDebugConfig() {
	config := UpdateClientConfig{
		// Update client configs
		UpdateServerURL:"http://192.168.1.112:8080/v1/update",
		DockerEndpoint: "unix:///var/run/docker.sock",
		UpdateDelayInMinutes: 60,
		Debug: true,
		// app configs
		OSVersion: "MyOSVersion",
		OSPlatform: "coreos",
		OSSp: "",
		OSArch: "x86_64",
		AppId: "aoeu",
		AppMachineID: "MyMachineJonfk",
		AppVersion: "{1.2.3.4}",
		AppPackageName:"opendaylight:hydrogen",

		DockerImageName: "",
		DockerContainerId: "",
	}
	writeConfig(config)
}
