package bench

import (
	"fmt"
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

func TestToSimResultPass(t *testing.T) {
	b := &Bench{}
	pass := true
	res := b.toSimResult(&sim.Result{Cycles: 42, Pass: &pass, Message: "PASS"}, nil)
	if !res.OK || res.Cycles != 42 || res.Pass == nil || !*res.Pass {
		t.Fatalf("unexpected %+v", res)
	}
}
