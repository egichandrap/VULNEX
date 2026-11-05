package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/vulnex/vulnex/backend/internal/authorization/domain"
	"github.com/vulnex/vulnex/backend/internal/platform/postgres"
)

var (
	ErrAuthorizationNotFound = errors.New("authorization not found")
)

// AuthorizationRepository implements domain.AuthorizationRepository backed by PostgreSQL.
type AuthorizationRepository struct {
	db postgres.DBTX
}

func NewAuthorizationRepository(db postgres.DBTX) *AuthorizationRepository {
	return &AuthorizationRepository{db: db}
}

func (r *AuthorizationRepository) Save(ctx context.Context, auth *domain.Authorization) error {
	methodsJSON, err := json.Marshal(auth.AllowedMethods())
	if err != nil {
		return fmt.Errorf("failed to marshal allowed methods: %w", err)
	}

	query := `
		INSERT INTO authorizations (
			id, project_id, status, authorized_party, approver_user_id,
			evidence_reference, evidence_checksum, valid_from, valid_until,
			allowed_methods, revoked_at, revocation_reason, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err = r.db.ExecContext(
		ctx,
		query,
		auth.ID().String(),
		auth.ProjectID().String(),
		string(auth.Status()),
		auth.AuthorizedParty(),
		auth.ApproverUserID().String(),
		auth.EvidenceReference(),
		auth.EvidenceChecksum(),
		auth.ValidFrom(),
		auth.ValidUntil(),
		methodsJSON,
		auth.RevokedAt(),
		auth.RevocationReason(),
		auth.CreatedAt(),
	)
	if err != nil {
		return fmt.Errorf("failed to save authorization: %w", err)
	}

	// Insert scope rules
	for _, rule := range auth.ScopeRules() {
		scopeQuery := `
			INSERT INTO authorization_scopes (
				authorization_id, rule_type, target_type, canonical_target, port_range, notes
			)
			VALUES ($1, $2, $3, $4, $5, $6)
		`
		_, err := r.db.ExecContext(
			ctx,
			scopeQuery,
			auth.ID().String(),
			string(rule.RuleType),
			string(rule.TargetType),
			rule.CanonicalTarget,
			rule.PortRange,
			rule.Notes,
		)
		if err != nil {
			return fmt.Errorf("failed to save authorization scope rule: %w", err)
		}
	}

	return nil
}

func (r *AuthorizationRepository) FindByID(ctx context.Context, id domain.AuthorizationID) (*domain.Authorization, error) {
	query := `
		SELECT id, project_id, status, authorized_party, approver_user_id,
		       evidence_reference, evidence_checksum, valid_from, valid_until,
		       allowed_methods, revoked_at, revocation_reason, created_at
		FROM authorizations
		WHERE id = $1
	`
	row := r.db.QueryRowContext(ctx, query, id.String())

	var (
		idStr            string
		projIDStr        string
		statusStr        string
		authorizedParty  string
		approverIDStr    string
		evidenceRef      string
		evidenceChecksum string
		validFrom        time.Time
		validUntil       time.Time
		methodsRaw       []byte
		revokedAt        *time.Time
		revocationReason sql.NullString
		createdAt        time.Time
	)

	err := row.Scan(
		&idStr, &projIDStr, &statusStr, &authorizedParty, &approverIDStr,
		&evidenceRef, &evidenceChecksum, &validFrom, &validUntil,
		&methodsRaw, &revokedAt, &revocationReason, &createdAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAuthorizationNotFound
		}
		return nil, fmt.Errorf("failed to scan authorization: %w", err)
	}

	authID, err := domain.NewAuthorizationID(idStr)
	if err != nil {
		return nil, fmt.Errorf("corrupted auth ID: %w", err)
	}

	projID, err := domain.NewProjectID(projIDStr)
	if err != nil {
		return nil, fmt.Errorf("corrupted project ID: %w", err)
	}

	approverID, err := domain.NewUserID(approverIDStr)
	if err != nil {
		return nil, fmt.Errorf("corrupted approver ID: %w", err)
	}

	var allowedMethods []string
	if len(methodsRaw) > 0 {
		_ = json.Unmarshal(methodsRaw, &allowedMethods)
	}

	rules, err := r.loadScopeRules(ctx, authID)
	if err != nil {
		return nil, err
	}

	return domain.ReconstituteAuthorization(
		authID,
		projID,
		domain.AuthorizationStatus(statusStr),
		authorizedParty,
		approverID,
		evidenceRef,
		evidenceChecksum,
		validFrom,
		validUntil,
		allowedMethods,
		revokedAt,
		revocationReason.String,
		rules,
		createdAt,
	), nil
}

func (r *AuthorizationRepository) FindByProjectID(ctx context.Context, projectID domain.ProjectID) ([]*domain.Authorization, error) {
	query := `
		SELECT id, project_id, status, authorized_party, approver_user_id,
		       evidence_reference, evidence_checksum, valid_from, valid_until,
		       allowed_methods, revoked_at, revocation_reason, created_at
		FROM authorizations
		WHERE project_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, projectID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to query authorizations: %w", err)
	}
	defer rows.Close()

	var list []*domain.Authorization
	for rows.Next() {
		var (
			idStr            string
			projIDStr        string
			statusStr        string
			authorizedParty  string
			approverIDStr    string
			evidenceRef      string
			evidenceChecksum string
			validFrom        time.Time
			validUntil       time.Time
			methodsRaw       []byte
			revokedAt        *time.Time
			revocationReason sql.NullString
			createdAt        time.Time
		)

		err := rows.Scan(
			&idStr, &projIDStr, &statusStr, &authorizedParty, &approverIDStr,
			&evidenceRef, &evidenceChecksum, &validFrom, &validUntil,
			&methodsRaw, &revokedAt, &revocationReason, &createdAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan authorization row: %w", err)
		}

		authID, err := domain.NewAuthorizationID(idStr)
		if err != nil {
			return nil, fmt.Errorf("corrupted auth ID: %w", err)
		}

		projID, err := domain.NewProjectID(projIDStr)
		if err != nil {
			return nil, fmt.Errorf("corrupted project ID: %w", err)
		}

		approverID, err := domain.NewUserID(approverIDStr)
		if err != nil {
			return nil, fmt.Errorf("corrupted approver ID: %w", err)
		}

		var allowedMethods []string
		if len(methodsRaw) > 0 {
			_ = json.Unmarshal(methodsRaw, &allowedMethods)
		}

		list = append(list, domain.ReconstituteAuthorization(
			authID,
			projID,
			domain.AuthorizationStatus(statusStr),
			authorizedParty,
			approverID,
			evidenceRef,
			evidenceChecksum,
			validFrom,
			validUntil,
			allowedMethods,
			revokedAt,
			revocationReason.String,
			nil, // Loaded on-demand per aggregate
			createdAt,
		))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("authorization row iteration error: %w", err)
	}

	return list, nil
}

