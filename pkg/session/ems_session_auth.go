package session

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"sync"

	restyv2 "github.com/go-resty/resty/v2"
)

const (
	sessionid = "sessionid"
)

var (
	sm *sessionManager
)

func init() {
	sm = newSessionManager()
}

//func sessionID(c *gin.Context) string {
//	for _, cookie := range c.Request.Cookies() {
//		if cookie.Name == sessionid {
//			return cookie.Value
//		}
//	}
//	return ""
//}

type sessionManager struct {
	sync.Map
	restCli *restyv2.Client
}

func newSessionManager() *sessionManager {
	return &sessionManager{
		restCli: restyv2.New().SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true}),
	}
}

func (sm *sessionManager) validateSession(r *http.Request) (*EmsSession, error) {
	// https://easystack.atlassian.net/browse/EAS-72968
	// 由于用户可能属于多个项目，且会随时切换。所以不能将session缓存起来，必须每次
	// API 调用时去查询当前的项目ID
	//sId := sessionID(cntx)
	//if s, ok := sm.Load(sId); ok {
	//	return s.(*EmsSession), nil
	//}

	session := &EmsSession{}
	resp, err := sm.restCli.
		R().
		SetResult(session).
		SetCookies(r.Cookies()).
		SetHeader("Accept", "application/json").
		SetHeader("X-Requested-With", "XMLHttpRequest").
		Get(fmt.Sprintf("%s://%s%s", "https", "ems-dashboard-api.ems.svc.cluster.local", "/ems_dashboard_api/api/user_session/"))

	if err != nil {
		return nil, fmt.Errorf("authentication failed: %v", err)
	}
	if resp.RawResponse.StatusCode != 200 {
		return nil, fmt.Errorf("authentication failed, code: %d", resp.RawResponse.StatusCode)
	}

	// 理由见上
	//go func() {
	//	sm.Store(sId, session)
	//	<-time.After(120 * time.Second)
	//	sm.Delete(sId)
	//}()

	return session, nil
}

func EmsSessionAuth(r *http.Request) (*EmsSession, error) {
	session, err := sm.validateSession(r)
	return session, err
}

type EmsSession struct {
	IdpAuthURL       interface{} `json:"idp_auth_url"`
	Name             string      `json:"name"`
	Roles            []string    `json:"roles"`
	SupportLdap      bool        `json:"support_ldap"`
	DomainID         string      `json:"domain_id"`
	IdpUnscopedToken interface{} `json:"idp_unscoped_token"`
	ProjectID        string      `json:"project_id"`
	ID               string      `json:"id"`
	Email            string      `json:"email"`
}
