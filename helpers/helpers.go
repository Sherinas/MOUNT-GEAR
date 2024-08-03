package helpers

import (
	"github.com/gin-gonic/gin"
)

type ResponseData struct {
	Status     string      `json:"status"`
	StatusCode int         `json:"status_code"`
	Message    string      `json:"message"`
	Error      string      `json:"error,omitempty"`
	Data       interface{} `json:"data,omitempty"`
}

func SendResponse(c *gin.Context, statusCode int, message string, err error, data ...interface{}) {
	response := ResponseData{
		StatusCode: statusCode,
		Message:    message,
	}

	if statusCode >= 400 {
		response.Status = "error"
		if err != nil {
			response.Error = err.Error()
		}
	} else {
		response.Status = "success"
		if len(data) > 0 && data[0] != nil {
			response.Data = data[0]
		}
	}

	c.JSON(statusCode, response)
}
