package pipass

import "fmt"

type Ledger interface {
	Log(path string, reason string, prev, next any)
}

type PrintLedger struct{}

func (l *PrintLedger) Log(path string, reason string, prev, next any) {
	fmt.Printf("[LEDGER] %s\n  Reason: %s\n  Delta:  %v -> %v\n\n", path, reason, prev, next)
}
