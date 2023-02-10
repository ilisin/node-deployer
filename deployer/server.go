package deployer

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/ilisin/utils.go/httpclient"
	"github.com/sirupsen/logrus"
)

type Server interface {
	Run()
}

type server struct {
	cfg *Config
	gin *gin.Engine

	httpClientFeishu *httpclient.HttpClient
}

func newServer(config *Config) *server {
	return &server{
		cfg: config,
		gin: gin.Default(),
	}
}

func (s *server) Run() {
	s.registerAPIs()
	if err := s.gin.Run(fmt.Sprintf(":%v", s.cfg.Port)); err != nil {
		logrus.Infof("api server shutdown with %v", err)
	}
}

func NewServer(cfgFile string) (Server, error) {
	cfg, err := loadConfig(cfgFile)
	if err != nil {
		return nil, err
	} else {
		go watcherConfig(cfgFile, cfg, make(chan struct{}))
	}
	return newServer(cfg), nil
}

func init() {
	logrus.SetLevel(logrus.DebugLevel)
}
