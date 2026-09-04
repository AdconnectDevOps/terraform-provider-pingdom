// Package pingdom implements an internal client for the Pingdom REST API.
//
// This client replaces github.com/russellcardullo/go-pingdom (archived
// upstream, Feb 2023) with a maintained, dependency-free implementation. The
// exported surface deliberately mirrors the legacy library — same type names
// and service method shapes — so consumer code in this provider migrates with
// only an import change.
//
// Endpoint reference: https://docs.pingdom.com/api/
package pingdom

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

const defaultBaseURL = "https://api.pingdom.com/api/3.1"

// Client is a Pingdom API client. Construct via NewClientWithConfig.
type Client struct {
	APIToken string
	BaseURL  *url.URL

	httpClient *http.Client

	Checks   *CheckService
	Contacts *ContactService
	Teams    *TeamService
}

// ClientConfig configures a new Client.
type ClientConfig struct {
	APIToken   string
	BaseURL    string
	HTTPClient *http.Client
}

// NewClientWithConfig returns a new Pingdom API client.
func NewClientWithConfig(config ClientConfig) (*Client, error) {
	if config.APIToken == "" {
		return nil, fmt.Errorf("pingdom: APIToken required")
	}

	raw := config.BaseURL
	if raw == "" {
		raw = defaultBaseURL
	}
	base, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("pingdom: invalid BaseURL: %w", err)
	}

	hc := config.HTTPClient
	if hc == nil {
		hc = http.DefaultClient
	}

	c := &Client{
		APIToken:   config.APIToken,
		BaseURL:    base,
		httpClient: hc,
	}
	c.Checks = &CheckService{client: c}
	c.Contacts = &ContactService{client: c}
	c.Teams = &TeamService{client: c}
	return c, nil
}

// PingdomError is the structured error returned for non-2xx API responses.
// The type name matches the legacy go-pingdom library so existing error
// assertions (`*pingdom.PingdomError`) continue to work.
type PingdomError struct {
	StatusCode int    `json:"statuscode"`
	StatusDesc string `json:"statusdesc"`
	Message    string `json:"errormessage"`
}

func (e *PingdomError) Error() string {
	return fmt.Sprintf("%d %s: %s", e.StatusCode, e.StatusDesc, e.Message)
}

// PingdomResponse is the generic success-with-message reply used by some
// Pingdom endpoints (Update/Delete on checks and contacts).
type PingdomResponse struct {
	Message string `json:"message"`
}

// errorEnvelope wraps the Pingdom error response: `{"error":{...}}`.
type errorEnvelope struct {
	Error *PingdomError `json:"error"`
}

// newQueryRequest creates a request with params in the query string. This is
// the encoding Pingdom uses for /checks endpoints (POST and PUT both send
// params via query string, not body).
func (c *Client) newQueryRequest(method, path string, params map[string]string) (*http.Request, error) {
	u := *c.BaseURL
	u.Path = u.Path + path

	if len(params) > 0 {
		q := url.Values{}
		for k, v := range params {
			q.Set(k, v)
		}
		u.RawQuery = q.Encode()
	}

	req, err := http.NewRequest(method, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.APIToken)
	return req, nil
}

// newJSONRequest creates a request with a JSON body. Used for /alerting/*
// endpoints (contacts, teams).
func (c *Client) newJSONRequest(method, path, body string) (*http.Request, error) {
	u := *c.BaseURL
	u.Path = u.Path + path

	req, err := http.NewRequest(method, u.String(), strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.APIToken)
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

// do executes a request and unmarshals 2xx responses into out (when non-nil).
// Non-2xx responses are returned as *PingdomError.
func (c *Client) do(req *http.Request, out interface{}) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("pingdom: read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		env := errorEnvelope{}
		if jerr := json.Unmarshal(body, &env); jerr == nil && env.Error != nil {
			return env.Error
		}
		return &PingdomError{
			StatusCode: resp.StatusCode,
			StatusDesc: http.StatusText(resp.StatusCode),
			Message:    string(body),
		}
	}

	if out == nil {
		return nil
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("pingdom: decode response: %w", err)
	}
	return nil
}

