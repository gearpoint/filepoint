package aws_repository

import (
	"github.com/aws/aws-sdk-go-v2/service/rekognition"
	"github.com/aws/aws-sdk-go-v2/service/rekognition/types"
)

// GetImageLabels returns the image labels. Suports only JPEG and PNG, up to 15MB.
func (r *AWSRepository) GetImageLabels(prefix string) ([]string, error) {
	var minConfidence float32 = 97
	var maxLabels = maxRekognitionLabels
	var labels []string

	result, err := r.rekoClient.DetectLabels(r.ctx, &rekognition.DetectLabelsInput{
		Image: &types.Image{
			S3Object: &types.S3Object{
				Bucket: &r.config.Bucket,
				Name:   &prefix,
			},
		},
		MaxLabels:     &maxLabels,
		MinConfidence: &minConfidence,
	})
	if err != nil {
		return nil, err
	}

	for _, label := range result.Labels {
		labels = append(labels, *label.Name)
	}

	moderation, _ := r.rekoClient.DetectModerationLabels(r.ctx, &rekognition.DetectModerationLabelsInput{
		Image: &types.Image{
			S3Object: &types.S3Object{
				Bucket: &r.config.Bucket,
				Name:   &prefix,
			},
		},
		MinConfidence: &minConfidence,
	})

	for _, label := range moderation.ModerationLabels {
		labels = append(labels, *label.Name)
	}

	text, _ := r.rekoClient.DetectText(r.ctx, &rekognition.DetectTextInput{
		Image: &types.Image{
			S3Object: &types.S3Object{
				Bucket: &r.config.Bucket,
				Name:   &prefix,
			},
		},
	})

	for _, label := range text.TextDetections {
		if *label.Confidence >= minConfidence {
			labels = append(labels, *label.DetectedText)
		}
	}

	return labels, err
}

// StartVideoLabelsDetection starts the video label and moderation detection.
func (r *AWSRepository) StartVideoLabelsDetection(prefix string) ([]string, error) {
	// var minConfidence float32 = 97

	// r.rekoClient.StartLabelDetection(r.ctx, &rekognition.StartLabelDetectionInput{
	// 	Video: &types.Video{
	// 		S3Object: &types.S3Object{
	// 			Bucket: &r.config.Bucket,
	// 			Name:   &prefix,
	// 		},
	// 	},
	// 	MinConfidence: &minConfidence,
	// 	NotificationChannel: &types.NotificationChannel{
	// 		RoleArn:     &r.config.RekognitionRole,
	// 		SNSTopicArn: &r.config.VideoLabelingTopic,
	// 	},
	// })
	// r.rekoClient.StartContentModeration(r.ctx, &rekognition.StartContentModerationInput{
	// 	Video: &types.Video{
	// 		S3Object: &types.S3Object{
	// 			Bucket: &r.config.Bucket,
	// 			Name:   &prefix,
	// 		},
	// 	},
	// 	MinConfidence: &minConfidence,
	// 	NotificationChannel: &types.NotificationChannel{
	// 		RoleArn:     &r.config.RekognitionRole,
	// 		SNSTopicArn: &r.config.VideoLabelingTopic,
	// 	},
	// })

	return nil, nil
}
