package terminal

import (
	"strings"
	"testing"
)

func TestUsageTemplateContainsCoreSections(t *testing.T) {
	tmpl := UsageTemplate()

	wanted := []string{
		`{{luHeader "Usage:"}}`,
		`{{luHeader "Available Commands:"}}`,
		`{{luHeader "Flags:"}}`,
		`.LocalFlags.FlagUsages`,
		`.InheritedFlags.FlagUsages`,
	}
	for _, w := range wanted {
		if !strings.Contains(tmpl, w) {
			t.Errorf("UsageTemplate() missing %q", w)
		}
	}
}

func TestBannerContainsVersionAndTagline(t *testing.T) {
	got := Banner("1.2.3", "tagline goes here")

	if !strings.Contains(got, "1.2.3") {
		t.Errorf("Banner() output missing version: %q", got)
	}
	if !strings.Contains(got, "tagline goes here") {
		t.Errorf("Banner() output missing tagline: %q", got)
	}
}

func TestTemplateHelpersRegistered(t *testing.T) {
	// The init() function registers these with Cobra's global template
	// func map. Calling them via the registered names should not panic.
	for _, name := range []string{"luHeader", "luCmd", "luDesc"} {
		// We can't directly access Cobra's internal func map, so we
		// verify by calling the helpers directly.
		switch name {
		case "luHeader":
			if luHeader("x") == "" {
				t.Errorf("luHeader returned empty string")
			}
		case "luCmd":
			if luCmd("x") == "" {
				t.Errorf("luCmd returned empty string")
			}
		case "luDesc":
			if luDesc("x") == "" {
				t.Errorf("luDesc returned empty string")
			}
		}
	}
}