// ---- Check types ---------------------------------------------------------

// Check is the marker interface for check payload variants (Http/Ping/TCP).
// Each variant implements postParams (for Create) and putParams (for Update).
type Check interface {
	postParams() map[string]string
	putParams() map[string]string
	valid() error
}

// HttpCheck represents an HTTP-type check payload.
type HttpCheck struct {
	Name                     string
	Hostname                 string
	Resolution               int
	Paused                   bool
	SendNotificationWhenDown int
	NotifyAgainEvery         int
	NotifyWhenBackup         bool
	Url                      string
	Encryption               bool
	Port                     int
	Username                 string
	Password                 string
	ShouldContain            string
	ShouldNotContain         string
	PostData                 string
	RequestHeaders           map[string]string
	IntegrationIds           []int
	ResponseTimeThreshold    int
	Tags                     string
	ProbeFilters             string
	UserIds                  []int
	TeamIds                  []int
	VerifyCertificate        *bool
	SSLDownDaysBefore        *int

	// ClearAuth asks putParams to send an empty auth= on update. Pingdom keeps
	// whatever basic auth a check already has unless the key is present in the
	// request, so dropping username/password from a configuration cannot take
	// effect without it. Only set when the credentials are actually being
	// removed - sending auth= on every update would change the request shape
	// for checks that never had auth.
	ClearAuth bool
}

// PingCheck represents an ICMP ping-type check payload.
type PingCheck struct {
	Name                     string
	Hostname                 string
	Resolution               int
	Paused                   bool
	SendNotificationWhenDown int
	NotifyAgainEvery         int
	NotifyWhenBackup         bool
	IntegrationIds           []int
	Tags                     string
	ResponseTimeThreshold    int
	ProbeFilters             string
	UserIds                  []int
	TeamIds                  []int
}

// TCPCheck represents a TCP-type check payload.
type TCPCheck struct {
	Name                     string
	Hostname                 string
	Resolution               int
	Paused                   bool
	SendNotificationWhenDown int
	NotifyAgainEvery         int
	NotifyWhenBackup         bool
	IntegrationIds           []int
	Tags                     string
	ProbeFilters             string
	UserIds                  []int
	TeamIds                  []int
	Port                     int
	StringToSend             string
	StringToExpect           string
}

// CheckResponse is the JSON shape returned by GET /checks/{id}.
type CheckResponse struct {
	ID                       int                 `json:"id"`
	Name                     string              `json:"name"`
	Resolution               int                 `json:"resolution,omitempty"`
	SendNotificationWhenDown int                 `json:"sendnotificationwhendown,omitempty"`
	NotifyAgainEvery         int                 `json:"notifyagainevery,omitempty"`
	NotifyWhenBackup         bool                `json:"notifywhenbackup,omitempty"`
	Created                  int64               `json:"created,omitempty"`
	Hostname                 string              `json:"hostname,omitempty"`
	Status                   string              `json:"status,omitempty"`
	Paused                   bool                `json:"paused,omitempty"`
	IntegrationIds           []int               `json:"integrationids,omitempty"`
	Type                     CheckResponseType   `json:"type,omitempty"`
	Tags                     []CheckResponseTag  `json:"tags,omitempty"`
	UserIds                  []int               `json:"userids,omitempty"`
	Teams                    []CheckTeamResponse `json:"teams,omitempty"`
	ResponseTimeThreshold    int                 `json:"responsetime_threshold,omitempty"`
	ProbeFilters             []string            `json:"probe_filters,omitempty"`

	// Backfilled from Teams[].ID — Pingdom does not return a flat teamids list.
	TeamIds []int `json:"-"`
}

// CheckTeamResponse is a team reference inside a CheckResponse.
type CheckTeamResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// CheckResponseTag is a tag entry on a check.
type CheckResponseTag struct {
	Name  string      `json:"name"`
	Type  string      `json:"type"`
	Count interface{} `json:"count"`
}

