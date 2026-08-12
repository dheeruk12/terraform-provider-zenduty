package zenduty

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// The Zenduty API permits unnamed layers and overrides (reads return
// name: "" for them), so building the create payload must not reject them —
// otherwise imported schedules with unnamed layers could never be applied.
func TestBuildScheduleLayerAllowsEmptyName(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceSchedules().Schema, map[string]interface{}{
		"name":      "primary",
		"team_id":   "e09a815d-5c23-4575-a346-e25bdbd4dbdb",
		"time_zone": "UTC",
		"layers": []interface{}{
			map[string]interface{}{
				"name":                "",
				"shift_length":        86400,
				"rotation_start_time": "2022-11-26 21:31",
				"users":               []interface{}{"someuser"},
			},
		},
	})

	layers, diags := buildScheduleLayer(context.Background(), d, "UTC")
	if diags.HasError() {
		t.Fatalf("buildScheduleLayer returned error for empty layer name: %v", diags)
	}
	if len(layers) != 1 {
		t.Fatalf("expected 1 layer, got %d", len(layers))
	}
	if layers[0].Name != "" {
		t.Errorf("expected empty layer name to pass through, got %q", layers[0].Name)
	}
	if layers[0].ShiftLength != 86400 {
		t.Errorf("shift_length = %d, want 86400", layers[0].ShiftLength)
	}
}
