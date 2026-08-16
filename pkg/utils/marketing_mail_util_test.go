package utils

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/resend/resend-go/v3"
	"github.com/sirini/goapi/pkg/models"
)

func TestResendMarketingWorkflow(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/segments":
			_, _ = w.Write([]byte(`{"object":"segment","id":"segment_1","name":"NUBO members"}`))
		case r.Method == http.MethodPost && r.URL.Path == "/contacts/imports":
			if err := r.ParseMultipartForm(1 << 20); err != nil {
				t.Fatal(err)
			}
			file, _, err := r.FormFile("file")
			if err != nil {
				t.Fatal(err)
			}
			defer file.Close()
			content, _ := io.ReadAll(file)
			if !strings.Contains(string(content), "member@example.com,NUBO 회원") {
				t.Fatalf("CSV content = %q", content)
			}
			if !strings.Contains(r.FormValue("segments"), "segment_1") || r.FormValue("on_conflict") != "upsert" {
				t.Fatalf("import fields = %#v", r.MultipartForm.Value)
			}
			_, _ = w.Write([]byte(`{"object":"contact_import","id":"import_1"}`))
		case r.Method == http.MethodGet && r.URL.Path == "/contacts/imports/import_1":
			_, _ = w.Write([]byte(`{"object":"contact_import","id":"import_1","status":"completed","counts":{"total":1,"created":1,"updated":0,"skipped":0,"failed":0}}`))
		case r.Method == http.MethodPost && r.URL.Path == "/broadcasts":
			var request resend.CreateBroadcastRequest
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				t.Fatal(err)
			}
			if request.SegmentId != "segment_1" || request.Send || request.ReplyTo[0] != "owner@example.net" {
				t.Fatalf("broadcast request = %#v", request)
			}
			_, _ = w.Write([]byte(`{"id":"broadcast_1"}`))
		case r.Method == http.MethodPost && r.URL.Path == "/broadcasts/broadcast_1/send":
			_, _ = w.Write([]byte(`{"id":"broadcast_1"}`))
		case r.Method == http.MethodGet && r.URL.Path == "/broadcasts/broadcast_1":
			_, _ = w.Write([]byte(`{"id":"broadcast_1","status":"draft"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := resend.NewCustomClient(server.Client(), "re_test")
	baseURL, err := url.Parse(server.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	client.BaseURL = baseURL
	mailer := &ResendMailer{
		client: client, apiKey: "re_test", fromEmail: "notifications@example.com",
		fromName: "NUBO", replyTo: "owner@example.net",
	}

	segmentId, err := mailer.CreateSegment("NUBO members")
	if err != nil || segmentId != "segment_1" {
		t.Fatalf("CreateSegment() = %q, %v", segmentId, err)
	}
	importId, err := mailer.ImportContacts(segmentId, []models.MailRecipient{{Email: "member@example.com", Name: "NUBO 회원"}})
	if err != nil || importId != "import_1" {
		t.Fatalf("ImportContacts() = %q, %v", importId, err)
	}
	status, err := mailer.GetImportStatus(importId)
	if err != nil || status.Status != "completed" || status.Total != 1 {
		t.Fatalf("GetImportStatus() = %#v, %v", status, err)
	}
	broadcastId, err := mailer.CreateBroadcast(models.MarketingBroadcast{
		SegmentId: segmentId, Name: "Campaign", Subject: "소식", HTML: "<p>본문</p>", Text: "본문",
	})
	if err != nil || broadcastId != "broadcast_1" {
		t.Fatalf("CreateBroadcast() = %q, %v", broadcastId, err)
	}
	if err := mailer.SendBroadcast(broadcastId); err != nil {
		t.Fatal(err)
	}
	statusName, err := mailer.GetBroadcastStatus(broadcastId)
	if err != nil || statusName != "draft" {
		t.Fatalf("GetBroadcastStatus() = %q, %v", statusName, err)
	}
}
