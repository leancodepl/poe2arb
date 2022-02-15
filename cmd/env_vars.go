package cmd

import "github.com/caarlos0/env/v6"

type envVars struct {
	Token string `env:"POEDITOR_TOKEN"`
}

func newEnvVars() (*envVars, error) {
	vars := &envVars{}
	if err := env.Parse(vars); err != nil {
		return nil, err
	}
	return vars, nil
}
