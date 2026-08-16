package utils

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/resend/resend-go/v3"
	"github.com/sirini/goapi/pkg/models"
)

const marketingRequestTimeout = 30 * time.Second

// MarketingMailer owns the Resend resources used for administrator broadcasts.
type MarketingMailer interface {
	Configured() bool
	CreateSegment(name string) (string, error)
	ImportContacts(segmentId string, contacts []models.MailRecipient) (string, error)
	GetImportStatus(importId string) (models.MailImportStatus, error)
	CreateBroadcast(message models.MarketingBroadcast) (string, error)
	GetBroadcastStatus(broadcastId string) (string, error)
	SendBroadcast(broadcastId string) error
}

func (m *ResendMailer) CreateSegment(name string) (string, error) {
	if !m.Configured() {
		return "", fmt.Errorf("resend mail delivery is not configured")
	}
	ctx, cancel := context.WithTimeout(context.Background(), marketingRequestTimeout)
	defer cancel()
	response, err := m.client.Segments.CreateWithContext(ctx, &resend.CreateSegmentRequest{Name: name})
	if err != nil {
		return "", fmt.Errorf("failed to create Resend segment: %w", err)
	}
	if response.Id == "" {
		return "", fmt.Errorf("Resend returned no segment id")
	}
	return response.Id, nil
}

func (m *ResendMailer) ImportContacts(segmentId string, contacts []models.MailRecipient) (string, error) {
	if !m.Configured() || segmentId == "" {
		return "", fmt.Errorf("Resend marketing delivery is not configured")
	}
	if len(contacts) == 0 {
		return "", fmt.Errorf("there are no eligible mail recipients")
	}

	var body bytes.Buffer
	writer := csv.NewWriter(&body)
	if err := writer.Write([]string{"email", "first_name"}); err != nil {
		return "", err
	}
	for _, contact := range contacts {
		if err := writer.Write([]string{contact.Email, Unescape(contact.Name)}); err != nil {
			return "", err
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(context.Background(), marketingRequestTimeout)
	defer cancel()
	response, err := m.client.Contacts.Imports.CreateWithContext(ctx, &resend.CreateContactImportRequest{
		File:       body.Bytes(),
		Filename:   "nubo-members.csv",
		ColumnMap:  map[string]any{"email": "email", "first_name": "first_name"},
		OnConflict: "upsert",
		Segments:   []resend.ContactImportSegment{{Id: segmentId}},
	})
	if err != nil {
		return "", fmt.Errorf("failed to import Resend contacts: %w", err)
	}
	if response.Id == "" {
		return "", fmt.Errorf("Resend returned no contact import id")
	}
	return response.Id, nil
}

func (m *ResendMailer) GetImportStatus(importId string) (models.MailImportStatus, error) {
	result := models.MailImportStatus{}
	if !m.Configured() || importId == "" {
		return result, fmt.Errorf("Resend marketing delivery is not configured")
	}
	ctx, cancel := context.WithTimeout(context.Background(), marketingRequestTimeout)
	defer cancel()
	response, err := m.client.Contacts.Imports.GetWithContext(ctx, importId)
	if err != nil {
		return result, fmt.Errorf("failed to load Resend contact import: %w", err)
	}
	result.Status = string(response.Status)
	if response.Counts != nil {
		result.Total = uint(response.Counts.Total)
		result.Failed = uint(response.Counts.Failed)
		result.Skipped = uint(response.Counts.Skipped)
	}
	return result, nil
}

func (m *ResendMailer) CreateBroadcast(message models.MarketingBroadcast) (string, error) {
	if !m.Configured() {
		return "", fmt.Errorf("Resend marketing delivery is not configured")
	}
	if message.SegmentId == "" || strings.TrimSpace(message.Subject) == "" || strings.TrimSpace(message.HTML) == "" {
		return "", fmt.Errorf("segment, subject and HTML body are required")
	}
	replyTo := make([]string, 0, 1)
	if m.replyTo != "" {
		replyTo = append(replyTo, m.replyTo)
	}
	ctx, cancel := context.WithTimeout(context.Background(), marketingRequestTimeout)
	defer cancel()
	response, err := m.client.Broadcasts.CreateWithContext(ctx, &resend.CreateBroadcastRequest{
		SegmentId: message.SegmentId,
		From:      (&mail.Address{Name: m.fromName, Address: m.fromEmail}).String(),
		ReplyTo:   replyTo,
		Name:      message.Name,
		Subject:   message.Subject,
		Html:      message.HTML,
		Text:      message.Text,
	})
	if err != nil {
		return "", fmt.Errorf("Resend rejected the broadcast: %w", err)
	}
	if response.Id == "" {
		return "", fmt.Errorf("Resend returned no broadcast id")
	}
	return response.Id, nil
}

func (m *ResendMailer) SendBroadcast(broadcastId string) error {
	if !m.Configured() || broadcastId == "" {
		return fmt.Errorf("Resend marketing delivery is not configured")
	}
	ctx, cancel := context.WithTimeout(context.Background(), marketingRequestTimeout)
	defer cancel()
	_, err := m.client.Broadcasts.SendWithContext(ctx, &resend.SendBroadcastRequest{BroadcastId: broadcastId})
	if err != nil {
		return fmt.Errorf("Resend rejected the broadcast send request: %w", err)
	}
	return nil
}

func (m *ResendMailer) GetBroadcastStatus(broadcastId string) (string, error) {
	if !m.Configured() || broadcastId == "" {
		return "", fmt.Errorf("Resend marketing delivery is not configured")
	}
	ctx, cancel := context.WithTimeout(context.Background(), marketingRequestTimeout)
	defer cancel()
	response, err := m.client.Broadcasts.GetWithContext(ctx, broadcastId)
	if err != nil {
		return "", fmt.Errorf("failed to load Resend broadcast: %w", err)
	}
	return response.Status, nil
}
