package pingdom

import (
	"strings"
	"testing"
)

func TestGetNotificationMethods_HighAndLow(t *testing.T) {
	d := resourcePingdomContact().TestResourceData()
	d.Set("sms_notification", []interface{}{
		map[string]interface{}{
			"number":       "1234567890",
			"country_code": "1",
			"severity":     "HIGH",
			"provider":     "nexmo",
		},
	})
	d.Set("email_notification", []interface{}{
		map[string]interface{}{
			"address":  "alerts@example.com",
			"severity": "LOW",
		},
	})

	got, err := getNotificationMethods(d)
	if err != nil {
		t.Fatalf("getNotificationMethods: %v", err)
	}
	if len(got.SMS) != 1 || got.SMS[0].Number != "1234567890" {
		t.Errorf("SMS = %v, want one entry with Number=1234567890", got.SMS)
	}
	if len(got.Email) != 1 || got.Email[0].Address != "alerts@example.com" {
		t.Errorf("Email = %v, want one entry with Address=alerts@example.com", got.Email)
	}
}

func TestGetNotificationMethods_MissingHighSeverity(t *testing.T) {
	d := resourcePingdomContact().TestResourceData()
	// Only LOW configured; HIGH missing → must error.
	d.Set("email_notification", []interface{}{
		map[string]interface{}{
			"address":  "alerts@example.com",
			"severity": "LOW",
		},
	})

	_, err := getNotificationMethods(d)
	if err == nil {
		t.Fatal("expected error for missing HIGH severity, got nil")
	}
	if !strings.Contains(err.Error(), "high and low severity") {
		t.Errorf("error %q should mention high and low severity requirement", err.Error())
	}
}

func TestGetNotificationMethods_MissingLowSeverity(t *testing.T) {
	d := resourcePingdomContact().TestResourceData()
	d.Set("sms_notification", []interface{}{
		map[string]interface{}{
			"number":       "1234567890",
			"country_code": "1",
			"severity":     "HIGH",
			"provider":     "nexmo",
		},
	})
	// No LOW anywhere → must error.

	_, err := getNotificationMethods(d)
	if err == nil {
		t.Fatal("expected error for missing LOW severity, got nil")
	}
}

func TestGetNotificationMethods_InvalidProvider(t *testing.T) {
	d := resourcePingdomContact().TestResourceData()
	d.Set("sms_notification", []interface{}{
		map[string]interface{}{
			"number":       "1234567890",
			"country_code": "1",
			"severity":     "HIGH",
			"provider":     "twilio", // not in the allow-list
		},
	})

	_, err := getNotificationMethods(d)
	if err == nil {
		t.Fatal("expected error for unsupported SMS provider, got nil")
	}
	if !strings.Contains(err.Error(), "SMS provider must be one of") {
		t.Errorf("error %q should mention allowed providers", err.Error())
	}
}

func TestContactForResource_BuildsName(t *testing.T) {
	d := resourcePingdomContact().TestResourceData()
	d.Set("name", "oncall")
	d.Set("sms_notification", []interface{}{
		map[string]interface{}{
			"number": "1234", "country_code": "1", "severity": "HIGH", "provider": "nexmo",
		},
	})
	d.Set("email_notification", []interface{}{
		map[string]interface{}{
			"address": "x@y.com", "severity": "LOW",
		},
	})

	got, err := contactForResource(d)
	if err != nil {
		t.Fatalf("contactForResource: %v", err)
	}
	if got.Name != "oncall" {
		t.Errorf("Name = %q, want oncall", got.Name)
	}
	if len(got.NotificationTargets.SMS) != 1 {
		t.Errorf("NotificationTargets.SMS = %v, want 1 entry", got.NotificationTargets.SMS)
	}
	if len(got.NotificationTargets.Email) != 1 {
		t.Errorf("NotificationTargets.Email = %v, want 1 entry", got.NotificationTargets.Email)
	}
}

func TestContactForResource_Paused(t *testing.T) {
	d := resourcePingdomContact().TestResourceData()
	d.Set("name", "p")
	d.Set("paused", true)
	d.Set("sms_notification", []interface{}{
		map[string]interface{}{
			"number": "1", "country_code": "1", "severity": "HIGH", "provider": "nexmo",
		},
	})
	d.Set("email_notification", []interface{}{
		map[string]interface{}{"address": "x@y.com", "severity": "LOW"},
	})

	got, err := contactForResource(d)
	if err != nil {
		t.Fatalf("contactForResource: %v", err)
	}
	if !got.Paused {
		t.Error("Paused = false, want true")
	}
}
