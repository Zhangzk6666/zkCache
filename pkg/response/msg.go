package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type responseMsg struct {
}

var ResponseMsg responseMsg
var msg = map[int]string{
	SUCCESS:         "success",
	ERROR:           "服务器异常",
	PARAMETER_ERROR: "参数不全或有误",
}

func getMsg(code int) interface{} {
	if data, ok := msg[code]; ok {
		return data
	}
	return nil
}

func (r *responseMsg) SuccessResponse(c *gin.Context, data interface{}) {
	c.JSON(200, gin.H{
		"code": SUCCESS,
		"msg":  getMsg(SUCCESS),
		"data": data,
	})
}

func (r *responseMsg) FailResponse(c *gin.Context, err error, data interface{}) {
	switch res := err.(type) {
	case *resErr:
		c.JSON(http.StatusBadRequest, gin.H{
			"code": res.Code,
			"msg":  res.Msg,
			"data": data,
		})
	default:
		r.ErrorResponse(c, nil)
	}
}

func (r *responseMsg) OtherStatusResponse(c *gin.Context, httpStatus int, code int, data interface{}) {
	c.JSON(httpStatus, gin.H{
		"code": code,
		"msg":  getMsg(code),
		"data": data,
	})
}

func (r *responseMsg) ErrorResponse(c *gin.Context, data interface{}) {
	c.JSON(500, gin.H{
		"code": ERROR,
		"msg":  getMsg(ERROR),
		"data": data,
	})
}

type PanicRes struct {
	Code int
	Data interface{}
}

func (this PanicRes) PanicResponse(c *gin.Context, panicRes PanicRes) {
	c.JSON(200, gin.H{
		"code": panicRes.Code,
		"msg":  getMsg(panicRes.Code),
		"data": panicRes.Data,
	})
}
