package session

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"sync"

	restyv2 "github.com/go-resty/resty/v2"
)

const (
	EmsSvc = "ems-dashboard-api.ems.svc.cluster.local"
	EmsAPI = "/ems_dashboard_api/api/user_session/"
)

var (
	sm *sessionManager
)

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

type sessionManager struct {
	sync.Map
	restCli *restyv2.Client
}

func init() {
	sm = newSessionManager()
}

func newSessionManager() *sessionManager {
	return &sessionManager{
		restCli: restyv2.New().SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true}),
	}
}

func (sm *sessionManager) validateSession(r *http.Request) (*EmsSession, error) {
	session := &EmsSession{}
	resp, err := sm.restCli.
		R().
		SetResult(session).
		SetCookies(r.Cookies()).
		SetHeader("Accept", "application/json").
		SetHeader("X-Requested-With", "XMLHttpRequest").
		Get(fmt.Sprintf("%s://%s%s", "https", EmsSvc, EmsAPI))

	if err != nil {
		return nil, fmt.Errorf("authentication failed: %v", err)
	}
	if resp.RawResponse.StatusCode != 200 {
		return nil, fmt.Errorf("authentication failed, code: %d", resp.RawResponse.StatusCode)
	}
	return session, nil
}

func EmsSessionAuth(r *http.Request) (*EmsSession, error) {
	session, err := sm.validateSession(r)
	return session, err
}
