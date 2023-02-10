package deployer

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ilisin/utils.go/httpclient"
	"github.com/sirupsen/logrus"
)

const (
	variableCommitMsg = "${COMMIT_MSG}"
)

type FeiShuBotNotifyReq struct {
	MsgType string `json:"msg_type"`
	Content struct {
		Text string `json:"text"`
	} `json:"content"`
}

type FeiShuBotNotifyRsp struct {
	StatusCode    int
	StatusMessage string
}

func renderText(tpl string, variables map[string]string) string {
	for k, v := range variables {
		tpl = strings.Replace(tpl, k, v, -1)
	}
	return tpl
}

func (s *server) providerNotify(tpl *NotifyTPL, content string) error {
	switch tpl.Provider {
	case providerFeiShu:
		req := &FeiShuBotNotifyReq{
			MsgType: "text",
			Content: struct {
				Text string `json:"text"`
			}{Text: content},
		}
		rsp := &FeiShuBotNotifyRsp{}
		if s.httpClientFeishu == nil {
			s.httpClientFeishu = httpclient.New(httpclient.NewTimeoutOption(10 * time.Second))
		}
		if err := s.httpClientFeishu.POST(tpl.URL, req, rsp); err != nil {
			logrus.WithError(err).Error("feishu httpclient notify fail")
			return err
		} else if rsp.StatusCode != 0 {
			logrus.WithFields(logrus.Fields{"rsp": rsp}).Error("feishu httpclient notify fail")
			return fmt.Errorf("[%v] %v", rsp.StatusCode, rsp.StatusMessage)
		}
		return nil
	default:
		return errors.New("invalid provider")
	}
}

func (s *server) notify(c *gin.Context, tpl *NotifyTPL, fileMsg string, params map[string]string) {
	variables := map[string]string{
		variableCommitMsg: fileMsg,
	}
	for k, v := range params {
		variables["${"+k+"}"] = v
	}
	content := renderText(tpl.MsgTPL, variables)
	if err := s.providerNotify(tpl, content); err != nil {
		c.Writer.Write([]byte(fmt.Sprintf("notify fail,err:%v", err)))
		c.Status(http.StatusForbidden)
		return
	}
	c.String(http.StatusOK, "notify success!")
}
