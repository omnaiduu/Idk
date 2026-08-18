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
