package config

import (
    "gopkg.in/yaml.v2"
    "io/ioutil"
)

type Config struct {
    SQLDSN            string `yaml:"sql_dsn"`
    MySQLDSN            string `yaml:"mysql_dsn"`
    MongoURI          string `yaml:"mongo_uri"`
    RedisURI          string `yaml:"redis_uri"`
    ElasticsearchURI  string `yaml:"es_uri"`
}

func LoadConfigFromFile(filePath string) (*Config, error) {
    data, err := ioutil.ReadFile(filePath)
    if err != nil {
        return nil, err
    }
    var cfg Config
    err = yaml.Unmarshal(data, &cfg)
    return &cfg, err
}
