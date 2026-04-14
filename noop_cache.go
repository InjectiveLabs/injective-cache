package cache

import "context"

var _ Cache = (*noopCache)(nil)

type noopCache struct{}

func NewNoopCache() Cache {
	return noopCache{}
}

func (noopCache) Set(_ context.Context, _ string, _ []byte) error           { return nil }
func (noopCache) Get(_ context.Context, _ string) ([]byte, error)           { return nil, ErrCacheMiss }
func (noopCache) Del(_ context.Context, _ string) error                     { return nil }
func (noopCache) BatchGet(_ context.Context, _ ...string) ([][]byte, error) { return nil, nil }
func (noopCache) BatchSet(_ context.Context, _ ...interface{}) error        { return nil }
func (noopCache) IsRunning(_ context.Context) bool                          { return false }
func (noopCache) Close() error                                               { return nil }
