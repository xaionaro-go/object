package object

type Option interface {
	apply(*config)
}

type config struct {
	VisitorFunc       VisitorFunc
	ProcessUnexported bool
}

type Options []Option

func (s Options) apply(cfg *config) {
	for _, opt := range s {
		opt.apply(cfg)
	}
}

func (s Options) config() config {
	cfg := config{}
	s.apply(&cfg)
	return cfg
}

type OptionWithUnexported bool

func (opt OptionWithUnexported) apply(cfg *config) {
	cfg.ProcessUnexported = bool(opt)
}

type OptionWithVisitorFunc VisitorFunc

func (opt OptionWithVisitorFunc) apply(cfg *config) {
	cfg.VisitorFunc = VisitorFunc(opt)
}
