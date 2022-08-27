package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/myuon/probable-chainsaw/model"
	"os"
)

func Exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func LoadProject(configFilePath string) (model.Project, error) {
	if !Exists(configFilePath) {
		return model.Project{}, errors.New(fmt.Sprintf("%v does not exist", configFilePath))
	}

	bin, err := os.ReadFile(configFilePath)
	if err != nil {
		return model.Project{}, err
	}

	var project model.Project
	if err := json.Unmarshal(bin, &project); err != nil {
		return model.Project{}, err
	}

	return project, nil
}

func SaveProject(configFilePath string, project model.Project) error {
	bin, err := json.Marshal(&project)
	if err != nil {
		return err
	}

	if err := os.WriteFile(configFilePath, bin, 0644); err != nil {
		return err
	}

	return nil
}
