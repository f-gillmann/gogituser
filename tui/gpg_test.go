package tui

import (
	"slices"
	"testing"
)

const gpgColons = `sec:u:255:22:8E7CA544EF5C944B:1700000000:::u:::scESC:::+:::ed25519:::0:
fpr:::::::::C7F0F53AFD7E1E1CFD4BDD7C8E7CA544EF5C944B:
grp:::::::::0000000000000000000000000000000000000000:
uid:u::::1700000000::HASH1::Me Personal <personal@example.test>::::::::::0:
uid:r::::1700000000::HASH2::Old Me <old@example.test>::::::::::0:
uid:u::::1700000000::HASH3::Me Work <WORK@example.test>::::::::::0:
ssb:u:255:22:AAAABBBBCCCCDDDD:1700000000::::::s:::+:::ed25519::
fpr:::::::::1111222233334444555566667777AAAABBBBCCCCDDDD:
grp:::::::::0000000000000000000000000000000000000000:
sec:u:255:22:1989EF9FE4BBD03D:1700000000:::u:::scESC:::+:::ed25519:::0:
fpr:::::::::D822BCECC351C4EC5D9C7DE51989EF9FE4BBD03D:
uid:u::::1700000000::HASH4::Other <other@example.test>::::::::::0:
`

func TestParseGPGKeys(t *testing.T) {
	keys := parseGPGKeys([]byte(gpgColons))
	if len(keys) != 2 {
		t.Fatalf("got %d keys", len(keys))
	}

	first := keys[0]
	if first.fingerprint != "C7F0F53AFD7E1E1CFD4BDD7C8E7CA544EF5C944B" {
		t.Errorf("fingerprint = %q", first.fingerprint)
	}
	if !slices.Equal(first.subkeys, []string{"1111222233334444555566667777AAAABBBBCCCCDDDD"}) {
		t.Errorf("subkeys = %q", first.subkeys)
	}
	if len(first.uids) != 2 {
		t.Errorf("revoked uid not skipped: %q", first.uids)
	}
	if first.keyID() != "8E7CA544EF5C944B" {
		t.Errorf("keyID = %q", first.keyID())
	}
}

func TestGPGKeyEmails(t *testing.T) {
	key := parseGPGKeys([]byte(gpgColons))[0]

	if !key.hasEmail(" work@EXAMPLE.test ") {
		t.Error("email match should ignore case and spaces")
	}
	if key.hasEmail("old@example.test") {
		t.Error("revoked uid must not count")
	}
	if got := key.uidFor("work@example.test"); got != "Me Work <WORK@example.test>" {
		t.Errorf("uidFor = %q", got)
	}
	if got := key.uidFor("nobody@example.test"); got != "Me Personal <personal@example.test>" {
		t.Errorf("uidFor fallback = %q", got)
	}
}

func TestFindGPGKey(t *testing.T) {
	keys := parseGPGKeys([]byte(gpgColons))
	tests := map[string]string{
		"C7F0F53AFD7E1E1CFD4BDD7C8E7CA544EF5C944B": "C7F0F53AFD7E1E1CFD4BDD7C8E7CA544EF5C944B",
		"8E7CA544EF5C944B":                         "C7F0F53AFD7E1E1CFD4BDD7C8E7CA544EF5C944B",
		"0x8e7ca544ef5c944b":                       "C7F0F53AFD7E1E1CFD4BDD7C8E7CA544EF5C944B",
		"EF5C944B":                                 "C7F0F53AFD7E1E1CFD4BDD7C8E7CA544EF5C944B",
		"AAAABBBBCCCCDDDD!":                        "C7F0F53AFD7E1E1CFD4BDD7C8E7CA544EF5C944B",
		"1989EF9FE4BBD03D":                         "D822BCECC351C4EC5D9C7DE51989EF9FE4BBD03D",
		"944B":                                     "",
		"DEADBEEFDEADBEEF":                         "",
	}
	for id, want := range tests {
		key, found := findGPGKey(keys, id)
		if found != (want != "") || key.fingerprint != want {
			t.Errorf("findGPGKey(%q) = %q, %v, want %q", id, key.fingerprint, found, want)
		}
	}
}
