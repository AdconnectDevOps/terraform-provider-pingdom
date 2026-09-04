package pingdom

import (
	"testing"
)

func TestSortString(t *testing.T) {
	cases := []struct {
		in, sep, want string
	}{
		{"z,a,m", ",", "a,m,z"},
		{"a", ",", "a"},
		{"", ",", ""},
		{"b,a", ",", "a,b"},
	}
	for _, c := range cases {
		if got := sortString(c.in, c.sep); got != c.want {
			t.Errorf("sortString(%q, %q) = %q, want %q", c.in, c.sep, got, c.want)
		}
	}
}

func TestCheckForResource_HTTP(t *testing.T) {
	d := resourcePingdomCheck().TestResourceData()
	d.Set("type", "http")
	d.Set("name", "example-http")
	d.Set("host", "example.com")
	d.Set("resolution", 5)
	d.Set("url", "/health")
	d.Set("encryption", true)
	d.Set("port", 443)
	d.Set("shouldcontain", "OK")
	d.Set("tags", "z,a,m")
	d.Set("verifycertificate", true)
	d.Set("ssldowndaysbefore", 7)

	check, err := checkForResource(d)
	if err != nil {
		t.Fatalf("checkForResource: %v", err)
	}

	http, ok := check.(*HttpCheck)
	if !ok {
		t.Fatalf("expected *HttpCheck, got %T", check)
	}

	if http.Name != "example-http" {
		t.Errorf("Name = %q, want %q", http.Name, "example-http")
	}
	if http.Hostname != "example.com" {
		t.Errorf("Hostname = %q, want %q", http.Hostname, "example.com")
	}
	if http.Resolution != 5 {
		t.Errorf("Resolution = %d, want 5", http.Resolution)
	}
	if http.Url != "/health" {
		t.Errorf("Url = %q, want %q", http.Url, "/health")
	}
	if !http.Encryption {
		t.Error("Encryption should be true")
	}
	if http.Port != 443 {
		t.Errorf("Port = %d, want 443", http.Port)
	}
	if http.ShouldContain != "OK" {
		t.Errorf("ShouldContain = %q, want %q", http.ShouldContain, "OK")
	}
	if http.Tags != "a,m,z" {
		t.Errorf("Tags = %q, want %q (sorted)", http.Tags, "a,m,z")
	}
	if http.VerifyCertificate == nil || !*http.VerifyCertificate {
		t.Error("VerifyCertificate should be true pointer")
	}
	if http.SSLDownDaysBefore == nil || *http.SSLDownDaysBefore != 7 {
		t.Errorf("SSLDownDaysBefore = %v, want pointer to 7", http.SSLDownDaysBefore)
	}
}

func TestCheckForResource_Ping(t *testing.T) {
	d := resourcePingdomCheck().TestResourceData()
	d.Set("type", "ping")
	d.Set("name", "example-ping")
	d.Set("host", "example.com")
	d.Set("resolution", 1)

	check, err := checkForResource(d)
	if err != nil {
		t.Fatalf("checkForResource: %v", err)
	}

	p, ok := check.(*PingCheck)
	if !ok {
		t.Fatalf("expected *PingCheck, got %T", check)
	}
	if p.Name != "example-ping" {
		t.Errorf("Name = %q, want %q", p.Name, "example-ping")
	}
	if p.Hostname != "example.com" {
		t.Errorf("Hostname = %q, want %q", p.Hostname, "example.com")
	}
	if p.Resolution != 1 {
		t.Errorf("Resolution = %d, want 1", p.Resolution)
	}
}

func TestCheckForResource_TCP(t *testing.T) {
	d := resourcePingdomCheck().TestResourceData()
	d.Set("type", "tcp")
	d.Set("name", "example-tcp")
	d.Set("host", "example.com")
	d.Set("resolution", 5)
	d.Set("port", 5432)
	d.Set("stringtosend", "PING\n")
	d.Set("stringtoexpect", "PONG")

	check, err := checkForResource(d)
	if err != nil {
		t.Fatalf("checkForResource: %v", err)
	}

	tcp, ok := check.(*TCPCheck)
	if !ok {
		t.Fatalf("expected *TCPCheck, got %T", check)
	}
	if tcp.Port != 5432 {
		t.Errorf("Port = %d, want 5432", tcp.Port)
	}
	if tcp.StringToSend != "PING\n" {
		t.Errorf("StringToSend = %q, want %q", tcp.StringToSend, "PING\n")
	}
	if tcp.StringToExpect != "PONG" {
		t.Errorf("StringToExpect = %q, want %q", tcp.StringToExpect, "PONG")
	}
}

