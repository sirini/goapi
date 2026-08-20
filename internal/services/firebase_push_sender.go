package services

import (
	"context"
	"fmt"
	"strings"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

type PushMessage struct {
	Title string
	Body  string
	Data  map[string]string
}

type PushSender interface {
	Send(ctx context.Context, installationIDs []string, message PushMessage) ([]string, error)
}

type disabledPushSender struct{}

func (disabledPushSender) Send(context.Context, []string, PushMessage) ([]string, error) {
	return nil, nil
}

type firebasePushSender struct {
	client *messaging.Client
}

// ADC 또는 GOOGLE_APPLICATION_CREDENTIALS를 사용하므로 자격 증명을 코드에 저장하지 않는다.
func newFirebasePushSender(ctx context.Context, projectID string, credentialsFile string) (PushSender, error) {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return disabledPushSender{}, nil
	}
	var options []option.ClientOption
	if credentialsFile = strings.TrimSpace(credentialsFile); credentialsFile != "" {
		options = append(options, option.WithCredentialsFile(credentialsFile))
	}
	app, err := firebase.NewApp(ctx, &firebase.Config{ProjectID: projectID}, options...)
	if err != nil {
		return disabledPushSender{}, fmt.Errorf("create Firebase app: %w", err)
	}
	client, err := app.Messaging(ctx)
	if err != nil {
		return disabledPushSender{}, fmt.Errorf("create Firebase messaging client: %w", err)
	}
	return &firebasePushSender{client: client}, nil
}

func (s *firebasePushSender) Send(
	ctx context.Context,
	installationIDs []string,
	message PushMessage,
) ([]string, error) {
	if len(installationIDs) == 0 {
		return nil, nil
	}
	invalid := make([]string, 0)
	for start := 0; start < len(installationIDs); start += 500 {
		end := min(start+500, len(installationIDs))
		batch := installationIDs[start:end]
		response, err := s.client.SendEachForMulticast(ctx, &messaging.MulticastMessage{
			Fids: batch,
			Data: message.Data,
			Notification: &messaging.Notification{
				Title: message.Title,
				Body:  message.Body,
			},
			Android: &messaging.AndroidConfig{Priority: "high"},
		})
		if err != nil {
			return invalid, err
		}
		for index, result := range response.Responses {
			if result.Error != nil && messaging.IsUnregistered(result.Error) {
				invalid = append(invalid, batch[index])
			}
		}
	}
	return invalid, nil
}
