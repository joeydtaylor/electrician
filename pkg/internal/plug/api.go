package plug

import (
	"fmt"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

func (p *Plug[T]) GetAdapterFuncs() []types.AdapterFunc[T] {
	return p.plugFunc
}

func (p *Plug[T]) GetConnectors() []types.Adapter[T] {
	return p.clients
}

func (p *Plug[T]) ConnectAdapter(pc ...types.Adapter[T]) {
	p.configLock.Lock()
	defer p.configLock.Unlock()
	p.clients = append(p.clients, pc...)
}

func (p *Plug[T]) AddAdapterFunc(pf ...types.AdapterFunc[T]) {
	p.configLock.Lock()
	defer p.configLock.Unlock()
	p.plugFunc = append(p.plugFunc, pf...)
}

// GetComponentMetadata returns the metadata.
func (p *Plug[T]) GetComponentMetadata() types.ComponentMetadata {
	return p.componentMetadata
}

// SetComponentMetadata sets the component metadata.
func (p *Plug[T]) SetComponentMetadata(name string, id string) {
	p.componentMetadata = types.ComponentMetadata{Name: name, ID: id}
}

func (p *Plug[T]) ConnectSensor(sensor ...types.Sensor[T]) {
	p.sensors = append(p.sensors, sensor...)
	for _, m := range sensor {
		p.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: ConnectSensor, target: %v => Connected sensor", p.componentMetadata, m.GetComponentMetadata())
	}
}

func (p *Plug[T]) ConnectLogger(l ...types.Logger) {
	p.loggersLock.Lock()
	defer p.loggersLock.Unlock()
	p.loggers = append(p.loggers, l...)
}

func (p *Plug[T]) NotifyLoggers(level types.LogLevel, format string, args ...interface{}) {
	if p.loggers != nil {
		msg := fmt.Sprintf(format, args...)
		for _, logger := range p.loggers {
			if logger == nil {
				continue
			}
			// Ensure we only acquire the lock once per logger to avoid deadlock or excessive locking overhead
			p.loggersLock.Lock()
			if logger.GetLevel() <= level {
				switch level {
				case types.DebugLevel:
					logger.Debug(msg)
				case types.InfoLevel:
					logger.Info(msg)
				case types.WarnLevel:
					logger.Warn(msg)
				case types.ErrorLevel:
					logger.Error(msg)
				case types.DPanicLevel:
					logger.DPanic(msg)
				case types.PanicLevel:
					logger.Panic(msg)
				case types.FatalLevel:
					logger.Fatal(msg)
				}
			}
			p.loggersLock.Unlock()
		}
	}
}
