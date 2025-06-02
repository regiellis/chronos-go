package cmd_test

import (
	"os/exec"
	"strings"
	"testing"
)

func runChronos(args ...string) (string, error) {
	cmd := exec.Command("go", append([]string{"run", "../main.go"}, args...)...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func TestBlockStartAndView(t *testing.T) {
	out, err := runChronos("block", "start", "Test Block", "--duration", "1h", "--client", "TestClient", "--project", "TestProject")
	if err != nil || !strings.Contains(out, "Started new block") {
		t.Fatalf("block start failed: %v\n%s", err, out)
	}
	out, err = runChronos("view", "block")
	if err != nil || !strings.Contains(out, "Active block") {
		t.Fatalf("view block failed: %v\n%s", err, out)
	}
}

func TestAddEntryAndList(t *testing.T) {
	out, err := runChronos("add", "30m today on Test Task -- test entry")
	if err != nil {
		t.Fatalf("add entry failed: %v\n%s", err, out)
	}
	out, err = runChronos("view", "list", "--project", "TestProject")
	if err != nil || !strings.Contains(out, "Test Task") {
		t.Fatalf("view list failed: %v\n%s", err, out)
	}
}

func TestInvoiceExport(t *testing.T) {
	out, err := runChronos("export", "invoice", "--format", "json")
	if err != nil || !strings.Contains(out, "total_amount") {
		t.Fatalf("export invoice failed: %v\n%s", err, out)
	}
}

func TestTemplateSaveAndUse(t *testing.T) {
	out, err := runChronos("template", "standup", "15m today on Standup -- Daily standup")
	if err != nil || !strings.Contains(out, "Template saved") {
		t.Fatalf("template save failed: %v\n%s", err, out)
	}
	out, err = runChronos("template", "standup")
	if err != nil || !strings.Contains(out, "Daily standup") {
		t.Fatalf("template use failed: %v\n%s", err, out)
	}
}

func TestSmartInvoice(t *testing.T) {
	out, err := runChronos("invoice-smart")
	if err != nil || !strings.Contains(out, "Marked") {
		t.Fatalf("invoice-smart failed: %v\n%s", err, out)
	}
}
