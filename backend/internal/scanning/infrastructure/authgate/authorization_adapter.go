package authgate

import (
	"context"
	"fmt"
	"time"

	authDomain "github.com/vulnex/vulnex/backend/internal/authorization/domain"
	scanApp "github.com/vulnex/vulnex/backend/internal/scanning/application"
)

// AuthorizationAdapter implements scanApp.AuthorizationValidatorPort by querying authDomain.AuthorizationRepository.
type AuthorizationAdapter struct {
	repo authDomain.AuthorizationRepository
}

func NewAuthorizationAdapter(repo authDomain.AuthorizationRepository) *AuthorizationAdapter {
	return &AuthorizationAdapter{repo: repo}
}

func (a *AuthorizationAdapter) VerifyAuthorization(ctx context.Context, authIDStr string) (bool, []scanApp.ScopeRuleDTO, error) {
	authID, err := authDomain.NewAuthorizationID(authIDStr)
	if err != nil {
		return false, nil, fmt.Errorf("invalid authorization ID: %w", err)
	}

	auth, err := a.repo.FindByID(ctx, authID)
	if err != nil {
		return false, nil, fmt.Errorf("failed to load authorization: %w", err)
	}

	if !auth.IsValidAt(time.Now().UTC()) {
		return false, nil, nil
	}

	var dtos []scanApp.ScopeRuleDTO
	for _, r := range auth.ScopeRules() {
		dtos = append(dtos, scanApp.ScopeRuleDTO{
			RuleType:   string(r.RuleType),
			TargetType: string(r.TargetType),
			Target:     r.CanonicalTarget,
		})
	}

	return true, dtos, nil
}