// CheckResponseType wraps the polymorphic `type` field on a CheckResponse:
// it arrives either as a bare string ("ping") or as a single-key object
// (`{"http":{...}}`, `{"tcp":{...}}`).
type CheckResponseType struct {
	Name string                    `json:"-"`
	HTTP *CheckResponseHTTPDetails `json:"http,omitempty"`
	TCP  *CheckResponseTCPDetails  `json:"tcp,omitempty"`
}

func (c *CheckResponseType) UnmarshalJSON(b []byte) error {
	var raw interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	switch v := raw.(type) {
	case string:
		c.Name = v
	case map[string]interface{}:
		if len(v) != 1 {
			return fmt.Errorf("check.type wrapper has %d keys, expected 1: %+v", len(v), v)
		}
		for k := range v {
			c.Name = k
		}
		type alias CheckResponseType
		var a alias
		if err := json.Unmarshal(b, &a); err != nil {
			return err
		}
		c.HTTP = a.HTTP
		c.TCP = a.TCP
	}
	return nil
}

// CheckResponseHTTPDetails are the HTTP-specific fields returned in
// CheckResponse.Type.HTTP.
type CheckResponseHTTPDetails struct {
	Url               string            `json:"url,omitempty"`
	Encryption        bool              `json:"encryption,omitempty"`
	Port              int               `json:"port,omitempty"`
	Username          string            `json:"username,omitempty"`
	Password          string            `json:"password,omitempty"`
	ShouldContain     string            `json:"shouldcontain,omitempty"`
	ShouldNotContain  string            `json:"shouldnotcontain,omitempty"`
	PostData          string            `json:"postdata,omitempty"`
	RequestHeaders    map[string]string `json:"requestheaders,omitempty"`
	VerifyCertificate bool              `json:"verify_certificate,omitempty"`
	SSLDownDaysBefore int               `json:"ssl_down_days_before,omitempty"`
}

// CheckResponseTCPDetails are the TCP-specific fields in CheckResponse.Type.TCP.
type CheckResponseTCPDetails struct {
	Port           int    `json:"port,omitempty"`
	StringToSend   string `json:"stringtosend,omitempty"`
	StringToExpect string `json:"stringtoexpect,omitempty"`
}

// ---- Check payload helpers -----------------------------------------------

func intsToCSV(ids []int) string {
	parts := make([]string, len(ids))
	for i, v := range ids {
		parts[i] = strconv.Itoa(v)
	}
	return strings.Join(parts, ",")
}

