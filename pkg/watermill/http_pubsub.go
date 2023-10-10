package watermill

import (
	watermill_http "github.com/ThreeDotsLabs/watermill-http/pkg/http"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/gearpoint/filepoint/pkg/logger"
)

// NewHttpPublisher creates a Publisher.
func NewHttpPublisher() (message.Publisher, error) {
	publisherCfg := watermill_http.PublisherConfig{
		MarshalMessageFunc: watermill_http.DefaultMarshalMessageFunc,
	}

	publisher, err := watermill_http.NewPublisher(
		publisherCfg,
		NewZapLoggerAdapter(logger.Logger),
	)

	return publisher, err
}
