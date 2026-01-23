package plug

import "github.com/joeydtaylor/electrician/pkg/internal/types"

// GetAdapterFuncs returns configured adapter funcs.
func (p *Plug[T]) GetAdapterFuncs() []types.AdapterFunc[T] {
	p.configLock.Lock()
	funcs := append([]types.AdapterFunc[T](nil), p.adapterFns...)
	p.configLock.Unlock()
	return funcs
}

// GetConnectors returns configured adapters.
func (p *Plug[T]) GetConnectors() []types.Adapter[T] {
	p.configLock.Lock()
	adapters := append([]types.Adapter[T](nil), p.adapters...)
	p.configLock.Unlock()
	return adapters
}

// GetComponentMetadata returns the plug metadata.
func (p *Plug[T]) GetComponentMetadata() types.ComponentMetadata {
	return p.componentMetadata
}
