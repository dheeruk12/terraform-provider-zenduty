package zenduty

import (
	"errors"
	"fmt"
	"testing"
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

func TestIsRetryableError(t *testing.T) {
	apiErr := func(status string) error {
		return fmt.Errorf("POST APIs call to https://www.zenduty.com/api/account/teams/ failed: %s error: {}", status)
	}
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"bad request", apiErr("400 Bad Request"), false},
		{"unauthorized", apiErr("401 Unauthorized"), false},
		{"not found", apiErr("404 Not Found"), false},
		{"rate limited", apiErr("429 Too Many Requests"), true},
		{"server error", apiErr("500 Internal Server Error"), true},
		{"gateway timeout", apiErr("504 Gateway Timeout"), true},
		// transport failures never reached the API and are safe to retry
		{"transport", errors.New("dial tcp: i/o timeout"), true},
	}
	for _, c := range cases {
		if got := isRetryableError(c.err); got != c.want {
			t.Errorf("%s: isRetryableError() = %v, want %v", c.name, got, c.want)
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
		{"empty vs empty object", "", "{}", true},
		{"different values", `{"a":1}`, `{"a":2}`, false},
		{"both unparseable and equal", "nope", "nope", true},
		{"both unparseable and different", "nope", "nah", false},
	}
	for _, c := range cases {
		if got := suppressEquivalentJSONDiffs("conditions", c.old, c.new, nil); got != c.want {
			t.Errorf("%s: suppressEquivalentJSONDiffs(%q, %q) = %v, want %v", c.name, c.old, c.new, got, c.want)
		}
	}
}
