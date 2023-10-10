package watermill

import (
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/gearpoint/filepoint/pkg/logger"
)

// NewGoChannel creates a GoChannel.
func NewGoChannel() *gochannel.GoChannel {
	publisherCfg := gochannel.Config{
		OutputChannelBuffer:            0,
		Persistent:                     false,
		BlockPublishUntilSubscriberAck: false,
	}

	publisher := gochannel.NewGoChannel(
		publisherCfg,
		NewZapLoggerAdapter(logger.Logger),
	)

	return publisher
}
