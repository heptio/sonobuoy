package status

import (
	"testing"

	"github.com/heptio/sonobuoy/pkg/plugin"
)

func TestCreateUpdater(t *testing.T) {
	expected := []plugin.ExpectedResult{
		{NodeName: "node1", ResultType: "systemd"},
		{NodeName: "node2", ResultType: "systemd"},
		{NodeName: "", ResultType: "e2e"},
	}

	updater := NewUpdater(expected)

	if err := updater.Receive(&Plugin{
		Status: Failed,
		Node:   "node1",
		Plugin: "systemd",
	}); err != nil {
		t.Errorf("unexpected error receiving update %v", err)
	}

	if updater.status.Status != Failed {
		t.Errorf("expected status to be failed, got %v", updater.status.Status)
	}
}