func TestCheckForResource_HTTPClearAuthNotSetSpuriously(t *testing.T) {
	cases := map[string]string{
		"credentials present": "someone",
		"no credentials":      "",
	}
	for name, username := range cases {
		t.Run(name, func(t *testing.T) {
			d := resourcePingdomCheck().TestResourceData()
			d.Set("type", "http")
			d.Set("name", "example-http")
			d.Set("host", "example.com")
			d.Set("resolution", 5)
			if username != "" {
				d.Set("username", username)
				d.Set("password", "secret")
			}

			check, err := checkForResource(d)
			if err != nil {
				t.Fatalf("checkForResource: %v", err)
			}
			http, ok := check.(*HttpCheck)
			if !ok {
				t.Fatalf("expected *HttpCheck, got %T", check)
			}
			if http.ClearAuth {
				t.Error("ClearAuth must only be set when credentials are being removed")
			}
		})
	}
}

func TestCheckForResource_UnknownType(t *testing.T) {
	d := resourcePingdomCheck().TestResourceData()
	d.Set("type", "bogus")
	d.Set("name", "x")
	d.Set("host", "y")
	d.Set("resolution", 5)

	_, err := checkForResource(d)
	if err == nil {
		t.Fatal("expected error for unknown check type, got nil")
	}
}

func TestCheckForResource_HTTPHeaders(t *testing.T) {
	d := resourcePingdomCheck().TestResourceData()
	d.Set("type", "http")
	d.Set("name", "with-headers")
	d.Set("host", "example.com")
	d.Set("resolution", 5)
	d.Set("requestheaders", map[string]interface{}{
		"X-Foo":      "bar",
		"User-Agent": "custom-agent",
	})

	check, err := checkForResource(d)
	if err != nil {
		t.Fatalf("checkForResource: %v", err)
	}

	http := check.(*HttpCheck)
	if http.RequestHeaders["X-Foo"] != "bar" {
		t.Errorf("RequestHeaders[X-Foo] = %q, want bar", http.RequestHeaders["X-Foo"])
	}
	if http.RequestHeaders["User-Agent"] != "custom-agent" {
		t.Errorf("RequestHeaders[User-Agent] = %q, want custom-agent", http.RequestHeaders["User-Agent"])
	}
}

func TestCheckForResource_HTTPIntegrationAndTeams(t *testing.T) {
	d := resourcePingdomCheck().TestResourceData()
	d.Set("type", "http")
	d.Set("name", "x")
	d.Set("host", "y")
	d.Set("resolution", 5)
	if err := d.Set("integrationids", []interface{}{101, 102}); err != nil {
		t.Fatal(err)
	}
	if err := d.Set("teamids", []interface{}{201, 202}); err != nil {
		t.Fatal(err)
	}
	if err := d.Set("userids", []interface{}{301}); err != nil {
		t.Fatal(err)
	}

	check, err := checkForResource(d)
	if err != nil {
		t.Fatalf("checkForResource: %v", err)
	}
	http := check.(*HttpCheck)

	got := func(s []int) map[int]bool {
		m := map[int]bool{}
		for _, v := range s {
			m[v] = true
		}
		return m
	}

	if !got(http.IntegrationIds)[101] || !got(http.IntegrationIds)[102] {
		t.Errorf("IntegrationIds = %v, want {101, 102}", http.IntegrationIds)
	}
	if !got(http.TeamIds)[201] || !got(http.TeamIds)[202] {
		t.Errorf("TeamIds = %v, want {201, 202}", http.TeamIds)
	}
	if !got(http.UserIds)[301] {
		t.Errorf("UserIds = %v, want {301}", http.UserIds)
	}
}
