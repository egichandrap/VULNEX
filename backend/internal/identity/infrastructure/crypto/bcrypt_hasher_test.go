package crypto

import "testing"

func TestBcryptPasswordHasher_HashAndCompare(t *testing.T) {
	hasher := NewBcryptPasswordHasher(10)

	password := "Correct-Horse-Battery-Staple-99"
	hash, err := hasher.Hash(password)
	if err != nil {
		t.Fatalf("unexpected hash error: %v", err)
	}

	if hash == password || len(hash) < 20 {
		t.Fatalf("invalid password hash format: %s", hash)
	}

	if !hasher.Compare(password, hash) {
		t.Errorf("expected compare to return true for correct password")
	}

	if hasher.Compare("Wrong-Password-12345", hash) {
		t.Errorf("expected compare to return false for incorrect password")
	}
}

func TestBcryptPasswordHasher_EmptyPasswordFails(t *testing.T) {
	hasher := NewBcryptPasswordHasher(10)
	_, err := hasher.Hash("")
	if err == nil {
		t.Error("expected error hashing empty password, got nil")
	}
}