// putParams renders an HttpCheck for a PUT (Update) request. Empty strings
// are kept so Pingdom can clear server-side fields that no longer apply.
func (ck *HttpCheck) putParams() map[string]string {
	m := map[string]string{
		"name":             ck.Name,
		"host":             ck.Hostname,
		"resolution":       strconv.Itoa(ck.Resolution),
		"paused":           strconv.FormatBool(ck.Paused),
		"notifyagainevery": strconv.Itoa(ck.NotifyAgainEvery),
		"notifywhenbackup": strconv.FormatBool(ck.NotifyWhenBackup),
		"url":              ck.Url,
		"encryption":       strconv.FormatBool(ck.Encryption),
		"postdata":         ck.PostData,
		"integrationids":   intsToCSV(ck.IntegrationIds),
		"tags":             ck.Tags,
		"probe_filters":    ck.ProbeFilters,
		"userids":          intsToCSV(ck.UserIds),
		"teamids":          intsToCSV(ck.TeamIds),
	}
	if ck.Port != 0 {
		m["port"] = strconv.Itoa(ck.Port)
	}
	if ck.SendNotificationWhenDown != 0 {
		m["sendnotificationwhendown"] = strconv.Itoa(ck.SendNotificationWhenDown)
	}
	if ck.ResponseTimeThreshold != 0 {
		m["responsetime_threshold"] = strconv.Itoa(ck.ResponseTimeThreshold)
	}
	if ck.VerifyCertificate != nil {
		m["verify_certificate"] = strconv.FormatBool(*ck.VerifyCertificate)
	}
	if ck.SSLDownDaysBefore != nil {
		m["ssl_down_days_before"] = strconv.Itoa(*ck.SSLDownDaysBefore)
	}
	// ShouldContain and ShouldNotContain are mutually exclusive; pick whichever
	// is populated. We must send at least one so the server can clear the field
	// on a subsequent change.
	if ck.ShouldContain != "" {
		m["shouldcontain"] = ck.ShouldContain
	} else {
		m["shouldnotcontain"] = ck.ShouldNotContain
	}
	if ck.Username != "" {
		m["auth"] = fmt.Sprintf("%s:%s", ck.Username, ck.Password)
	} else if ck.ClearAuth {
		m["auth"] = ""
	}
	// Pingdom expects headers as requestheader0=key:value, requestheader1=...
	keys := make([]string, 0, len(ck.RequestHeaders))
	for k := range ck.RequestHeaders {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for i, k := range keys {
		m[fmt.Sprintf("requestheader%d", i)] = fmt.Sprintf("%s:%s", k, ck.RequestHeaders[k])
	}
	return m
}

// postParams renders an HttpCheck for a POST (Create) request. Empty strings
// are stripped (Pingdom rejects them at create time).
func (ck *HttpCheck) postParams() map[string]string {
	m := ck.putParams()
	for k, v := range m {
		if v == "" {
			delete(m, k)
		}
	}
	m["type"] = "http"
	return m
}

func (ck *HttpCheck) valid() error {
	if ck.Name == "" {
		return fmt.Errorf("HttpCheck.Name is required")
	}
	if ck.Hostname == "" {
		return fmt.Errorf("HttpCheck.Hostname is required")
	}
	switch ck.Resolution {
	case 1, 5, 15, 30, 60:
	default:
		return fmt.Errorf("HttpCheck.Resolution must be one of [1,5,15,30,60], got %d", ck.Resolution)
	}
	if ck.ShouldContain != "" && ck.ShouldNotContain != "" {
		return fmt.Errorf("HttpCheck: ShouldContain and ShouldNotContain are mutually exclusive")
	}
	return nil
}

func (ck *PingCheck) putParams() map[string]string {
	m := map[string]string{
		"name":             ck.Name,
		"host":             ck.Hostname,
		"resolution":       strconv.Itoa(ck.Resolution),
		"paused":           strconv.FormatBool(ck.Paused),
		"notifyagainevery": strconv.Itoa(ck.NotifyAgainEvery),
		"notifywhenbackup": strconv.FormatBool(ck.NotifyWhenBackup),
		"integrationids":   intsToCSV(ck.IntegrationIds),
		"probe_filters":    ck.ProbeFilters,
		"userids":          intsToCSV(ck.UserIds),
		"teamids":          intsToCSV(ck.TeamIds),
	}
	if ck.SendNotificationWhenDown != 0 {
		m["sendnotificationwhendown"] = strconv.Itoa(ck.SendNotificationWhenDown)
	}
	if ck.ResponseTimeThreshold != 0 {
		m["responsetime_threshold"] = strconv.Itoa(ck.ResponseTimeThreshold)
	}
	return m
}

func (ck *PingCheck) postParams() map[string]string {
	m := ck.putParams()
	for k, v := range m {
		if v == "" {
			delete(m, k)
		}
	}
	m["type"] = "ping"
	return m
}

func (ck *PingCheck) valid() error {
	if ck.Name == "" {
		return fmt.Errorf("PingCheck.Name is required")
	}
	if ck.Hostname == "" {
		return fmt.Errorf("PingCheck.Hostname is required")
	}
	switch ck.Resolution {
	case 1, 5, 15, 30, 60:
	default:
		return fmt.Errorf("PingCheck.Resolution must be one of [1,5,15,30,60], got %d", ck.Resolution)
	}
	return nil
}

func (ck *TCPCheck) putParams() map[string]string {
	m := map[string]string{
		"name":             ck.Name,
		"host":             ck.Hostname,
		"resolution":       strconv.Itoa(ck.Resolution),
		"paused":           strconv.FormatBool(ck.Paused),
		"notifyagainevery": strconv.Itoa(ck.NotifyAgainEvery),
		"notifywhenbackup": strconv.FormatBool(ck.NotifyWhenBackup),
		"integrationids":   intsToCSV(ck.IntegrationIds),
		"probe_filters":    ck.ProbeFilters,
		"tags":             ck.Tags,
		"userids":          intsToCSV(ck.UserIds),
		"teamids":          intsToCSV(ck.TeamIds),
		"port":             strconv.Itoa(ck.Port),
	}
	if ck.SendNotificationWhenDown != 0 {
		m["sendnotificationwhendown"] = strconv.Itoa(ck.SendNotificationWhenDown)
	}
	if ck.StringToSend != "" {
		m["stringtosend"] = ck.StringToSend
	}
	if ck.StringToExpect != "" {
		m["stringtoexpect"] = ck.StringToExpect
	}
	return m
}

func (ck *TCPCheck) postParams() map[string]string {
	m := ck.putParams()
	for k, v := range m {
		if v == "" {
			delete(m, k)
		}
	}
	m["type"] = "tcp"
	return m
}

func (ck *TCPCheck) valid() error {
	if ck.Name == "" {
		return fmt.Errorf("TCPCheck.Name is required")
	}
	if ck.Hostname == "" {
		return fmt.Errorf("TCPCheck.Hostname is required")
	}
	switch ck.Resolution {
	case 1, 5, 15, 30, 60:
	default:
		return fmt.Errorf("TCPCheck.Resolution must be one of [1,5,15,30,60], got %d", ck.Resolution)
	}
	if ck.Port < 1 {
		return fmt.Errorf("TCPCheck.Port must be ≥ 1, got %d", ck.Port)
	}
	return nil
}

// ---- Contact types -------------------------------------------------------

// NotificationTargets groups SMS and email notification targets for a Contact.
type NotificationTargets struct {
	SMS   []SMSNotification   `json:"sms,omitempty"`
	Email []EmailNotification `json:"email,omitempty"`
}

// SMSNotification is one SMS contact target.
type SMSNotification struct {
	CountryCode string `json:"country_code"`
	Number      string `json:"number"`
	Provider    string `json:"provider"`
	Severity    string `json:"severity"`
}

// EmailNotification is one email contact target.
type EmailNotification struct {
	Address  string `json:"address"`
	Severity string `json:"severity"`
}

// ContactTeam is a team reference returned inside a Contact response.
type ContactTeam struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Contact represents a Pingdom contact (used for both payload and response).
type Contact struct {
	ID                  int                 `json:"id,omitempty"`
	Name                string              `json:"name"`
	NotificationTargets NotificationTargets `json:"notification_targets"`
	Owner               bool                `json:"owner,omitempty"`
	Paused              bool                `json:"paused"`
	Teams               []ContactTeam       `json:"teams,omitempty"`
	Type                string              `json:"type,omitempty"`
}

// renderJSON serialises the create/update payload for Contact.
func (c *Contact) renderJSON() string {
	payload := map[string]interface{}{
		"name":                 c.Name,
		"notification_targets": c.NotificationTargets,
		"paused":               c.Paused,
	}
	b, _ := json.Marshal(payload)
	return string(b)
}

// ---- Team types ----------------------------------------------------------

// Team is the create/update payload for a Pingdom alerting team.
type Team struct {
	Name      string `json:"name"`
	MemberIDs []int  `json:"member_ids,omitempty"`
}

// renderJSON serialises a Team for POST/PUT requests.
func (t *Team) renderJSON() string {
	b, _ := json.Marshal(map[string]interface{}{
		"name":       t.Name,
		"member_ids": t.MemberIDs,
	})
	return string(b)
}

// TeamResponse is the JSON shape returned by GET /alerting/teams/{id}.
type TeamResponse struct {
	ID      int                  `json:"id"`
	Name    string               `json:"name,omitempty"`
	Members []TeamMemberResponse `json:"members,omitempty"`
}

// TeamMemberResponse is a contact reference inside a TeamResponse.
type TeamMemberResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name,omitempty"`
	Type string `json:"type,omitempty"`
}

