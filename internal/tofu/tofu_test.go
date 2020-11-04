package tofu

import "testing"

func TestFingerprint(t *testing.T) {
	fp := Fingerprint([]byte("abc"))
	if fp.Encode() != "YWJj" {
		t.Errorf("Unexpected fingerprint: '%s'", fp.Encode())
	}
}

func TestCompareFingerprint(t *testing.T) {
	expected, err := FingerprintOfString("YWJj")
	if err != nil {
		t.Fatal(err)
	}
	if !Fingerprint([]byte("abc")).Matches(expected) {
		t.Errorf("expected same fingerprint")
	}
	if Fingerprint([]byte("hej")).Matches(expected) {
		t.Errorf("expected different fingerprint")
	}
}
