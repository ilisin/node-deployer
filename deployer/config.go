package deployer

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type provider string

const (
	providerFeiShu provider = "feishu"
)

type NotifyTPL struct {
	Name     string   `yaml:"name"`
	Provider provider `yaml:"provider"`
	URL      string   `yaml:"url"` // provider url
	MsgTPL   string   `yaml:"msg_tpl"`
}

type ConfigService struct {
	Name         string       `yaml:"name"`
	Token        string       `yaml:"token"`
	BeforeScript []string     `yaml:"before_script"`
	AfterScript  []string     `yaml:"after_script"`
	Directory    string       `yaml:"directory"`
	NotifyTPLs   []*NotifyTPL `yaml:"notify_tpls"`
}

type Config struct {
	Port     int `yaml:"port"`
	Services []*ConfigService
}

func (c *Config) update(to *Config) {
	c.Services = to.Services
}

func loadConfig(fromFile string) (*Config, error) {
	data, err := os.ReadFile(fromFile)
	if err != nil {
		return nil, fmt.Errorf("load file config fail,file:%v,err:%v", fromFile, err)
	}
	cfg := &Config{}
	if err = yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config fail,file:%v,err:%v", fromFile, err)
	}
	return cfg, nil
}

func watcherConfig(fromFile string, cfg *Config, stop chan struct{}) {
	dir := filepath.Dir(fromFile)
	filename := filepath.Base(fromFile)
	// Create new watcher.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logrus.WithError(err).Error("config file watch fail")
		return
	}
	// Add a path.
	err = watcher.Add(dir)
	if err != nil {
		logrus.WithError(err).Error("add directory watcher fail")
	}

	// Start listening for events.
LOOP:
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok || event.Name != filename {
				continue
			}
			if newCfg, err := loadConfig(fromFile); err != nil {
				logrus.WithError(err).Error("file change ,reload config fail")
			} else {
				cfg.update(newCfg)
				logrus.Infof("update config success")
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			logrus.WithError(err).Error("file watcher fail")
		case <-stop:
			break LOOP
		}
	}

	watcher.Close()

}