// TeamDeleteResponse mirrors the legacy library's delete response wrapper.
type TeamDeleteResponse struct {
	Message string `json:"message"`
}

// ---- Service envelopes (for JSON wrappers) -------------------------------

type checkDetailsResp struct {
	Check *CheckResponse `json:"check"`
}

type listChecksResp struct {
	Checks []CheckResponse `json:"checks"`
}

type contactDetailsResp struct {
	Contact *Contact `json:"contact"`
}

type listContactsResp struct {
	Contacts []Contact `json:"contacts"`
}

type teamDetailsResp struct {
	Team *TeamResponse `json:"team"`
}

type listTeamsResp struct {
	Teams []TeamResponse `json:"teams"`
}

// ---- CheckService --------------------------------------------------------

// CheckService gives access to /checks endpoints.
type CheckService struct {
	client *Client
}

// List returns all checks.
func (s *CheckService) List() ([]CheckResponse, error) {
	req, err := s.client.newQueryRequest("GET", "/checks", nil)
	if err != nil {
		return nil, err
	}
	out := &listChecksResp{}
	if err := s.client.do(req, out); err != nil {
		return nil, err
	}
	return out.Checks, nil
}

// Read fetches a single check by ID. Pingdom does not return a flat teamids
// field, so the response's TeamIds slice is backfilled from Teams[].ID.
func (s *CheckService) Read(id int) (*CheckResponse, error) {
	req, err := s.client.newQueryRequest("GET", "/checks/"+strconv.Itoa(id), map[string]string{"include_teams": "true"})
	if err != nil {
		return nil, err
	}
	out := &checkDetailsResp{}
	if err := s.client.do(req, out); err != nil {
		return nil, err
	}
	if out.Check == nil {
		return nil, fmt.Errorf("pingdom: empty check in response")
	}
	out.Check.TeamIds = make([]int, len(out.Check.Teams))
	for i, t := range out.Check.Teams {
		out.Check.TeamIds[i] = t.ID
	}
	return out.Check, nil
}

