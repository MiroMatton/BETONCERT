package main

import (
	"github.com/tkanos/gonfig"
)

type Configuration struct {
	User     string
	Password string
}

func GetConfig() Configuration {
	configuration := Configuration{}
	err := gonfig.GetConf("./config.json", &configuration)
	if err != nil {
		panic(err)
	}
	return configuration

}
