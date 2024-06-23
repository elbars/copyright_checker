package conf

import (
	"os"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v2"
)

type Pattern struct {
	Literal string `yaml:"literal"`
}

type Conf []struct {
	ID      string   `yaml:"id"`
	Message string   `yaml:"message"`
	Glob    []string `yaml:"glob" validate:"required"`
	Exclude []string `yaml:"exclude"`
	Pattern Pattern
	Vars    map[string]interface{} `yaml:"vars"`
}

func GetConf(configFileParh string) (*Conf, error) {

	var config *Conf
	yamlFile, err := os.ReadFile(configFileParh)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(yamlFile, &config)

	if err != nil {
		return nil, err
	}

	return config, nil
}

var validate *validator.Validate

func CheckConfFields(configStruct Conf) error {

	for _, c := range configStruct {

		validate = validator.New(validator.WithRequiredStructEnabled())
		err := validate.Struct(c)

		if err != nil {
			return err
		}

	}

	return nil

}
