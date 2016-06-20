package main

import (
	log "github.com/Sirupsen/logrus"
	"fmt"
	"github.com/dghubble/sling"
	"net/http"
	"strings"
	"io/ioutil"
	"errors"
)

type SpectoLab struct {
	Host   string
	APIKey string
}

type SpectoLabSimulation struct {
	Version     string `json:"version"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (s *SpectoLab) CheckAPIKey() (error) {
	url := s.buildURL("/api/v1/simulations")
	request, err := sling.New().Post(url).BodyJSON("{}").Add("Authorization", s.buildAuthorizationHeaderValue()).Request()
	if err != nil {
		log.Debug(err.Error())
		return errors.New("Could not create a request to check API key against SpectoLab")
	}


	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Debug(err.Error())
		return errors.New("Could not communicate with SpectoLab")
	}

	log.Info(response.StatusCode)
	log.Info(s.buildAuthorizationHeaderValue())
	log.Info(url)
	if response.StatusCode == 401 {
		return errors.New("You don't have a valid API key, please sign in at https://lab.specto.io to generate a new API key")
	}

	return nil
}

func (s *SpectoLab) CreateSimulation(simulationName Simulation) (error) {
	simulation := SpectoLabSimulation{Version: simulationName.Version, Name: simulationName.Name, Description: "A description could go here"}

	url := s.buildURL("/api/v1/simulations")
	request, err := sling.New().Post(url).BodyJSON(simulation).Add("Authorization", s.buildAuthorizationHeaderValue()).Request()
	if err != nil {
		return err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}

	defer response.Body.Close()

	return nil
}

func (s *SpectoLab) UploadSimulation(simulation Simulation, data []byte) (bool, error) {
	err := s.CreateSimulation(simulation)

	if err != nil {
		return false, err
	}

	url := s.buildURL(fmt.Sprintf("/api/v1/users/%v/simulations/%v/versions/%v/data", simulation.Vendor,  simulation.Name, simulation.Version))

	request, err := sling.New().Put(url).Add("Authorization", s.buildAuthorizationHeaderValue()).Add("Content-Type", "application/json").Body(strings.NewReader(string(data))).Request()

	if err != nil {
		return false, err
	}

	response, err := http.DefaultClient.Do(request)

	if err != nil {
		return false, err
	}

	defer response.Body.Close()

	return response.StatusCode >= 200 && response.StatusCode <= 299, nil
}

func (s *SpectoLab) GetSimulation(simulation Simulation, overrideHost string) ([]byte, error) {
	var url string
	if len(overrideHost) > 0 {
		url = s.buildURL(fmt.Sprintf("/api/v1/users/%v/simulations/%v/versions/%v/data?override-host=%v", simulation.Vendor, simulation.Name, simulation.Version, overrideHost))
	} else {
		url = s.buildURL(fmt.Sprintf("/api/v1/users/%v/simulations/%v/versions/%v/data", simulation.Vendor, simulation.Name, simulation.Version))
	}

	request, err := sling.New().Get(url).Add("Authorization", s.buildAuthorizationHeaderValue()).Add("Content-Type", "application/json").Request()
	if err != nil {
		log.Debug(err.Error())
		return nil, errors.New("Could not create a request to the SpectoLab")
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Debug(err.Error())
		return nil, errors.New("Could not communicate with the SpectoLab")
	}

	defer response.Body.Close()

	if response.StatusCode != 200 {
		return nil, errors.New("Simulation not found in the SpectoLab")
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Debug(err.Error())
		return nil, errors.New("Could not pull simulation from the SpectoLab")
	}

	return body, nil
}

func (s *SpectoLab) buildURL(endpoint string) string {
	return fmt.Sprintf("%v%v", s.Host, endpoint)
}

func (s *SpectoLab) buildAuthorizationHeaderValue() string {
	return fmt.Sprintf("Bearer %v", s.APIKey)
}