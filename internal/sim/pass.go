package sim

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// Expected is the hello-gpu PASS contract from templates/hello-gpu/expected.json.
type Expected struct {
	UARTContains string `json:"uart_contains"`
	LEDsNonZero  bool   `json:"leds_nonzero"`
	LEDsFinal    *int   `json:"leds_final"`
	PWMDuty      []int  `json:"pwm_duty"`
	FBSHA256     string `json:"fb_sha256"`
	Halted       bool   `json:"halted"`
}

// EvaluateExpected checks dumps against expected.json. PASS is numbers, not vibes.
func EvaluateExpected(exp Expected, uart string, leds int, pwm [4]int, fb []byte, halted bool) (bool, string) {
	pass := true
	var reasons []string
	if exp.UARTContains != "" && !strings.Contains(uart, exp.UARTContains) {
		pass = false
		reasons = append(reasons, "uart missing expected text")
	}
	if exp.LEDsNonZero && leds == 0 {
		pass = false
		reasons = append(reasons, "leds are zero")
	}
	if exp.LEDsFinal != nil && leds != *exp.LEDsFinal {
		pass = false
		reasons = append(reasons, "leds final value mismatch")
	}
	if len(exp.PWMDuty) == 4 {
		for i := 0; i < 4; i++ {
			if pwm[i] != exp.PWMDuty[i] {
				pass = false
				reasons = append(reasons, "pwm duty mismatch")
				break
			}
		}
	}
	if exp.FBSHA256 != "" {
		sum := sha256.Sum256(fb)
		got := hex.EncodeToString(sum[:])
		if !strings.EqualFold(got, exp.FBSHA256) {
			pass = false
			reasons = append(reasons, "framebuffer hash mismatch")
		}
	}
	if exp.Halted && !halted {
		pass = false
		reasons = append(reasons, "CPU did not halt")
	}
	msg := "PASS"
	if !pass {
		msg = "FAIL: " + strings.Join(reasons, "; ")
	}
	return pass, msg
}

const (
	OutcomePass = "pass"
	OutcomeFail = "fail"
	OutcomeRan  = "ran"
)

// ScoreRun maps dumps to a bench verdict.
//
//	PASS — hello-gpu golden matched
//	FAIL — the machine did not finish (never halted, or later failState)
//	RAN  — firmware halted; golden demo checks did not match (custom code)
func ScoreRun(exp Expected, uart string, leds int, pwm [4]int, fb []byte, halted bool) (pass *bool, outcome, msg string) {
	ok, evalMsg := EvaluateExpected(exp, uart, leds, pwm, fb, halted)
	if ok {
		t := true
		return &t, OutcomePass, "PASS"
	}
	if halted {
		reasons := strings.TrimPrefix(evalMsg, "FAIL: ")
		return nil, OutcomeRan, "RAN: " + reasons
	}
	f := false
	return &f, OutcomeFail, evalMsg
}
