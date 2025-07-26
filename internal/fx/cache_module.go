package fx_utils

import (
	"context"

	"go.uber.org/fx"
)

var CacheModule = fx.Module(
	"CacheService",
	fx.Provide(func() context.Context {
		return context.Background()
	}),
)
