package deployer

import (
	"archive/zip"
	"bytes"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	keyContextService = "__service"
)

func (s *server) registerAPIs() {
	group := s.gin.Group("/deploy", s.serviceCheck)
	{
		group.POST("/upload", s.uploadFile)
		group.POST("/notify", s.deployNotify)
	}
}

func (s *server) serviceCheck(c *gin.Context) {
	serviceName := c.Query("service")
	token := c.Query("token")
	if serviceName == "" {
		c.String(http.StatusNotFound, "invalid service")
		c.Abort()
		return
	}
	var service *ConfigService
	for _, svc := range s.cfg.Services {
		if svc.Name == serviceName {
			service = svc
			break
		}
	}
	if service == nil {
		c.String(http.StatusNotFound, "invalid service")
		c.Abort()
		return
	}
	if token != service.Token {
		c.String(http.StatusNotFound, "invalid service token")
		c.Abort()
		return
	}
	c.Set(keyContextService, service)
	c.Next()
}

func mustGetContextService(c *gin.Context) *ConfigService {
	serviceVal, ok := c.Get(keyContextService)
	if !ok {
		panic("get service context fail")
	}
	return serviceVal.(*ConfigService)
}

func (s *server) uploadFile(c *gin.Context) {
	service := mustGetContextService(c)
	file, err := c.FormFile("file")
	if err != nil {
		c.String(http.StatusNotFound, "invalid file")
		return
	}
	f, err := file.Open()
	if err != nil {
		c.String(http.StatusNotFound, "invalid file")
		return
	}
	data, err := io.ReadAll(f)
	if err != nil {
		c.String(http.StatusNotFound, "read zip fail fail")
		return
	}
	zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		c.String(http.StatusNotFound, "not a zip file")
		return
	}
	_ = s.serviceRun(c, service, c.Query("command"), zipReader.File)
	return
}

func getAllQueryParam(c *gin.Context) map[string]string {
	result := make(map[string]string, 0)
	for key := range c.Request.URL.Query() {
		result[key] = c.Query(key)
	}
	return result
}

func (s *server) deployNotify(c *gin.Context) {
	service := mustGetContextService(c)
	params := getAllQueryParam(c)
	var rawMsg []byte
	if msg, ok := params["message"]; ok {
		rawMsg = []byte(msg)
	} else {
		file, err := c.FormFile("file")
		if err != nil {
			c.String(http.StatusNotFound, "invalid file")
			return
		} else if openFile, err := file.Open(); err != nil {
			c.String(http.StatusNotFound, "invalid file")
			return
		} else if rawMsg, err = io.ReadAll(openFile); err != nil {
			c.String(http.StatusNotFound, "invalid file")
			return
		} else {
			_ = openFile.Close()
		}
	}
	var tpl *NotifyTPL
	tplName, ok := params["tpl"]
	if !ok && len(service.NotifyTPLs) != 1 {
		c.String(http.StatusNotFound, "invalid notify tpl")
		return
	} else if !ok {
		tpl = service.NotifyTPLs[0]
	} else {
		for _, t := range service.NotifyTPLs {
			if t.Name == tplName {
				tpl = t
				break
			}
		}
	}
	if tpl == nil {
		c.String(http.StatusNotFound, "invalid notify tpl")
		return
	}
	s.notify(c, tpl, string(rawMsg), params)
}
