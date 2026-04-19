package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/yoosuf/hopper/internal/platform/logger"
)

// ProviderSender is a sender implementation with pluggable provider names.
type ProviderSender struct {
	integrationsEnabled bool
	smsProvider         string
	pushProvider        string
	twilioAccountSID    string
	twilioAuthToken     string
	twilioFromNumber    string
	firebaseServerKey   string
	firebaseEndpoint    string
	httpClient          *http.Client
	log                 logger.Logger
}

// NewProviderSender creates a notification sender that can route to external providers.
func NewProviderSender(integrationsEnabled bool, smsProvider, pushProvider string, twilioAccountSID, twilioAuthToken, twilioFromNumber, firebaseServerKey, firebaseEndpoint string, log logger.Logger) *ProviderSender {
	if firebaseEndpoint == "" {
		firebaseEndpoint = "https://fcm.googleapis.com/fcm/send"
	}
	return &ProviderSender{
		integrationsEnabled: integrationsEnabled,
		smsProvider:         smsProvider,
		pushProvider:        pushProvider,
		twilioAccountSID:    twilioAccountSID,
		twilioAuthToken:     twilioAuthToken,
		twilioFromNumber:    twilioFromNumber,
		firebaseServerKey:   firebaseServerKey,
		firebaseEndpoint:    firebaseEndpoint,
		httpClient:          &http.Client{Timeout: 8 * time.Second},
		log:                 log,
	}
}

// SendPush sends push notifications or logs when integrations are disabled.
func (s *ProviderSender) SendPush(ctx context.Context, userID uuid.UUID, title, body string, data map[string]interface{}) error {
	if !s.integrationsEnabled {
		s.log.Info("Push notification skipped; integrations disabled", logger.F("user_id", userID), logger.F("title", title), logger.F("body", body), logger.F("data", data))
		return nil
	}
	if strings.EqualFold(s.pushProvider, "firebase") {
		return s.sendFirebasePush(ctx, userID, title, body, data)
	}

	s.log.Info("Push notification dispatched", logger.F("provider", s.pushProvider), logger.F("user_id", userID), logger.F("title", title))
	return nil
}

// SendEmail sends email notifications. SMTP integration can be added by extending this sender.
func (s *ProviderSender) SendEmail(ctx context.Context, to string, subject, body string) error {
	if !s.integrationsEnabled {
		s.log.Info("Email notification skipped; integrations disabled", logger.F("to", to), logger.F("subject", subject))
		return nil
	}

	s.log.Info("Email notification dispatched", logger.F("to", to), logger.F("subject", subject), logger.F("body", body))
	return nil
}

// SendSMS sends SMS notifications or logs when integrations are disabled.
func (s *ProviderSender) SendSMS(ctx context.Context, to, message string) error {
	if !s.integrationsEnabled {
		s.log.Info("SMS notification skipped; integrations disabled", logger.F("to", to), logger.F("message", message))
		return nil
	}
	if strings.EqualFold(s.smsProvider, "twilio") {
		return s.sendTwilioSMS(ctx, to, message)
	}

	s.log.Info("SMS notification dispatched", logger.F("provider", s.smsProvider), logger.F("to", to), logger.F("message", message))
	return nil
}

func (s *ProviderSender) sendTwilioSMS(ctx context.Context, to, message string) error {
	if s.twilioAccountSID == "" || s.twilioAuthToken == "" || s.twilioFromNumber == "" {
		return fmt.Errorf("twilio credentials are not configured")
	}

	values := url.Values{}
	values.Set("To", to)
	values.Set("From", s.twilioFromNumber)
	values.Set("Body", message)

	u := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", s.twilioAccountSID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, strings.NewReader(values.Encode()))
	if err != nil {
		return err
	}
	req.SetBasicAuth(s.twilioAccountSID, s.twilioAuthToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("twilio send failed with status %d", resp.StatusCode)
	}

	s.log.Info("SMS notification dispatched", logger.F("provider", "twilio"), logger.F("to", to))
	return nil
}

func (s *ProviderSender) sendFirebasePush(ctx context.Context, userID uuid.UUID, title, body string, data map[string]interface{}) error {
	if s.firebaseServerKey == "" {
		return fmt.Errorf("firebase server key is not configured")
	}

	payload := map[string]interface{}{
		"to": "/topics/user-" + userID.String(),
		"notification": map[string]string{
			"title": title,
			"body":  body,
		},
		"data": data,
	}
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.firebaseEndpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "key="+s.firebaseServerKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("firebase push failed with status %d", resp.StatusCode)
	}

	s.log.Info("Push notification dispatched", logger.F("provider", "firebase"), logger.F("user_id", userID), logger.F("title", title))
	return nil
}
