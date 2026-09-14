package app

import (
	"crypto/ed25519"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MrZloHex/monolink/marshal"
)

// cheap makes sealing quick for a test; the bounds still apply.
func cheap(t *testing.T) {
	old := defaultCost
	defaultCost = sealCost{Memory: 8 * 1024, Time: 1, Threads: 1}
	t.Cleanup(func() { defaultCost = old })
}

func TestAKeyIsSealedWithItsPassphrase(t *testing.T) {
	cheap(t)
	path := filepath.Join(t.TempDir(), "monoview.key")
	k, priv, cred, err := NewKeyFile("mzh", "laptop", "correct horse battery")
	if err != nil {
		t.Fatal(err)
	}
	if err := k.Save(path); err != nil {
		t.Fatal(err)
	}
	if fi, _ := os.Stat(path); fi.Mode().Perm() != 0o600 {
		t.Fatalf("key file mode %v", fi.Mode().Perm())
	}
	raw, _ := os.ReadFile(path)
	if strings.Contains(string(raw), b64(priv.Seed())) {
		t.Fatal("the seed is in the file in the clear")
	}

	back, err := LoadKeyFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := back.Open("wrong horse battery"); !errors.Is(err, ErrPassphrase) {
		t.Fatalf("a wrong passphrase: %v", err)
	}
	opened, err := back.Open("correct horse battery")
	if err != nil {
		t.Fatal(err)
	}
	sig := ed25519.Sign(opened, marshal.KeyMessage("mzh", "MONOVIEW", "n"))
	if err := marshal.VerifyKey(cred, "mzh", "MONOVIEW", "n", sig); err != nil {
		t.Fatalf("the opened key is not the one marshal was given: %v", err)
	}
	if err := cred.Check(); err != nil || cred.Label != "laptop" {
		t.Fatalf("credential %+v, %v", cred, err)
	}
}

func TestAKeyFileIsNeverReplaced(t *testing.T) {
	cheap(t)
	path := filepath.Join(t.TempDir(), "monoview.key")
	a, _, _, _ := NewKeyFile("mzh", "", "correct horse battery")
	b, _, _, _ := NewKeyFile("mzh", "", "correct horse battery")
	if err := a.Save(path); err != nil {
		t.Fatal(err)
	}
	if err := b.Save(path); err == nil {
		t.Fatal("a second key replaced the first")
	}
	if back, _ := LoadKeyFile(path); back.ID != a.ID {
		t.Fatal("the first key is gone")
	}
}

// The seal covers whose key it is: a file edited to say dasha does not open
// as dasha's.
func TestAKeyFileEditedDoesNotOpen(t *testing.T) {
	cheap(t)
	k, _, _, _ := NewKeyFile("mzh", "", "correct horse battery")
	edited := k
	edited.Person = "dasha"
	if _, err := edited.Open("correct horse battery"); err == nil {
		t.Fatal("an edited file opened")
	}
	costly := k
	costly.KDF.Memory = 1 << 31
	if _, err := costly.Open("correct horse battery"); err == nil || errors.Is(err, ErrPassphrase) {
		t.Fatalf("a file asking for two terabytes: %v", err)
	}
}
