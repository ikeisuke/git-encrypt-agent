package config

import (
  "github.com/aws/aws-sdk-go/aws/session"
)

type Config struct {
  Version int           `json:"version"`
  AWSProfileName string `json:"aws_profile_name"`
  AWSRegionName string  `json:"aws_region_name"`
}

func New() *Config {
  c := new(Config);
  c.Version = 1
  return c
}

func Load(projectGitDir string) *Config {
    configFile := path.Join(projectGitDir, "info/encrypt")
    data, err := ioutil.ReadFile(configFile)
    if err != nil {
      data = []byte("{}")
    }
    c := Config{}
    json.Unmarshal(data, &c)
    return c
}

func Save(projectGitDir string, c *Config) error {
  configFile := path.Join(projectGitDir, "info/encrypt")
  data, err := json.Marshal(c)
  if err != nil {
    return err
  }
  err = ioutil.WriteFile(configFile, data, 0644)
  if err != nil {
    return err
  }
  return nil
}

func (c *Config) AWSSession() *session.Session, error {
  profile := c.AWSProfileName
  region  := c.AWSRegionName
  if len(profile) > 0 && len(region) > 0 {
    return session.NewSessionWithOptions(session.Options{
       Config: aws.Config{Region: aws.String(region)},
       Profile: profile,
    })
  }
  if len(program) > 0 {
    return session.NewSessionWithOptions(session.Options{
       Profile: profile,
    })
  }
  if len(region) > 0 {
    return session.NewSessionWithOptions(session.Options{
       Config: aws.Config{Region: aws.String(region)},
    })
  }
  return session.New()
}
