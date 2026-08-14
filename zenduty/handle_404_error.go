package zenduty

import (
	"log"
	"net/http"

	"github.com/Zenduty/zenduty-go-sdk/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// handleReadError converts an SDK error from a resource Read into
// diagnostics. A 404 means the object is gone server-side, so the resource is
// removed from state and Terraform plans to recreate it. The status comes
// from the SDK's typed *client.Error rather than string-matching diagnostics
// text, so an error message that merely mentions "404 Not Found" can no
// longer silently drop a live resource from state.
func handleReadError(d *schema.ResourceData, err error) diag.Diagnostics {
	if client.StatusCode(err) == http.StatusNotFound {
		log.Printf("[INFO] Removing %s from state because it no longer exists", d.Id())
		d.SetId("")
		return nil
	}
	return diag.FromErr(err)
}
