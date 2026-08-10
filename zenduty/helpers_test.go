package zenduty

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/Zenduty/zenduty-go-sdk/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestIsValidUUID(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"1cd4b3f5-5f4c-4d1e-9c2a-0a1b2c3d4e5f", true},
		// any UUID version, not just v4
		{"1cd4b3f5-5f4c-1d1e-9c2a-0a1b2c3d4e5f", true},
		// the old validator's char class literally matched "|"
		{"1cd4b3f5-5f4c-4d1e-|c2a-0a1b2c3d4e5f", false},
		// a username is not a full UUID
		{"1cd4b3f5-5f4c-4d1e-9c2a-0", false},
		{"", false},
	}
	for _, c := range cases {
		if got := IsValidUUID(c.in); got != c.want {
			t.Errorf("IsValidUUID(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestIsValidUsername(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		// usernames are str(uuid4())[:25]
		{"1cd4b3f5-5f4c-4d1e-9c2a-0", true},
		{"1cd4b3f5-5f4c-1d1e-9c2a-0", true},
		// a full UUID is too long
		{"1cd4b3f5-5f4c-4d1e-9c2a-0a1b2c3d4e5f", false},
		{"1cd4b3f5-5f4c-4d1e-9c2a", false},
		{"", false},
	}
	for _, c := range cases {
		if got := IsValidUsername(c.in); got != c.want {
			t.Errorf("IsValidUsername(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestIsJSONString(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{`{"a":1}`, true},
		// arrays are valid JSON too; the old check rejected them
		{`[{"a":1}]`, true},
		{`"a"`, true},
		{`{`, false},
		{"", false},
	}
	for _, c := range cases {
		if got := isJSONString(c.in); got != c.want {
			t.Errorf("isJSONString(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

// apiErr builds the typed error the SDK returns for a non-2xx response.
func apiErr(status int) error {
	return &client.Error{
		Code: status,
		ErrorResponse: &client.Response{
			Response: &http.Response{
				StatusCode: status,
				Status:     fmt.Sprintf("%d %s", status, http.StatusText(status)),
				Request: &http.Request{
					Method: "GET",
					URL:    &url.URL{Scheme: "https", Host: "www.zenduty.com", Path: "/api/account/teams/"},
				},
			},
		},
	}
}

func TestIsRetryableError(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"bad request", apiErr(400), false},
		{"unauthorized", apiErr(401), false},
		{"not found", apiErr(404), false},
		{"rate limited", apiErr(429), true},
		{"server error", apiErr(500), true},
		{"gateway timeout", apiErr(504), true},
		// transport failures never reached the API and are safe to retry
		{"transport", errors.New("dial tcp: i/o timeout"), true},
	}
	for _, c := range cases {
		if got := isRetryableError(c.err); got != c.want {
			t.Errorf("%s: isRetryableError() = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestHandleReadError(t *testing.T) {
	testSchema := map[string]*schema.Schema{
		"name": {Type: schema.TypeString, Optional: true},
	}

	cases := []struct {
		name        string
		err         error
		wantErr     bool
		wantCleared bool
	}{
		{"404 clears state without error", apiErr(404), false, true},
		{"other API errors surface", apiErr(500), true, false},
		{"transport errors surface", errors.New("dial tcp: i/o timeout"), true, false},
		// the old string-matching wrapper would have dropped this live
		// resource from state
		{"404 mentioned in message only", errors.New("upstream said 404 Not Found"), true, false},
	}
	for _, c := range cases {
		d := schema.TestResourceDataRaw(t, testSchema, map[string]interface{}{})
		d.SetId("some-id")
		diags := handleReadError(d, c.err)
		if diags.HasError() != c.wantErr {
			t.Errorf("%s: HasError() = %v, want %v", c.name, diags.HasError(), c.wantErr)
		}
		if gotCleared := d.Id() == ""; gotCleared != c.wantCleared {
			t.Errorf("%s: id cleared = %v, want %v", c.name, gotCleared, c.wantCleared)
		}
	}
}

func TestPadTimeOfDay(t *testing.T) {
	cases := []struct{ in, want string }{
		{"9:5:3", "09:05:03"},
		{"09:05:03", "09:05:03"},
		{"23:59:59", "23:59:59"},
		{"not a time", "not a time"},
	}
	for _, c := range cases {
		if got := padTimeOfDay(c.in); got != c.want {
			t.Errorf("padTimeOfDay(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestSuppressEquivalentJSONDiffs(t *testing.T) {
	cases := []struct {
		name     string
		old, new string
		want     bool
	}{
		{"formatting only", `{"a":1,"b":2}`, "{\n  \"a\": 1,\n  \"b\": 2\n}", true},
		{"key order only", `{"a":1,"b":2}`, `{"b":2,"a":1}`, true},
		{"different values", `{"a":1}`, `{"a":2}`, false},
		// unset is NOT the same as an empty object here: suppressing that
		// would stop rule_json from ever being written
		{"empty vs empty object", "", "{}", false},
		{"both unparseable and equal", "nope", "nope", true},
		{"both unparseable and different", "nope", "nah", false},
	}
	for _, c := range cases {
		if got := suppressEquivalentJSONDiffs("rule_json", c.old, c.new, nil); got != c.want {
			t.Errorf("%s: suppressEquivalentJSONDiffs(%q, %q) = %v, want %v", c.name, c.old, c.new, got, c.want)
		}
	}
}

func TestSuppressEquivalentJSONObjectDiffs(t *testing.T) {
	cases := []struct {
		name     string
		old, new string
		want     bool
	}{
		// the backend stores "{}" as the empty value, so these are the same
		{"empty vs empty object", "", "{}", true},
		{"empty object vs empty", "{}", "", true},
		{"formatting only", `{"a":1}`, "{\n  \"a\": 1\n}", true},
		{"empty vs non-empty", "", `{"a":1}`, false},
		{"different values", `{"a":1}`, `{"a":2}`, false},
	}
	for _, c := range cases {
		if got := suppressEquivalentJSONObjectDiffs("conditions", c.old, c.new, nil); got != c.want {
			t.Errorf("%s: suppressEquivalentJSONObjectDiffs(%q, %q) = %v, want %v", c.name, c.old, c.new, got, c.want)
		}
	}
}
