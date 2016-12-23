package cmd

type Config struct {
  Version int           `json:"version"`
  AWSProfileName string `json:"aws_profile_name"`
  AWSRegionName string  `json:"aws_region_name"`
}

func NewConfig() *Config {
  c := new(Config);
  c.Version = 1
  return c
}
