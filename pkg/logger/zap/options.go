package zap

import "go-auth/pkg/logger"

type SamplingConfig struct {
	Initial    int
	Thereafter int
}

const samplingKey = "zap_sampling_config"

func WithSampling(initial, thereafter int) logger.Option {
	return func(c *logger.Config) {
		if c.DriverOptions == nil {
			c.DriverOptions = make(map[string]any)
		}
		c.DriverOptions[samplingKey] = SamplingConfig{
			Initial:    initial,
			Thereafter: thereafter,
		}
	}
}
