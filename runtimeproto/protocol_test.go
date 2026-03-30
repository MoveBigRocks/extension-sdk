package runtimeproto

import "testing"

func TestSocketPathDefaultsAndSanitizesPackageKey(t *testing.T) {
	got := SocketPath("", "DemandOps/Sales Pipeline")
	want := "tmp/extensions/demandops_sales_pipeline.sock"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestInternalPathsTrimLeadingSlash(t *testing.T) {
	if got := InternalConsumerPath("/fleet.heartbeat"); got != "/__mbr/runtime/consumers/fleet.heartbeat" {
		t.Fatalf("unexpected consumer path %q", got)
	}
	if got := InternalJobPath("/daily-rollup"); got != "/__mbr/runtime/jobs/daily-rollup" {
		t.Fatalf("unexpected job path %q", got)
	}
}
