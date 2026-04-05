package provider

import (
	"github.com/bensadeh/circumflex/hn"
	"github.com/bensadeh/circumflex/hn/services/firebase"
	"github.com/bensadeh/circumflex/hn/services/mock"
)

func NewService(debugMode, debugFallible bool) hn.Service {
	if debugFallible {
		return mock.NewFallibleService()
	}

	if debugMode {
		return mock.Service{}
	}

	return firebase.NewService()
}
