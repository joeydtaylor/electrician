package receivingrelay

import "github.com/joeydtaylor/electrician/pkg/internal/types"

// ConnectLogger registers loggers for the relay.
func (rr *ReceivingRelay[T]) ConnectLogger(loggers ...types.Logger) {
	rr.requireNotFrozen("ConnectLogger")

	if len(loggers) == 0 {
		return
	}

	n := 0
	for _, logger := range loggers {
		if logger != nil {
			loggers[n] = logger
			n++
		}
	}
	if n == 0 {
		return
	}
	loggers = loggers[:n]

	rr.loggersLock.Lock()
	rr.Loggers = append(rr.Loggers, loggers...)
	rr.loggersLock.Unlock()
}

// ConnectOutput registers downstream submitters.
func (rr *ReceivingRelay[T]) ConnectOutput(outputs ...types.Submitter[T]) {
	rr.requireNotFrozen("ConnectOutput")

	if len(outputs) == 0 {
		return
	}

	n := 0
	for _, output := range outputs {
		if output != nil {
			outputs[n] = output
			n++
		}
	}
	if n == 0 {
		return
	}
	outputs = outputs[:n]

	rr.Outputs = append(rr.Outputs, outputs...)
	for _, out := range outputs {
		rr.logKV(
			types.DebugLevel,
			"Output connected",
			"event", "ConnectOutput",
			"result", "SUCCESS",
			"output", out.GetComponentMetadata(),
		)
	}
}
