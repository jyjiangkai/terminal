package token

import (
	"net/http"
	"testing"
)

const (
	session  = "ma2z5cdlcw9v792q1nzofdp8414f02cw"
	ecsip    = "172.33.0.2"
	tokenAPI = "/api/ecns/eks/token/"
)

func Test_getToken(t *testing.T) {
	mgr := newEksTokenManager()
	token, err := mgr.issue([]*http.Cookie{{Name: sessionid, Value: session}}, "https", ecsip, tokenAPI)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(token)

}
