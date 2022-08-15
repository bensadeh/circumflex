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

func NewLesskey() *Lesskey {
	tempLesskeyFile, _ := os.CreateTemp("", "lesskey*")
	_, _ = tempLesskeyFile.WriteString(lesskey)

	key := new(Lesskey)
	key.tempLesskeyFile = tempLesskeyFile

	return key
}

func (key *Lesskey) GetPath() string {
	return key.tempLesskeyFile.Name()
}

func (key *Lesskey) Remove() {
	_ = os.Remove(key.tempLesskeyFile.Name())
}
