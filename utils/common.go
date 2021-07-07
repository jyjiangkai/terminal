package utils

import (
	"terminal/models"
	"terminal/pkg/eks"
	"terminal/pkg/eks/token"
	"encoding/base64"
	"fmt"
	"github.com/go-openapi/runtime"
	"net/http"
	"time"
)

const (
	OK                    = 200
	CREATED               = 201
	BAD_REQUEST           = 400
	UNAUTHORIZED          = 401
	FORBIDDEN             = 403
	NOT_FOUND             = 404
	INTERNAL_SERVER_ERROR = 500

	EcnsProjectID    = "project_id"
	EcnsRoles        = "roles"
	EcnsIsAdmin      = "is_admin"
	EcnsIsCloudAdmin = "is_cloud_admin"
)

func pBool(b bool) *bool {
	return &b
}

func pString(s string) *string {
	return &s
}

func pInt64(n int) *int64 {
	n64 := int64(n)
	return &n64
}

func pInt32(n int) *int32 {
	n32 := int32(n)
	return &n32
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
func time2string(t time.Time) *string {
	y, m, d := t.Date()
	s := fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02dZ", y, m, d, t.Hour(), t.Minute(), t.Second())
	return &s
}

func newAPIResponse(code int, message string, data interface{}) *APIResponse {
	return &APIResponse{
		Code:    int32(code),
		Message: message,
		Data:    data,
	}
}

// APIResponse is copied from appstore/models/api_response.go to define our own ResponseWriter
// for simplicity
type APIResponse struct {

	// code
	Code int32 `json:"code,omitempty"`

	// data
	Data interface{} `json:"data,omitempty"`

	// message
	Message string `json:"message,omitempty"`
}

func (r *APIResponse) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(int(r.Code))
	if err := producer.Produce(rw, r); err != nil {
		panic(err) // let the recovery middleware deal with this
	}
}

type myCluster models.Cluster

func (c myCluster) validateToken(token string) error {
	cli, err := eks.NewEKSClient(*c.APIServerAddress, token)
	if err != nil {
		return err
	}
	_, err = cli.ServerVersion()
	return err
}

func GetToken(r *http.Request, c *models.Cluster, pid string) (string, error) {
	//从缓存中获取token
	t, err := token.GetToken(r.Cookies(), pid)
	if err != nil {
		return "", err
	}
	// 如果token有效，就返回
	err = myCluster(*c).validateToken(t)
	if err == nil {
		return t, nil
	}

	//否则重新从openstack查询一次token
	t, err = token.Reissue(r.Cookies(), pid)
	if err != nil {
		return "", err
	}
	// 再验证一次有效性
	err = myCluster(*c).validateToken(t)
	if err != nil {
		return "", err
	}
	return t, err
}
