package tofu

import "testing"

func TestFingerprint(t *testing.T) {
	fp := getFingerprint([]byte("abc"))
	if fp != "BA:78:16:BF:8F:01:CF:EA:41:41:40:DE:5D:AE:22:23:B0:03:61:A3:96:17:7A:9C:B4:10:FF:61:F2:00:15:AD" {
		t.Errorf("Unexpected fingerprint: '%s'", fp)
	}
}

func TestCompareFingerprint(t *testing.T) {
	expected := fingerprintBytes("BA:78:16:BF:8F:01:CF:EA:41:41:40:DE:5D:AE:22:23:B0:03:61:A3:96:17:7A:9C:B4:10:FF:61:F2:00:15:AD")
	if !compareFingerprint([]byte("abc"), expected) {
		t.Errorf("expected same fingerprint")
	}
	if compareFingerprint([]byte("hej"), expected) {
		t.Errorf("expected different fingerprint")
	}
}
