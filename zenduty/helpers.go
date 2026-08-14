package zenduty

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/mail"
	"regexp"
	"strings"

	"github.com/Zenduty/zenduty-go-sdk/client"
	"github.com/google/uuid"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func isJSONString(s string) bool {
	var js interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}

func IsValidUUID(uuid string) bool {
	r := regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$")
	return r.MatchString(uuid)
}

// IsValidUsername matches Zenduty usernames, which are the first 25 characters
// of a UUID (8-4-4-4-1) — not a full UUID.
func IsValidUsername(username string) bool {
	r := regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]$")
	return r.MatchString(username)
}

func ValidateUserName() schema.SchemaValidateDiagFunc {
	return func(v interface{}, path cty.Path) diag.Diagnostics {
		var diags diag.Diagnostics
		id, ok := v.(string)
		if !ok {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Invalid",
				Detail:   "expected type of string",
			})
		}
		if !IsValidUsername(id) {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Invalid username",
				Detail:   fmt.Sprintf("expected %s to be a valid Zenduty username (the 25-character prefix of a UUID)", path),
			})
		}

		return diags
	}
}

func emptyString(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}

// isRetryableError reports whether an SDK error is worth retrying: 5xx, 429,
// or transport-level failures (timeouts, connection resets). A transport
// error carries no HTTP status because it never reached the API, and is safe
// to retry.
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	if code := client.StatusCode(err); code != 0 {
		return code == http.StatusTooManyRequests || code >= 500
	}
	return true
}

func ValidateUUID() schema.SchemaValidateDiagFunc {
	return func(v interface{}, path cty.Path) diag.Diagnostics {
		var diags diag.Diagnostics
		id, ok := v.(string)
		if !ok {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Invalid",
				Detail:   "expected type of string",
			})
		}
		if !IsValidUUID(id) {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Invalid ID",
				Detail:   fmt.Sprintf("expected %s to be a valid UUID", path),
			})
		}

		return diags
	}
}

func isEmailValid(e string) bool {
	emailAddress, err := mail.ParseAddress(e)
	return err == nil && emailAddress.Address == e
}

func ValidateEmail() schema.SchemaValidateDiagFunc {
	return func(v interface{}, path cty.Path) diag.Diagnostics {
		var diags diag.Diagnostics
		id, ok := v.(string)
		if !ok {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Invalid",
				Detail:   "expected type of string",
			})
		}
		if !isEmailValid(id) {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Invalid Email",
				Detail:   fmt.Sprintf("expected %s to be a valid Email", path),
			})
		}

		return diags
	}

}

func ValidateRequired() schema.SchemaValidateDiagFunc {
	return func(v interface{}, path cty.Path) diag.Diagnostics {
		var diags diag.Diagnostics
		id, ok := v.(string)
		if !ok {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Invalid",
				Detail:   "expected type of string",
			})
		}
		if emptyString(id) {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "This field is required",
				Detail:   "This field cannot be empty",
			})
		}

		return diags
	}

}

// ValidateNonZeroLength rejects only zero-length strings. Unlike
// ValidateRequired it accepts whitespace-only values, which the Zenduty API
// permits for some fields (e.g. a user's last name) — rejecting them would
// make real remote state unrepresentable in config.
func ValidateNonZeroLength() schema.SchemaValidateDiagFunc {
	return func(v interface{}, path cty.Path) diag.Diagnostics {
		var diags diag.Diagnostics
		s, ok := v.(string)
		if !ok {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Invalid",
				Detail:   "expected type of string",
			})
			return diags
		}
		if len(s) == 0 {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "This field is required",
				Detail:   "This field cannot be empty",
			})
		}
		return diags
	}
}

func generateUUID() string {
	id := uuid.New()
	uuidString := id.String()
	return uuidString
}

func normalizeJSON(jsonString string) (string, error) {
	var jsonData interface{}
	if err := json.Unmarshal([]byte(jsonString), &jsonData); err != nil {
		return "", err
	}

	normalizedBytes, err := json.Marshal(jsonData)
	if err != nil {
		return "", err
	}

	return string(normalizedBytes), nil
}

// suppressEquivalentJSONDiffs suppresses diffs between JSON strings that
// decode to the same value. Reads store the API's compact, key-sorted JSON, so
// without this any human-formatted config diffs on every plan.
func suppressEquivalentJSONDiffs(k, old, new string, d *schema.ResourceData) bool {
	normalizedOld, errOld := normalizeJSON(old)
	normalizedNew, errNew := normalizeJSON(new)
	if errOld != nil || errNew != nil {
		return old == new
	}
	return normalizedOld == normalizedNew
}

// suppressEquivalentJSONObjectDiffs is suppressEquivalentJSONDiffs for fields
// whose backend default is an empty object, where an unset value and "{}" mean
// the same thing.
func suppressEquivalentJSONObjectDiffs(k, old, new string, d *schema.ResourceData) bool {
	if old == "" {
		old = "{}"
	}
	if new == "" {
		new = "{}"
	}
	return suppressEquivalentJSONDiffs(k, old, new, d)
}
