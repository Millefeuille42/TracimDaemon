package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

type tracimConfiguration struct {
	Url      string `json:"url"`
	Username string `json:"username"`
	Mail     string `json:"mail"`
	Password string `json:"password"`
}

type configuration struct {
	Tracim     tracimConfiguration `json:"tracim"`
	SocketPath string              `json:"socket_path"`
}

func createDirIfNotExist(path string) error {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.Mkdir(path, os.ModePerm)
		}
	}
	return err
}

func createFileIfNotExist(path string) error {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			_, err = os.Create(path)
		}
	}
	return err
}

func isExist(path string) (bool, error) {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func createDefaultConfig(path string) (*configuration, error) {
	defaultConfig := configuration{
		Tracim: tracimConfiguration{
			Url:      "http://localhost:8080/api",
			Username: "Me",
			Mail:     "me@example.com",
			Password: "S3crƎtP4s$woRd",
		},
		SocketPath: path + "master.sock",
	}

	configBytes, err := json.MarshalIndent(defaultConfig, "", "\t")
	if err != nil {
		return nil, err
	}

	err = createFileIfNotExist(path + "config.json")
	if err != nil {
		return nil, err
	}

	return &defaultConfig, os.WriteFile(path+"config.json", configBytes, os.ModePerm)
}

func getConfigDir() (string, error) {
	var dir = ""
	var err error = nil

	if len(os.Args) > 1 && !strings.HasPrefix(os.Args[1], "-p") {
		dir = os.Args[1]
	} else {
		dir, err = os.UserConfigDir()
		if err != nil {
			dir, err = os.UserHomeDir()
			if err != nil {
				return dir, err
			}
			dir = dir + "./config"
			err = createDirIfNotExist(dir)
		}
	}

	return dir, err
}

func readConfigFromFile() (*configuration, error) {
	dir, err := getConfigDir()
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("%s/TracimDaemon/", dir)
	configDir = path
	err = createDirIfNotExist(path)
	if err != nil {
		return nil, err
	}

	exist, err := isExist(path + "config.json")
	if err != nil {
		return nil, err
	}

	if !exist {
		return createDefaultConfig(path)
	}

	err = createDirIfNotExist(path + "log")
	if err != nil {
		return nil, err
	}

	configBytes, err := os.ReadFile(path + "config.json")
	if err != nil {
		return nil, err
	}

	conf := configuration{}
	err = json.Unmarshal(configBytes, &conf)
	return &conf, err
}

func setGlobalConfig() {
	newConf, err := readConfigFromFile()
	if err != nil {
		log.Fatal(err)
		return
	}
	globalConfig = *newConf
}

var globalConfig configuration
var configDir string
