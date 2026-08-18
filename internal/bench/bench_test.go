package bench

import (
	"fmt"
	"strings"
	"testing"

	"tiny-gpu-bench/internal/sim"
)

func TestToSimResultNilOnError(t *testing.T) {
	b := &Bench{}
	res := b.toSimResult(nil, fmt.Errorf("build exploded"))
	if res.OK {
		t.Fatal("expected not ok")
	}
	if res.Error != "build exploded" {
		t.Fatalf("error = %q", res.Error)
	}
	if res.Log != "" {
		t.Fatalf("log should be empty when result is nil, got %q", res.Log)
	}
}

func TestToSimResultBusy(t *testing.T) {
	b := &Bench{}
	res := b.toSimResult(nil, fmt.Errorf("busy"))
	if res.Error != "busy" {
		t.Fatalf("got %q", res.Error)
	}
}

func TestToSimResultRan(t *testing.T) {
	b := &Bench{}
	res := b.toSimResult(&sim.Result{Cycles: 17911, Outcome: sim.OutcomeRan, Message: "RAN: uart missing expected text"}, nil)
	if !res.OK || res.Pass != nil || res.Outcome != sim.OutcomeRan {
		t.Fatalf("unexpected %+v", res)
	}
	if strings.Contains(res.Message, "FAIL") {
		t.Fatalf("RAN message must not say FAIL: %q", res.Message)
	}
}
