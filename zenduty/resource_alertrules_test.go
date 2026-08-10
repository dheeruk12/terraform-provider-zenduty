package zenduty

import (
	"reflect"
	"testing"

	"github.com/Zenduty/zenduty-go-sdk/client"
)

func TestExpandAlertRuleConditions(t *testing.T) {
	got := expandAlertRuleConditions([]interface{}{
		map[string]interface{}{"alert_condition_type": 2, "alert_field": "summary", "pattern": "CRITICAL"},
		map[string]interface{}{"alert_condition_type": 1, "alert_field": "alert_type", "pattern": ""},
	})
	want := []client.AlertRuleCondition{
		{AlertConditionType: 2, AlertField: "summary", Pattern: "CRITICAL", Position: 1},
		{AlertConditionType: 1, AlertField: "alert_type", Pattern: "", Position: 2},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("expandAlertRuleConditions() = %+v, want %+v", got, want)
	}
	if got := expandAlertRuleConditions(nil); len(got) != 0 {
		t.Errorf("expandAlertRuleConditions(nil) = %+v, want empty", got)
	}
}

func TestFlattenAlertRuleConditions(t *testing.T) {
	// API order is not guaranteed; flatten must return position order
	got := flattenAlertRuleConditions([]client.AlertRuleCondition{
		{UniqueID: "b", AlertConditionType: 1, AlertField: "alert_type", Pattern: "p2", Position: 2},
		{UniqueID: "a", AlertConditionType: 2, AlertField: "summary", Pattern: "p1", Position: 1},
	})
	want := []map[string]interface{}{
		{"unique_id": "a", "alert_condition_type": 2, "alert_field": "summary", "pattern": "p1"},
		{"unique_id": "b", "alert_condition_type": 1, "alert_field": "alert_type", "pattern": "p2"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("flattenAlertRuleConditions() = %+v, want %+v", got, want)
	}
}
