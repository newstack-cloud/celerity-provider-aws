//go:build integration

package e2e

import (
	"os"
	"testing"
)

// TestMain runs the suite, then sweeps for resources these tests created and failed to
// tear down. Nothing else checks teardown: a test asserts on what it deployed, never on
// what it left behind.
//
// A survivor fails the run, so the leak stays attached to the run that caused it rather
// than surfacing later as an unrelated run hitting an account limit. An already-failing
// suite keeps its own exit code, since a leak is usually downstream of that failure.
//
// The sweep is unsupervised: -timeout stops applying once m.Run returns, so a hang here
// would have no stack trace and no timeout. Hence its own budget, and waits that give up
// on their own. See runLeakSweep, which only deletes what this process named.
func TestMain(m *testing.M) {
	code := m.Run()

	if runLeakSweep() && code == 0 {
		code = 1
	}

	os.Exit(code)
}
