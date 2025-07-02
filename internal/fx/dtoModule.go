package fx_utils

import (
	wsdto "github.com/nihal-ramaswamy/RunnerIO/internal/dto/ws"
	"go.uber.org/fx"
)

var DTOModule = fx.Module(
	"DTO",
	fx.Provide(wsdto.NewPersistAuditDataManagerMap),
)
