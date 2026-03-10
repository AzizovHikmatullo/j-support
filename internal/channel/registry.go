package channel

type Registry struct {
	channels map[string]Channel
}

func NewRegistry() *Registry {
	return &Registry{
		channels: make(map[string]Channel),
	}
}

func (r *Registry) Register(ch Channel) {
	r.channels[ch.Name()] = ch
}

func (r *Registry) Get(name string) (Channel, error) {
	ch, ok := r.channels[name]
	if !ok {
		return nil, ErrChannelNotFound
	}
	return ch, nil
}
