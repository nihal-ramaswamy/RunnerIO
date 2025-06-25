package dto

import (
	"context"
	"slices"
)

type CustomClaims struct {
	Permissions []string `json:"permissions"`
}

func (c CustomClaims) Validate(ctx context.Context) error {
	return nil
}

func (c CustomClaims) HasPermissions(expectedClaims []string) bool {
	if len(expectedClaims) == 0 {
		return false
	}
	for _, scope := range expectedClaims {
		if !slices.Contains(c.Permissions, scope) {
			return false
		}
	}
	return true
}
