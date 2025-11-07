package data

import (
	"github.com/google/wire"
	"kratos-realworld/internal/pkg/middleware/sms"
)

var ProviderSet = wire.NewSet(
	NewUserRepo,
	NewProfileRepo,
	NewMessageRepo,
	NewSmsRepo,
	sms.NewSmsService,
)