// Create creates a new check.
func (s *CheckService) Create(check Check) (*CheckResponse, error) {
	if err := check.valid(); err != nil {
		return nil, err
	}
	req, err := s.client.newQueryRequest("POST", "/checks", check.postParams())
	if err != nil {
		return nil, err
	}
	out := &checkDetailsResp{}
	if err := s.client.do(req, out); err != nil {
		return nil, err
	}
	return out.Check, nil
}

// Update applies new field values to an existing check.
func (s *CheckService) Update(id int, check Check) (*PingdomResponse, error) {
	if err := check.valid(); err != nil {
		return nil, err
	}
	req, err := s.client.newQueryRequest("PUT", "/checks/"+strconv.Itoa(id), check.putParams())
	if err != nil {
		return nil, err
	}
	out := &PingdomResponse{}
	if err := s.client.do(req, out); err != nil {
		return nil, err
	}
	return out, nil
}

// Delete removes a check.
func (s *CheckService) Delete(id int) (*PingdomResponse, error) {
	req, err := s.client.newQueryRequest("DELETE", "/checks/"+strconv.Itoa(id), nil)
	if err != nil {
		return nil, err
	}
	out := &PingdomResponse{}
	if err := s.client.do(req, out); err != nil {
		return nil, err
	}
	return out, nil
}

// ---- ContactService ------------------------------------------------------

// ContactService gives access to /alerting/contacts endpoints.
type ContactService struct {
	client *Client
}

// List returns all contacts.
func (s *ContactService) List() ([]Contact, error) {
	req, err := s.client.newQueryRequest("GET", "/alerting/contacts", nil)
	if err != nil {
		return nil, err
	}
	out := &listContactsResp{}
	if err := s.client.do(req, out); err != nil {
		return nil, err
	}
	return out.Contacts, nil
}

