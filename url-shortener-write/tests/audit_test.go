package tests

import (
	"OnurCeliiik/urlShortener-write/internal/audit"
	"testing"
)

func TestAuditNoopDoesNotBlock(t *testing.T) {
	emitter := audit.Noop()
	emitter.Emit(audit.Event{Action: "shorten", Status: 201})
	emitter.Close()
}
