package base

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/core/validation"
	"github.com/beego/beego/v2/server/web"

	"mall/controllers/base/session"
)

type BaseController struct {
	web.Controller

	IsAuth   bool
	IsCheck  bool
	SourceIp string
}

type GeneralResponse struct {
	Code    int         `json:"code,omitempty"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func (c *BaseController) Prepare() {
	c.SourceIp = c.Ctx.Input.IP()
}

func (c *BaseController) ServeResponse(errMsg *GeneralResponse, args ...interface{}) {
	if nil == errMsg {
		return
	}

	data := errMsg.Data
	if len(args) >= 1 {
		if nil != args[0] {
			data = args[0]
		}
	}

	result := &GeneralResponse{
		Code:    errMsg.Code,
		Message: errMsg.Message,
		Data:    data,
	}

	c.Data["json"] = result

	if err := c.ServeJSON(); err != nil {
		logs.Warn("base controller sends a json response failed, err: %s", err.Error())
	}
}

func (c *BaseController) MyGetSession(name string) (interface{}, bool) {
	sessItf := c.GetSession(name)
	sess, ok := sessItf.(string)
	if !ok {
		return nil, false
	}

	switch name {
	case session.KSessKeyUser:
		result := session.User{}
		err := json.Unmarshal([]byte(sess), &result)
		if err != nil {
			logs.Debug("json unmarshal session [%s] err: %s", name, err)
			return nil, false
		}
		return result, true
	}
	return nil, false
}

func (c *BaseController) InputCheck(obj interface{}) *GeneralResponse {
	if obj == nil {
		return ErrSystem
	}

	c.IsCheck = true

	//parse request parameters
	if err := c.ParseIParams(obj); err != nil {
		logs.Warn("parse err: %s", err.Error())
		return ErrInputData
	}

	if err := c.VerifyParams(obj); err != nil {
		logs.Warn("verify err: %s", err.Error())
		return ErrInputData
	}

	return nil
}

func (c *BaseController) ParseIParams(obj interface{}) error {
	if c.Ctx.Request.Method == http.MethodGet {
		if err := c.ParseForm(obj); err != nil {
			return err
		}
		return nil
	}

	if len(c.Ctx.Input.RequestBody) == 0 {
		return nil
	}

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, obj); err != nil {
		logs.Warn("unmarshal request body failed, err: %s", err.Error())
		return err
	}

	return nil
}

func (c *BaseController) VerifyParams(obj interface{}) error {
	//verify
	valid := validation.Validation{}
	ok, err := valid.RecursiveValid(obj)
	if err != nil {
		return err
	}

	if !ok {
		str := ""
		for _, err := range valid.Errors {
			str += err.Key + ":" + err.Message + ";"
		}
		return fmt.Errorf(str)
	}
	return nil
}
