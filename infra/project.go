package infra

import (
	"fmt"
	"github.com/goccy/go-yaml"
	"github.com/myuon/probable-chainsaw/model"
	"github.com/pkg/errors"
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
		return model.Project{}, errors.WithStack(err)
	}

	var project model.Project
	if err := yaml.Unmarshal(bin, &project); err != nil {
		return model.Project{}, errors.WithStack(err)
	}

	return project, nil
}

func SaveProject(configFilePath string, project model.Project) error {
	bin, err := yaml.Marshal(&project)
	if err != nil {
		return errors.WithStack(err)
	}

	if err := os.WriteFile(configFilePath, bin, 0644); err != nil {
		return errors.WithStack(err)
	}

	return nil
}
