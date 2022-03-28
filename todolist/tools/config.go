package tools

import "github.com/kelseyhightower/envconfig"

/*
  GetConfig sets the configuration depending on the environment or
  in the future we can load from the vault or injections with another strategy in the same func
*/

func GetConfig(prefix string, cfg interface{}) error {
	return envconfig.Process(prefix, cfg)
}
