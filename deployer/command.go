package deployer

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"sort"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (s *server) runCommand(c *gin.Context, command string) error {
	cmd := exec.Command("sh", "-c", command)
	data, err := cmd.CombinedOutput()
	if err != nil {
		c.Writer.Write([]byte(fmt.Sprintf("run command fail,err:%v, cmd:%v", err, command)))
		logrus.WithError(err).Error("run command error")
		return err
	}
	_, err = c.Writer.Write(data)
	if err != nil {
		logrus.WithError(err).Error("response command result error")
	}
	return nil
}

func (s *server) serviceRun(c *gin.Context, service *ConfigService, script string, files []*zip.File) error {
	succ := false
	defer func() {
		if !succ {
			c.Status(http.StatusForbidden)
		}
	}()
	if len(files) > 0 {
		if fileinfo, err := os.Stat(service.Directory); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				if err = os.MkdirAll(service.Directory, os.ModePerm); err != nil {
					c.Writer.Write([]byte(fmt.Sprintf("can not create directory:%v", service.Directory)))
					logrus.WithFields(logrus.Fields{"error": err, "service": service.Name, "directory": service.Directory}).
						Error("can not create directory")
					return err
				}
			} else {
				c.Writer.Write([]byte(fmt.Sprintf("can not stat directory:%v", service.Directory)))
				logrus.WithFields(logrus.Fields{"error": err, "service": service.Name, "directory": service.Directory}).
					Error("can not stat directory")
				return err
			}
		} else if !fileinfo.IsDir() {
			c.Writer.Write([]byte(fmt.Sprintf("service %v directory %v invalid", service.Name, service.Directory)))
			logrus.WithFields(logrus.Fields{"error": err, "service": service.Name, "directory": service.Directory}).
				Error("invalid service directory")
			return err
		}
	}
	for _, command := range service.BeforeScript {
		if err := s.runCommand(c, command); err != nil {
			return err
		}
	}
	sort.Slice(files, func(i, j int) bool {
		dirI := files[i].FileInfo().IsDir()
		dirJ := files[j].FileInfo().IsDir()
		if dirI == dirJ {
			if dirI {
				return len(files[i].Name) < len(files[j].Name)
			}
			return false
		} else if dirI {
			return true
		} else {
			return false
		}
	})
	for _, fileItem := range files {
		filename := path.Join(service.Directory, fileItem.Name)
		if fileItem.FileInfo().IsDir() {
			if err := os.MkdirAll(filename, os.ModePerm); err != nil && !errors.Is(err, os.ErrExist) {
				logrus.WithFields(logrus.Fields{"error": err, "filename": fileItem.Name}).Error("create directory fail")
				c.Writer.Write([]byte(fmt.Sprintf("create directory:%v fail", fileItem.Name)))
				return err
			}
			continue
		}
		reader, err := fileItem.Open()
		if err != nil {
			logrus.WithFields(logrus.Fields{"error": err, "filename": fileItem.Name}).Error("read zip file fail")
			c.Writer.Write([]byte(fmt.Sprintf("read zip file:%v fail", fileItem.Name)))
			return err
		}
		fileWriter, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, os.ModePerm)
		if err != nil {
			logrus.WithFields(logrus.Fields{"error": err, "filename": filename}).Error("open file fail")
			c.Writer.Write([]byte(fmt.Sprintf("can not open file:%v, err:%v", filename, err)))
			return err
		}
		if _, err = io.Copy(fileWriter, reader); err != nil {
			logrus.WithFields(logrus.Fields{"error": err, "filename": filename}).Error("write file fail")
			c.Writer.Write([]byte(fmt.Sprintf("can not write file:%v, err:%v", filename, err)))
			return err
		}
		reader.Close()
		fileWriter.Close()
		logrus.WithFields(logrus.Fields{"file": filename}).Info("copy file success")
	}
	logrus.WithFields(logrus.Fields{"command": script}).Debug("get script")
	if script != "" {
		if err := s.runCommand(c, script); err != nil {
			return err
		}
	}
	for _, command := range service.AfterScript {
		if err := s.runCommand(c, command); err != nil {
			return err
		}
	}
	succ = true
	c.Status(http.StatusOK)
	return nil
}
