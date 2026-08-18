package sim

import (
	"strings"
	"testing"
)

func TestEvaluateExpectedPass(t *testing.T) {
	leds := 170
	exp := Expected{
		UARTContains: "hello from tiny-gpu",
		LEDsNonZero:  true,
		LEDsFinal:    &leds,
		PWMDuty:      []int{32, 96, 160, 224},
		FBSHA256:     "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		Halted:       true,
	}
	ok, msg := EvaluateExpected(exp, "hello from tiny-gpu\n", 170, [4]int{32, 96, 160, 224}, nil, true)
	if !ok {
		t.Fatalf("expected PASS, got %s", msg)
	}
	pass, outcome, smsg := ScoreRun(exp, "hello from tiny-gpu\n", 170, [4]int{32, 96, 160, 224}, nil, true)
	if pass == nil || !*pass || outcome != OutcomePass || smsg != "PASS" {
		t.Fatalf("ScoreRun pass=%v outcome=%q msg=%q", pass, outcome, smsg)
	}
}

func TestEvaluateExpectedFailReasons(t *testing.T) {
	exp := Expected{
		UARTContains: "hello from tiny-gpu",
		LEDsNonZero:  true,
		Halted:       true,
	}
	ok, msg := EvaluateExpected(exp, "nope", 0, [4]int{}, nil, false)
	if ok {
		t.Fatal("expected FAIL")
	}
	for _, want := range []string{"uart missing expected text", "leds are zero", "CPU did not halt"} {
		if !strings.Contains(msg, want) {
			t.Errorf("missing reason %q in %q", want, msg)
		}
	}
}

func TestScoreRunCustomHaltedIsRanNotFail(t *testing.T) {
	exp := Expected{
		UARTContains: "hello from tiny-gpu",
		LEDsNonZero:  true,
		LEDsFinal:    intPtr(170),
		PWMDuty:      []int{32, 96, 160, 224},
		FBSHA256:     "deadbeef",
		Halted:       true,
	}
	pass, outcome, msg := ScoreRun(exp, "custom lab2 program\n", 85, [4]int{10, 20, 30, 40}, nil, true)
	if pass != nil {
		t.Fatalf("custom halted should not set pass, got %v", *pass)
	}
	if outcome != OutcomeRan {
		t.Fatalf("outcome=%q", outcome)
	}
	if !strings.HasPrefix(msg, "RAN:") {
		t.Fatalf("msg=%q", msg)
	}
	if strings.Contains(msg, "FAIL") {
		t.Fatalf("must not say FAIL for custom halted code: %q", msg)
	}
}

func TestScoreRunNoHaltIsFail(t *testing.T) {
	exp := Expected{Halted: true}
	pass, outcome, msg := ScoreRun(exp, "", 0, [4]int{}, nil, false)
	if pass == nil || *pass {
		t.Fatal("expected FAIL")
	}
	if outcome != OutcomeFail {
		t.Fatalf("outcome=%q", outcome)
	}
	if !strings.Contains(msg, "CPU did not halt") {
		t.Fatalf("msg=%q", msg)
	}
}

func intPtr(n int) *int { return &n }
