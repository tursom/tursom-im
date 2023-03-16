package distributed

type (
	ClusterConfig[Msg any] struct {
		maxJmp    uint32
		processor MessageProcessor[Msg]
	}
)

func NewConfig[Msg any](processor MessageProcessor[Msg]) *ClusterConfig[Msg] {
	return &ClusterConfig[Msg]{
		maxJmp:    128,
		processor: processor,
	}
}

func (c *ClusterConfig[Msg]) MaxJmp(maxJmp uint32) *ClusterConfig[Msg] {
	c.maxJmp = maxJmp
	return c
}
