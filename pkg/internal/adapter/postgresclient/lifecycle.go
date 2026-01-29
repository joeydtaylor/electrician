package postgresclient

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

func (p *PostgresClient[T]) Stop() {
	if atomic.SwapInt32(&p.isServing, 0) == 1 {
		p.cancel()
	}
	if p.db != nil {
		_ = p.db.Close()
	}
}

func (p *PostgresClient[T]) ConnectLogger(l ...types.Logger) {
	p.loggersLock.Lock()
	defer p.loggersLock.Unlock()
	p.loggers = append(p.loggers, l...)
}

func (p *PostgresClient[T]) ConnectSensor(s ...types.Sensor[T]) {
	p.sensorLock.Lock()
	defer p.sensorLock.Unlock()
	p.sensors = append(p.sensors, s...)
}

func (p *PostgresClient[T]) NotifyLoggers(level types.LogLevel, msg string, keysAndValues ...interface{}) {
	loggers := p.snapshotLoggers()
	if len(loggers) == 0 {
		return
	}
	for _, logger := range loggers {
		if logger == nil || logger.GetLevel() > level {
			continue
		}
		switch level {
		case types.DebugLevel:
			logger.Debug(msg, keysAndValues...)
		case types.InfoLevel:
			logger.Info(msg, keysAndValues...)
		case types.WarnLevel:
			logger.Warn(msg, keysAndValues...)
		case types.ErrorLevel:
			logger.Error(msg, keysAndValues...)
		case types.DPanicLevel:
			logger.DPanic(msg, keysAndValues...)
		case types.PanicLevel:
			logger.Panic(msg, keysAndValues...)
		case types.FatalLevel:
			logger.Fatal(msg, keysAndValues...)
		}
	}
}

func (p *PostgresClient[T]) GetComponentMetadata() types.ComponentMetadata {
	return p.componentMetadata
}

func (p *PostgresClient[T]) SetComponentMetadata(name string, id string) {
	p.componentMetadata.Name = name
	if id != "" {
		p.componentMetadata.ID = id
	}
}

func (p *PostgresClient[T]) Name() string {
	return "postgres_client"
}

func (p *PostgresClient[T]) snapshotLoggers() []types.Logger {
	p.loggersLock.Lock()
	defer p.loggersLock.Unlock()
	if len(p.loggers) == 0 {
		return nil
	}
	out := make([]types.Logger, len(p.loggers))
	copy(out, p.loggers)
	return out
}

func (p *PostgresClient[T]) ConnectInput(wires ...types.Wire[T]) {
	p.inputWires = append(p.inputWires, wires...)
}

func (p *PostgresClient[T]) mergeInputs() <-chan T {
	if len(p.inputWires) == 0 {
		return nil
	}
	if p.mergedIn != nil {
		return p.mergedIn
	}
	out := make(chan T)
	p.mergedIn = out
	for _, w := range p.inputWires {
		wire := w
		go func() {
			for item := range wire.GetOutputChannel() {
				out <- item
			}
		}()
	}
	return out
}

func (p *PostgresClient[T]) ensureTable(ctx context.Context) error {
	if !p.writerCfg.AutoCreateTable {
		return nil
	}
	ddl := p.writerCfg.CreateTableDDL
	if ddl == "" {
		ddl = defaultTableDDL(p.writerCfg)
	}
	if ddl == "" {
		return fmt.Errorf("postgres: create table ddl is empty")
	}
	db, err := p.ensureDB(ctx)
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, ddl)
	return err
}
