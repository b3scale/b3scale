package config

import "log"


type FrontendsConfig struct {
    File string
}

func NewFrontendsConfig(file string) *FrontendsConfig {
    c := &FrontendsConfig{
        File: file,
    }

    return c
}


