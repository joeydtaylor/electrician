package websocketclient

import "github.com/joeydtaylor/electrician/pkg/internal/types"

// ConnectLogger attaches one or more loggers.
func (c *WebSocketClientAdapter[T]) ConnectLogger(loggers ...types.Logger) {
	if len(loggers) == 0 {
		return
	}
	c.loggersLock.Lock()
	c.loggers = append(c.loggers, loggers...)
	c.loggersLock.Unlock()
}

// ConnectSensor attaches one or more sensors.
func (c *WebSocketClientAdapter[T]) ConnectSensor(sensors ...types.Sensor[T]) {
	if len(sensors) == 0 {
		return
	}
	c.sensorsLock.Lock()
	c.sensors = append(c.sensors, sensors...)
	c.sensorsLock.Unlock()
}