// Read fetches a single contact.
func (s *ContactService) Read(id int) (*Contact, error) {
	req, err := s.client.newQueryRequest("GET", "/alerting/contacts/"+strconv.Itoa(id), nil)
	if err != nil {
		return nil, err
	}
	out := &contactDetailsResp{}
	if err := s.client.do(req, out); err != nil {
		return nil, err
	}
	if out.Contact == nil {
		return nil, fmt.Errorf("pingdom: empty contact in response")
	}
	return out.Contact, nil
}

// Create creates a new contact.
func (s *ContactService) Create(c *Contact) (*Contact, error) {
	if c.Name == "" {
		return nil, fmt.Errorf("Contact.Name is required")
	}
	req, err := s.client.newJSONRequest("POST", "/alerting/contacts", c.renderJSON())
	if err != nil {
		return nil, err
	}
	out := &contactDetailsResp{}
	if err := s.client.do(req, out); err != nil {
		return nil, err
	}
	return out.Contact, nil
}

// Update applies new values to an existing contact.
func (s *ContactService) Update(id int, c *Contact) (*PingdomResponse, error) {
	if c.Name == "" {
		return nil, fmt.Errorf("Contact.Name is required")
	}
	req, err := s.client.newJSONRequest("PUT", "/alerting/contacts/"+strconv.Itoa(id), c.renderJSON())
	if err != nil {
		return nil, err
	}
	out := &PingdomResponse{}
	if err := s.client.do(req, out); err != nil {
		return nil, err
	}
	return out, nil
}

// Delete removes a contact.
func (s *ContactService) Delete(id int) (*PingdomResponse, error) {
	req, err := s.client.newQueryRequest("DELETE", "/alerting/contacts/"+strconv.Itoa(id), nil)
	if err != nil {
		return nil, err
	}
	out := &PingdomResponse{}
	if err := s.client.do(req, out); err != nil {
		return nil, err
	}
	return out, nil
}

// ---- TeamService ---------------------------------------------------------

// TeamService gives access to /alerting/teams endpoints.
type TeamService struct {
	client *Client
}

// List returns all teams.
func (s *TeamService) List() ([]TeamResponse, error) {
	req, err := s.client.newQueryRequest("GET", "/alerting/teams", nil)
	if err != nil {
		return nil, err
	}
	out := &listTeamsResp{}
	if err := s.client.do(req, out); err != nil {
		return nil, err
	}
	return out.Teams, nil
}

// Read fetches a single team.
func (s *TeamService) Read(id int) (*TeamResponse, error) {
	req, err := s.client.newQueryRequest("GET", "/alerting/teams/"+strconv.Itoa(id), nil)
	if err != nil {
		return nil, err
	}
	out := &teamDetailsResp{}
	if err := s.client.do(req, out); err != nil {
		return nil, err
	}
	if out.Team == nil {
		return nil, fmt.Errorf("pingdom: empty team in response")
	}
	return out.Team, nil
}

// Create creates a new team.
func (s *TeamService) Create(t *Team) (*TeamResponse, error) {
	if t.Name == "" {
		return nil, fmt.Errorf("Team.Name is required")
	}
	req, err := s.client.newJSONRequest("POST", "/alerting/teams", t.renderJSON())
	if err != nil {
		return nil, err
	}
	out := &teamDetailsResp{}
	if err := s.client.do(req, out); err != nil {
		return nil, err
	}
	return out.Team, nil
}

// Update applies new values to an existing team.
func (s *TeamService) Update(id int, t *Team) (*TeamResponse, error) {
	req, err := s.client.newJSONRequest("PUT", "/alerting/teams/"+strconv.Itoa(id), t.renderJSON())
	if err != nil {
		return nil, err
	}
	out := &teamDetailsResp{}
	if err := s.client.do(req, out); err != nil {
		return nil, err
	}
	return out.Team, nil
}

// Delete removes a team.
func (s *TeamService) Delete(id int) (*TeamDeleteResponse, error) {
	req, err := s.client.newQueryRequest("DELETE", "/alerting/teams/"+strconv.Itoa(id), nil)
	if err != nil {
		return nil, err
	}
	out := &TeamDeleteResponse{}
	if err := s.client.do(req, out); err != nil {
		return nil, err
	}
	return out, nil
}