func (r *AuthorizationRepository) IsValidForExecution(ctx context.Context, id domain.AuthorizationID, at time.Time) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM authorizations
		WHERE id = $1
		  AND status = 'ACTIVE'
		  AND valid_from <= $2
		  AND valid_until >= $2
	`
	var count int
	utc := at.UTC()
	err := r.db.QueryRowContext(ctx, query, id.String(), utc).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check authorization validity: %w", err)
	}
	return count > 0, nil
}

func (r *AuthorizationRepository) Update(ctx context.Context, auth *domain.Authorization) error {
	query := `
		UPDATE authorizations
		SET status = $1, revoked_at = $2, revocation_reason = $3
		WHERE id = $4
	`
	res, err := r.db.ExecContext(
		ctx,
		query,
		string(auth.Status()),
		auth.RevokedAt(),
		auth.RevocationReason(),
		auth.ID().String(),
	)
	if err != nil {
		return fmt.Errorf("failed to update authorization: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to read rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ErrAuthorizationNotFound
	}

	return nil
}

func (r *AuthorizationRepository) loadScopeRules(ctx context.Context, authID domain.AuthorizationID) ([]domain.ScopeRule, error) {
	query := `
		SELECT id, rule_type, target_type, canonical_target, port_range, notes
		FROM authorization_scopes
		WHERE authorization_id = $1
	`
	rows, err := r.db.QueryContext(ctx, query, authID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to load authorization scopes: %w", err)
	}
	defer rows.Close()

	var rules []domain.ScopeRule
	for rows.Next() {
		var (
			id              string
			ruleTypeStr     string
			targetTypeStr   string
			canonicalTarget string
			portRange       sql.NullString
			notes           string
		)

		if err := rows.Scan(&id, &ruleTypeStr, &targetTypeStr, &canonicalTarget, &portRange, &notes); err != nil {
			return nil, fmt.Errorf("failed to scan scope rule: %w", err)
		}

		rules = append(rules, domain.ScopeRule{
			ID:              id,
			RuleType:        domain.ScopeRuleType(ruleTypeStr),
			TargetType:      domain.ScopeTargetType(targetTypeStr),
			CanonicalTarget: canonicalTarget,
			PortRange:       portRange.String,
			Notes:           notes,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("scope rule iteration error: %w", err)
	}

	return rules, nil
}
