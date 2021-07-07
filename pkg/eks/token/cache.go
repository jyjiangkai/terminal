package token

import (
	"crypto/tls"
	"fmt"
	restyv2 "github.com/go-resty/resty/v2"
	"net/http"
	"sync"
	"time"
)

const (
	sessionid = "sessionid"
)

var (
	tokenManager *eksTokenManager
)

func init() {
	tokenManager = newEksTokenManager()
}

func sessionID(cookies []*http.Cookie) string {
	for _, cookie := range cookies {
		if cookie.Name == sessionid {
			return cookie.Value
		}
	}
	return ""
}

// 获取当前session的 token，如果token不存在，就从openstack获取，并存到缓存中
func GetToken(cookies []*http.Cookie, projectID string) (string, error) {
	// 根据 https://easystack.atlassian.net/browse/EAS-73995
	// 当用户切换了项目时，需要重新issue当前项目的token，这样才对项目中的 eks 集群有
	// 足够权限。所以在缓存中，key要加上projectID作为相同session下切换了项目后的区分
	if token, ok := tokenManager.Load(sessionID(cookies) + projectID); ok {
		return token.(userToken).Token, nil
	}
	return Reissue(cookies, projectID)
}

// 从openstack 获取session的 token，并将它存入缓存，如果已经有了，就覆盖
func Reissue(cookies []*http.Cookie, projectID string) (string, error) {
	token, err := tokenManager.issue(cookies, "http", "eks-dashboard-api.eks.svc.cluster.local", "/api/ecns/token/")
	if err != nil {
		return "", err
	}
	tokenManager.Store(sessionID(cookies)+projectID, token)
	return token.Token, err
}

type userToken struct {
	Token   string `json:"token"`
	Expires string `json:"expires"`
}

func (s userToken) expired() bool {
	tt, err := time.Parse(time.RFC3339, s.Expires)
	if err != nil {
		return true
	}

	if tt.After(time.Now().UTC()) {
		return false
	} else {
		return true
	}
}

type eksTokenManager struct {
	sync.Map
	restCli *restyv2.Client
}

func newEksTokenManager() *eksTokenManager {
	m := &eksTokenManager{restCli: restyv2.New().SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})}
	go m.cleanExpired()
	return m
}

func (e *eksTokenManager) issue(cookies []*http.Cookie, scheme, eksSvc, apiPath string) (userToken, error) {
	var token userToken
	tokenResp, err := e.restCli.
		R().
		SetResult(&token).
		SetCookies(cookies).
		SetHeader("Accept", "application/json").
		SetHeader("X-Requested-With", "XMLHttpRequest").
		Get(fmt.Sprintf("%s://%s%s", scheme, eksSvc, apiPath))
	if err != nil {
		return userToken{}, err
	}
	defer tokenResp.RawResponse.Body.Close()

	if tokenResp.IsError() {
		return userToken{}, fmt.Errorf("getting token Error")
	}
	return token, nil
}

// 每十秒钟清理一次过期token，无限循环
func (e *eksTokenManager) cleanExpired() {
	for {
		<-time.After(10 * time.Second)
		var keysToRemove []interface{}
		e.Range(func(key, value interface{}) bool {
			if value.(userToken).expired() {
				keysToRemove = append(keysToRemove, key)
			}
			return true
		})

		for _, key := range keysToRemove {
			e.Delete(key)
		}
	}
}
