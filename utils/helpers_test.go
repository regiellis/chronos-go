package utils

import "testing"

func TestSanitizeString(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"Project_1", "Project_1"},
		{"Client-Name", "Client-Name"},
		{"Bad!@#$%^&*()", "Bad"},
		{"A B_C-D", "A B_C-D"},
		{"<script>alert(1)</script>", "scriptalert1script"},
	}
	for _, c := range cases {
		got := SanitizeString(c.in)
		if got != c.want {
			t.Errorf("SanitizeString(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestSanitizeDescription(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"Normal description.", "Normal description."},
		{"Line\nbreak", "Line break"},
		{"Control\x01\x02chars", "Controlchars"},
		{"<b>bold</b>", "<b>bold</b>"},
	}
	for _, c := range cases {
		got := SanitizeDescription(c.in)
		if got != c.want {
			t.Errorf("SanitizeDescription(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestValidateField(t *testing.T) {
	if !ValidateField("abc") {
		t.Error("ValidateField should return true for non-empty string")
	}
	if ValidateField("   ") {
		t.Error("ValidateField should return false for whitespace-only string")
	}
	if ValidateField("") {
		t.Error("ValidateField should return false for empty string")
	}
}

// Ensure file ends with a newline
