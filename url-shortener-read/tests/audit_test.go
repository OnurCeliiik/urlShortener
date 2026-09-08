package tests

import (
	"OnurCeliiik/urlShortener-read/internal/audit"
	"testing"
)

func TestAuditNoopDoesNotBlock(t *testing.T) {
	emitter := audit.Noop()
	emitter.Emit(audit.Event{Action: "resolve", Status: 302})
	emitter.Close()
}
