package main

import (
	"errors"
	"github.com/go-viper/mapstructure/v2"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"os"
)

func loadConfig(configFile string, config any) error {
	k := koanf.New(".")
	if configFile != "" {
		err := k.Load(file.Provider(configFile), yaml.Parser())
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	err := k.UnmarshalWithConf("", config, koanf.UnmarshalConf{
		Tag: "koanf",
		DecoderConfig: &mapstructure.DecoderConfig{
			DecodeHook: mapstructure.OrComposeDecodeHookFunc(
				mapstructure.StringToTimeDurationHookFunc()),
			Result:      config,
			ErrorUnused: true,
			ErrorUnset:  true,
		},
	})
	if err != nil {
		return err
	}
	return nil
}
