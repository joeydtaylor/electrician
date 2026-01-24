package forwardrelay

import "github.com/joeydtaylor/electrician/pkg/internal/types"

// ConnectInput registers one or more receivers as inputs for the relay.
func (fr *ForwardRelay[T]) ConnectInput(inputs ...types.Receiver[T]) {
	fr.requireNotFrozen("ConnectInput")

	if len(inputs) == 0 {
		return
	}

	n := 0
	for _, input := range inputs {
		if input != nil {
			inputs[n] = input
			n++
		}
	}
	if n == 0 {
		return
	}
	inputs = inputs[:n]

	fr.Input = append(fr.Input, inputs...)
	for _, input := range inputs {
		fr.NotifyLoggers(types.DebugLevel, "ConnectInput: added %s", input.GetComponentMetadata())
	}
}

// ConnectLogger registers one or more loggers.
func (fr *ForwardRelay[T]) ConnectLogger(loggers ...types.Logger) {
	fr.requireNotFrozen("ConnectLogger")

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

	fr.loggersLock.Lock()
	fr.Loggers = append(fr.Loggers, loggers...)
	fr.loggersLock.Unlock()
}
