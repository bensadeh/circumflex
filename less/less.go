package less

import (
	_ "embed"
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

	_, _ = tempLesskeyFile.WriteString(lesskey)

	key := new(Lesskey)
	key.tempLesskeyFile = tempLesskeyFile

	return key, nil
}

func (key *Lesskey) GetPath() string {
	return key.tempLesskeyFile.Name()
}

func (key *Lesskey) Remove() {
	_ = os.Remove(key.tempLesskeyFile.Name()) //nolint:gosec // temp file created by os.CreateTemp
}
