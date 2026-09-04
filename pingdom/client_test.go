package pingdom

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// newTestClient returns a client wired to a httptest server. Use the returned
// teardown to release server resources.
func newTestClient(t *testing.T, handler http.HandlerFunc) (*Client, func()) {
	t.Helper()
	srv := httptest.NewServer(handler)
	client, err := NewClientWithConfig(ClientConfig{
		APIToken: "test-token",
		BaseURL:  srv.URL,
	})
	if err != nil {
		srv.Close()
		t.Fatalf("NewClientWithConfig: %v", err)
	}
	return client, srv.Close
}

func TestClient_ReadCheck_Success(t *testing.T) {
	var capturedAuth, capturedQuery string
	client, teardown := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		capturedAuth = r.Header.Get("Authorization")
		capturedQuery = r.URL.RawQuery
		if r.URL.Path != "/checks/42" {
			t.Errorf("URL.Path = %q, want /checks/42", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{"check":{"id":42,"name":"api-prod","hostname":"api.example.com","status":"up","type":"ping","teams":[{"id":7,"name":"oncall"}]}}`)
	})
	defer teardown()

	ck, err := client.Checks.Read(42)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if ck.ID != 42 || ck.Name != "api-prod" {
		t.Errorf("got %+v, want id=42 name=api-prod", ck)
	}
	if ck.Type.Name != "ping" {
		t.Errorf("Type.Name = %q, want ping", ck.Type.Name)
	}
	if len(ck.TeamIds) != 1 || ck.TeamIds[0] != 7 {
		t.Errorf("TeamIds backfill = %v, want [7]", ck.TeamIds)
	}
	if capturedAuth != "Bearer test-token" {
		t.Errorf("Authorization = %q, want Bearer test-token", capturedAuth)
	}
	if !strings.Contains(capturedQuery, "include_teams=true") {
		t.Errorf("query = %q, want include_teams=true", capturedQuery)
	}
}

func TestClient_ReadCheck_404(t *testing.T) {
	client, teardown := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, `{"error":{"statuscode":404,"statusdesc":"Not Found","errormessage":"Check not found"}}`)
	})
	defer teardown()

	_, err := client.Checks.Read(99)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	perr, ok := err.(*PingdomError)
	if !ok {
		t.Fatalf("err type = %T, want *PingdomError", err)
	}
	if perr.StatusCode != 404 {
		t.Errorf("StatusCode = %d, want 404", perr.StatusCode)
	}
}

func TestClient_CheckResponseType_HTTPDecode(t *testing.T) {
	client, teardown := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{"check":{"id":1,"name":"web","hostname":"x.com","type":{"http":{"url":"/health","encryption":true,"port":443,"verify_certificate":true}}}}`)
	})
	defer teardown()

	ck, err := client.Checks.Read(1)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if ck.Type.Name != "http" {
		t.Errorf("Type.Name = %q, want http", ck.Type.Name)
	}
	if ck.Type.HTTP == nil {
		t.Fatal("Type.HTTP is nil")
	}
	if ck.Type.HTTP.Url != "/health" {
		t.Errorf("Type.HTTP.Url = %q, want /health", ck.Type.HTTP.Url)
	}
	if ck.Type.HTTP.Port != 443 {
		t.Errorf("Type.HTTP.Port = %d, want 443", ck.Type.HTTP.Port)
	}
}

