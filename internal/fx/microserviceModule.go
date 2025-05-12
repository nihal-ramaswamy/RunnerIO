package fx_utils

import (
	"go.uber.org/fx"
)

var MicroServicesModule = fx.Module(
	"MicroServices",
	serverModule,
)
