package websocket

import "github.com/joeydtaylor/electrician/pkg/internal/types"

func (s *serverAdapter[T]) notifyLifecycleStart() {
	sensors := s.snapshotSensors()
	for _, sensor := range sensors {
		if sensor == nil {
			continue
		}
		sensor.InvokeOnStart(s.componentMetadata)
	}
	if len(sensors) > 0 {
		s.NotifyLoggers(
			types.DebugLevel,
			"WebSocket server started",
			"component", s.componentMetadata,
			"event", "Start",
			"result", "SUCCESS",
		)
	}
}

func (s *serverAdapter[T]) notifyLifecycleStop() {
	sensors := s.snapshotSensors()
	for _, sensor := range sensors {
		if sensor == nil {
			continue
		}
		sensor.InvokeOnStop(s.componentMetadata)
	}
	if len(sensors) > 0 {
		s.NotifyLoggers(
			types.DebugLevel,
			"WebSocket server stopped",
			"component", s.componentMetadata,
			"event", "Stop",
			"result", "SUCCESS",
		)
	}
}
