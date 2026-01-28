package plug

import "github.com/joeydtaylor/electrician/pkg/internal/types"

// AddAdapterFunc registers adapter funcs.
// Panics if called after Freeze.
func (p *Plug[T]) AddAdapterFunc(funcs ...types.AdapterFunc[T]) {
	p.requireNotFrozen("AddAdapterFunc")

	if len(funcs) == 0 {
		return
	}

	n := 0
	for _, fn := range funcs {
		if fn != nil {
			funcs[n] = fn
			n++
		}
	}
	if n == 0 {
		return
	}
	funcs = funcs[:n]

	p.configLock.Lock()
	p.adapterFns = append(p.adapterFns, funcs...)
	p.configLock.Unlock()
}

// ConnectAdapter registers adapters.
// Panics if called after Freeze.
func (p *Plug[T]) ConnectAdapter(adapters ...types.Adapter[T]) {
	p.requireNotFrozen("ConnectAdapter")

	if len(adapters) == 0 {
		return
	}

	n := 0
	for _, a := range adapters {
		if a != nil {
			adapters[n] = a
			n++
		}
	}
	if n == 0 {
		return
	}
	adapters = adapters[:n]

	p.configLock.Lock()
	p.adapters = append(p.adapters, adapters...)
	p.configLock.Unlock()
}

// ConnectSensor registers sensors for the plug.
// Panics if called after Freeze.
func (p *Plug[T]) ConnectSensor(sensors ...types.Sensor[T]) {
	p.requireNotFrozen("ConnectSensor")

	if len(sensors) == 0 {
		return
	}

	n := 0
	for _, s := range sensors {
		if s != nil {
			sensors[n] = s
			n++
		}
	}
	if n == 0 {
		return
	}
	sensors = sensors[:n]

	p.sensors = append(p.sensors, sensors...)
	for _, s := range sensors {
		p.NotifyLoggers(
			types.DebugLevel,
			"ConnectSensor",
			"component", p.componentMetadata,
			"event", "ConnectSensor",
			"sensor", s.GetComponentMetadata(),
		)
	}
}

// ConnectLogger registers loggers for the plug.
// Panics if called after Freeze.
func (p *Plug[T]) ConnectLogger(loggers ...types.Logger) {
	p.requireNotFrozen("ConnectLogger")

	if len(loggers) == 0 {
		return
	}

	n := 0
	for _, l := range loggers {
		if l != nil {
			loggers[n] = l
			n++
		}
	}
	if n == 0 {
		return
	}
	loggers = loggers[:n]

	p.loggersLock.Lock()
	p.loggers = append(p.loggers, loggers...)
	p.loggersLock.Unlock()
}