func TestClient_CreateCheck_HTTP_QueryParams(t *testing.T) {
	var captured map[string]string
	client, teardown := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Method = %q, want POST", r.Method)
		}
		if r.URL.Path != "/checks" {
			t.Errorf("URL.Path = %q, want /checks", r.URL.Path)
		}
		captured = map[string]string{}
		for k, v := range r.URL.Query() {
			captured[k] = v[0]
		}
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{"check":{"id":99}}`)
	})
	defer teardown()

	verify := true
	ssl := 14
	ck := &HttpCheck{
		Name:              "shop",
		Hostname:          "shop.example.com",
		Resolution:        5,
		Url:               "/health",
		Encryption:        true,
		Port:              443,
		ShouldContain:     "OK",
		Tags:              "prod,web",
		VerifyCertificate: &verify,
		SSLDownDaysBefore: &ssl,
		IntegrationIds:    []int{1, 2, 3},
	}
	if _, err := client.Checks.Create(ck); err != nil {
		t.Fatalf("Create: %v", err)
	}

	expect := map[string]string{
		"type":                 "http",
		"name":                 "shop",
		"host":                 "shop.example.com",
		"resolution":           "5",
		"url":                  "/health",
		"encryption":           "true",
		"port":                 "443",
		"shouldcontain":        "OK",
		"tags":                 "prod,web",
		"verify_certificate":   "true",
		"ssl_down_days_before": "14",
		"integrationids":       "1,2,3",
	}
	for k, want := range expect {
		if got := captured[k]; got != want {
			t.Errorf("param %q = %q, want %q", k, got, want)
		}
	}
}

func TestClient_CreateCheck_OmitsEmptyOnPost(t *testing.T) {
	var captured map[string]string
	client, teardown := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		captured = map[string]string{}
		for k, v := range r.URL.Query() {
			captured[k] = v[0]
		}
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{"check":{"id":1}}`)
	})
	defer teardown()

	ck := &PingCheck{Name: "x", Hostname: "y", Resolution: 1}
	if _, err := client.Checks.Create(ck); err != nil {
		t.Fatalf("Create: %v", err)
	}
	// postParams() should strip empty-string keys to satisfy Pingdom Create.
	for _, banned := range []string{"integrationids", "userids", "teamids", "probe_filters"} {
		if _, present := captured[banned]; present {
			t.Errorf("Create should omit empty %q, got %q", banned, captured[banned])
		}
	}
}

