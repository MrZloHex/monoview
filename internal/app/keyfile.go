package app

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"golang.org/x/crypto/argon2"

	"github.com/MrZloHex/monolink/marshal"
)

// This panel's key. SECURITY.txt §6: a panel that is not a browser signs its
// person in with an Ed25519 key of its own, and marshal keeps the public
// half. The private half lives in a file, sealed with the person's
// passphrase — argon2id, then AES-256-GCM — so the file alone, copied off a
// stolen laptop, signs nobody in. A hardware key is to take its place.

// KeyFile is what the file holds. Nothing in it is secret but what is sealed.
type KeyFile struct {
	Version int      `json:"version"`
	Person  string   `json:"person"`
	ID      string   `json:"id"` // the credential id marshal knows the key by
	KDF     sealCost `json:"kdf"`
	Salt    string   `json:"salt"`
	Nonce   string   `json:"nonce"`
	Sealed  string   `json:"sealed"` // the key's seed
}

type sealCost struct {
	Memory  uint32 `json:"memory"` // KiB
	Time    uint32 `json:"time"`
	Threads uint8  `json:"threads"`
}

// defaultCost is argon2id at a quarter of a gigabyte: about a second here,
// and that much memory per guess for whoever copied the file.
var defaultCost = sealCost{Memory: 256 * 1024, Time: 3, Threads: 4}

// ErrPassphrase is a passphrase that does not open the key.
var ErrPassphrase = errors.New("wrong passphrase")

// NewKeyFile makes a key for person, sealed with passphrase, and returns it
// with the credential marshal is to keep.
func NewKeyFile(person, label, passphrase string) (KeyFile, ed25519.PrivateKey, marshal.Credential, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return KeyFile{}, nil, marshal.Credential{}, err
	}
	cred, err := marshal.Ed25519Credential(pub, label)
	if err != nil {
		return KeyFile{}, nil, marshal.Credential{}, err
	}
	k := KeyFile{Version: 1, Person: person, ID: cred.ID, KDF: defaultCost,
		Salt: b64(random(16)), Nonce: b64(random(12))}
	aead, err := k.aead(passphrase)
	if err != nil {
		return KeyFile{}, nil, marshal.Credential{}, err
	}
	k.Sealed = b64(aead.Seal(nil, unb64(k.Nonce), priv.Seed(), k.bound()))
	return k, priv, cred, nil
}

// Open unseals the key. Clear it once it has signed.
func (k KeyFile) Open(passphrase string) (ed25519.PrivateKey, error) {
	aead, err := k.aead(passphrase)
	if err != nil {
		return nil, err
	}
	nonce, sealed := unb64(k.Nonce), unb64(k.Sealed)
	if len(nonce) != aead.NonceSize() || sealed == nil {
		return nil, errors.New("the key file is damaged")
	}
	seed, err := aead.Open(nil, nonce, sealed, k.bound())
	if err != nil {
		return nil, ErrPassphrase
	}
	defer clear(seed)
	if len(seed) != ed25519.SeedSize {
		return nil, errors.New("the key file is damaged")
	}
	priv := ed25519.NewKeyFromSeed(seed)
	if c, _ := marshal.Ed25519Credential(priv.Public().(ed25519.PublicKey), ""); c.ID != k.ID {
		clear(priv)
		return nil, errors.New("the key file is damaged")
	}
	return priv, nil
}

// bound is what the seal covers besides the seed: whose key it is, and which.
func (k KeyFile) bound() []byte {
	return []byte("monolith-panel-key\x00" + k.Person + "\x00" + k.ID)
}

// aead derives the sealing key. A file's cost must be within bounds: one
// written to make this panel spend a terabyte, or no effort at all, is not
// opened.
func (k KeyFile) aead(passphrase string) (cipher.AEAD, error) {
	c := k.KDF
	salt := unb64(k.Salt)
	if k.Version != 1 || c.Memory < 8*1024 || c.Memory > 1024*1024 || c.Time < 1 || c.Time > 10 ||
		c.Threads < 1 || c.Threads > 16 || len(salt) < 16 {
		return nil, errors.New("the key file is not one this monoview reads")
	}
	key := argon2.IDKey([]byte(passphrase), salt, c.Time, c.Memory, c.Threads, 32)
	defer clear(key)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

// LoadKeyFile reads the key file at path.
func LoadKeyFile(path string) (KeyFile, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return KeyFile{}, fmt.Errorf("no key at %s: [e] enrol, or [i] take up an invitation", path)
		}
		return KeyFile{}, err
	}
	var k KeyFile
	if err := json.Unmarshal(b, &k); err != nil {
		return KeyFile{}, fmt.Errorf("%s is not a key file", path)
	}
	return k, nil
}

// Save writes the key file at path, readable by its owner only. It never
// replaces one: that would lose a key marshal may still hold.
func (k KeyFile) Save(path string) error {
	b, err := json.MarshalIndent(k, "", "  ")
	if err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if errors.Is(err, os.ErrExist) {
		return fmt.Errorf("a key is already at %s; move it away first", path)
	}
	if err != nil {
		return err
	}
	_, werr := f.Write(append(b, '\n'))
	serr := f.Sync()
	if err := errors.Join(werr, serr, f.Close()); err != nil {
		os.Remove(path)
		return err
	}
	return nil
}

func random(n int) []byte {
	b := make([]byte, n)
	rand.Read(b)
	return b
}

func b64(b []byte) string { return base64.RawURLEncoding.EncodeToString(b) }

func unb64(s string) []byte {
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return nil
	}
	return b
}
