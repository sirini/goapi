package services

import "testing"

func TestBuildFirebaseMulticastMessageSupportsAndroidAndIOS(t *testing.T) {
	installationIDs := []string{"installation-id-1", "installation-id-2"}
	message := PushMessage{
		Title: "SENSTA",
		Body:  "새 알림이 있습니다.",
		Data:  map[string]string{"type": "4", "fromUserUid": "27"},
	}

	got := buildFirebaseMulticastMessage(installationIDs, message)
	if len(got.Fids) != 2 || got.Fids[0] != installationIDs[0] || got.Fids[1] != installationIDs[1] {
		t.Fatalf("installation ids = %#v", got.Fids)
	}
	if got.Notification == nil || got.Notification.Title != message.Title || got.Notification.Body != message.Body {
		t.Fatalf("notification = %#v", got.Notification)
	}
	if got.Android == nil || got.Android.Priority != "high" {
		t.Fatalf("android config = %#v", got.Android)
	}
	if got.APNS == nil || got.APNS.Headers["apns-push-type"] != "alert" || got.APNS.Headers["apns-priority"] != "10" {
		t.Fatalf("APNs config = %#v", got.APNS)
	}
	if got.APNS.Payload == nil || got.APNS.Payload.Aps == nil || got.APNS.Payload.Aps.Sound != "default" {
		t.Fatalf("APNs payload = %#v", got.APNS.Payload)
	}
	if got.Data["fromUserUid"] != "27" {
		t.Fatalf("data = %#v", got.Data)
	}
}