func TestClient_CreateContact_JSONBody(t *testing.T) {
	var capturedCT string
	var capturedBody map[string]interface{}
	client, teardown := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/alerting/contacts" {
			t.Errorf("got %s %s, want POST /alerting/contacts", r.Method, r.URL.Path)
		}
		capturedCT = r.Header.Get("Content-Type")
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &capturedBody)
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{"contact":{"id":555,"name":"oncall","paused":false}}`)
	})
	defer teardown()

	c := &Contact{
		Name:   "oncall",
		Paused: false,
		NotificationTargets: NotificationTargets{
			Email: []EmailNotification{{Address: "alerts@example.com", Severity: "HIGH"}},
		},
	}
	got, err := client.Contacts.Create(c)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if got.ID != 555 {
		t.Errorf("ID = %d, want 555", got.ID)
	}
	if capturedCT != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", capturedCT)
	}
	if capturedBody["name"] != "oncall" {
		t.Errorf("body.name = %v, want oncall", capturedBody["name"])
	}
}

func TestClient_TeamLifecycle(t *testing.T) {
	calls := []string{}
	client, teardown := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.Method+" "+r.URL.Path)
		w.WriteHeader(http.StatusOK)
		switch {
		case r.Method == "POST" && r.URL.Path == "/alerting/teams":
			_, _ = io.WriteString(w, `{"team":{"id":11,"name":"team-a"}}`)
		case r.Method == "GET" && r.URL.Path == "/alerting/teams/11":
			_, _ = io.WriteString(w, `{"team":{"id":11,"name":"team-a","members":[{"id":1,"name":"alice"}]}}`)
		case r.Method == "PUT" && r.URL.Path == "/alerting/teams/11":
			_, _ = io.WriteString(w, `{"team":{"id":11,"name":"team-b","members":[]}}`)
		case r.Method == "DELETE" && r.URL.Path == "/alerting/teams/11":
			_, _ = io.WriteString(w, `{"message":"deleted"}`)
		}
	})
	defer teardown()

	created, err := client.Teams.Create(&Team{Name: "team-a", MemberIDs: []int{1}})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.ID != 11 {
		t.Errorf("Create ID = %d, want 11", created.ID)
	}

	read, err := client.Teams.Read(11)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(read.Members) != 1 || read.Members[0].Name != "alice" {
		t.Errorf("Members = %v, want one (alice)", read.Members)
	}

	updated, err := client.Teams.Update(11, &Team{Name: "team-b", MemberIDs: nil})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Name != "team-b" {
		t.Errorf("Update Name = %q, want team-b", updated.Name)
	}

	del, err := client.Teams.Delete(11)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if del.Message != "deleted" {
		t.Errorf("Delete message = %q, want deleted", del.Message)
	}

	want := []string{
		"POST /alerting/teams",
		"GET /alerting/teams/11",
		"PUT /alerting/teams/11",
		"DELETE /alerting/teams/11",
	}
	if len(calls) != len(want) {
		t.Fatalf("calls = %v, want %v", calls, want)
	}
	for i := range want {
		if calls[i] != want[i] {
			t.Errorf("call[%d] = %q, want %q", i, calls[i], want[i])
		}
	}
}

func TestClient_HTTPCheck_PostParamsContainsHeaders(t *testing.T) {
	verify := true
	ck := &HttpCheck{
		Name:       "h",
		Hostname:   "h.example.com",
		Resolution: 1,
		Url:        "/",
		RequestHeaders: map[string]string{
			"X-Beta":  "y",
			"X-Alpha": "x",
		},
		VerifyCertificate: &verify,
	}
	p := ck.postParams()
	// Headers must be emitted sorted by key (alpha < beta) into requestheaderN.
	if p["requestheader0"] != "X-Alpha:x" {
		t.Errorf("requestheader0 = %q, want X-Alpha:x", p["requestheader0"])
	}
	if p["requestheader1"] != "X-Beta:y" {
		t.Errorf("requestheader1 = %q, want X-Beta:y", p["requestheader1"])
	}
}

func TestClient_HTTPCheck_AuthInline(t *testing.T) {
	verify := true
	ck := &HttpCheck{
		Name:              "h",
		Hostname:          "h.example.com",
		Resolution:        1,
		Url:               "/",
		Username:          "alice",
		Password:          "s3cret",
		VerifyCertificate: &verify,
	}
	p := ck.postParams()
	if p["auth"] != "alice:s3cret" {
		t.Errorf("auth = %q, want alice:s3cret", p["auth"])
	}
}

func TestClient_HTTPCheck_AuthOmittedWhenUnset(t *testing.T) {
	verify := true
	ck := &HttpCheck{
		Name:              "h",
		Hostname:          "h.example.com",
		Resolution:        1,
		Url:               "/",
		VerifyCertificate: &verify,
	}
	if _, present := ck.putParams()["auth"]; present {
		t.Error("auth must be absent from an update that is not changing credentials")
	}
}

func TestClient_HTTPCheck_ClearAuthSendsEmptyAuth(t *testing.T) {
	verify := true
	ck := &HttpCheck{
		Name:              "h",
		Hostname:          "h.example.com",
		Resolution:        1,
		Url:               "/",
		ClearAuth:         true,
		VerifyCertificate: &verify,
	}
	p := ck.putParams()
	v, present := p["auth"]
	if !present {
		t.Fatal("auth must be present so the server clears the stored credentials")
	}
	if v != "" {
		t.Errorf("auth = %q, want empty", v)
	}
}

func TestClient_HTTPCheck_ClearAuthIgnoredWhenCredentialsSet(t *testing.T) {
	verify := true
	ck := &HttpCheck{
		Name:              "h",
		Hostname:          "h.example.com",
		Resolution:        1,
		Url:               "/",
		Username:          "alice",
		Password:          "s3cret",
		ClearAuth:         true,
		VerifyCertificate: &verify,
	}
	if got := ck.putParams()["auth"]; got != "alice:s3cret" {
		t.Errorf("auth = %q, want alice:s3cret", got)
	}
}

func TestClient_HTTPCheck_ClearAuthStrippedOnCreate(t *testing.T) {
	verify := true
	ck := &HttpCheck{
		Name:              "h",
		Hostname:          "h.example.com",
		Resolution:        1,
		Url:               "/",
		ClearAuth:         true,
		VerifyCertificate: &verify,
	}
	if _, present := ck.postParams()["auth"]; present {
		t.Error("create must not carry an empty auth - Pingdom rejects empty values at create time")
	}
}

func TestClient_TCPCheck_PortRequired(t *testing.T) {
	ck := &TCPCheck{Name: "x", Hostname: "y", Resolution: 5, Port: 0}
	if err := ck.valid(); err == nil {
		t.Fatal("expected validation error for Port=0, got nil")
	}
}

func TestClient_HTTPCheck_ShouldContainAndShouldNotContainMutuallyExclusive(t *testing.T) {
	ck := &HttpCheck{
		Name:             "x",
		Hostname:         "y",
		Resolution:       5,
		ShouldContain:    "a",
		ShouldNotContain: "b",
	}
	if err := ck.valid(); err == nil {
		t.Fatal("expected validation error for both ShouldContain and ShouldNotContain set, got nil")
	}
}

func TestClient_ErrorEnvelope_FallbackOnMalformed(t *testing.T) {
	client, teardown := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, `<html>nginx 500</html>`) // not JSON
	})
	defer teardown()

	_, err := client.Checks.Read(1)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	perr, ok := err.(*PingdomError)
	if !ok {
		t.Fatalf("err type = %T, want *PingdomError fallback", err)
	}
	if perr.StatusCode != 500 {
		t.Errorf("StatusCode = %d, want 500", perr.StatusCode)
	}
}
