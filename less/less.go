package less

import (
	_ "embed"
	"fmt"
	"os"
)

//go:embed lesskey
var lesskey string

type Lesskey struct {
	tempLesskeyFile *os.File
}

func NewLesskey() (*Lesskey, error) {
	tempLesskeyFile, err := os.CreateTemp("", "lesskey*")
	if err != nil {
		return nil, err
	}

	if _, err := tempLesskeyFile.WriteString(lesskey); err != nil {
		_ = tempLesskeyFile.Close()

		return nil, fmt.Errorf("could not write lesskey: %w", err)
	}

	if err := tempLesskeyFile.Close(); err != nil {
		return nil, fmt.Errorf("could not close lesskey file: %w", err)
	}

	key := new(Lesskey)
	key.tempLesskeyFile = tempLesskeyFile

	return key, nil
}

func (key *Lesskey) Path() string {
	return key.tempLesskeyFile.Name()
}

func (key *Lesskey) Remove() {
	_ = os.Remove(key.tempLesskeyFile.Name())
}
