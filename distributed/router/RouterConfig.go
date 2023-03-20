package router

type (
	Config[Msg any] struct {
		maxJmp    uint32
		processor Processor[Msg]
	}
)

func DefaultConfig[Msg any](processor Processor[Msg]) *Config[Msg] {
	return &Config[Msg]{
		maxJmp:    128,
		processor: processor,
	}
}

func (c *Config[Msg]) MaxJmp(maxJmp uint32) *Config[Msg] {
	c.maxJmp = maxJmp
	return c
}
