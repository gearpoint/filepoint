package watermill

import (
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/gearpoint/filepoint/pkg/logger"
)

// NewHttpPublisher creates a Publisher.
func NewRouter() (*message.Router, error) {
	routerCfg := message.RouterConfig{}

	return message.NewRouter(routerCfg, NewZapLoggerAdapter(logger.Logger))
}
