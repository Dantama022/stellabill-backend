package service

import (
	"context"

	"stellarbill-backend/internal/repository"
)

// StatementService defines the business logic interface for billing statements.
type StatementService interface {
	GetDetail(ctx context.Context, callerID string, roles []string, statementID string) (*StatementDetail, []string, error)
	ListByCustomer(ctx context.Context, callerID string, roles []string, customerID string, q repository.StatementQuery) (*ListStatementsDetail, int, []string, error)
}

// statementService is the concrete implementation of StatementService.
type statementService struct {
	subRepo  repository.SubscriptionRepository
	stmtRepo repository.StatementRepository
}

// NewStatementService constructs a StatementService with the given repositories.
func NewStatementService(subRepo repository.SubscriptionRepository, stmtRepo repository.StatementRepository) StatementService {
	return &statementService{subRepo: subRepo, stmtRepo: stmtRepo}
}

// GetDetail retrieves a full StatementDetail for the given statementID.
// It enforces RBAC (admin/merchant/owner) and handles soft-deletes.
func (s *statementService) GetDetail(ctx context.Context, callerID string, roles []string, statementID string) (*StatementDetail, []string, error) {
	var warnings []string

	// 1. Fetch statement row.
	row, err := s.stmtRepo.FindByID(ctx, statementID)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, nil, ErrNotFound
		}
		return nil, nil, err
	}

	// 2. Soft-delete check.
	if row.DeletedAt != nil {
		return nil, nil, ErrDeleted
	}

	// 3. RBAC/Ownership check.
	isAuthorized := false
	for _, role := range roles {
		if role == "admin" {
			isAuthorized = true
			break
		}
		if role == "merchant" {
			// In a real app, we'd check if this merchant manages this customer.
			// For now, merchants are allowed to see any statement if they have the perm.
			isAuthorized = true
			break
		}
	}
	if !isAuthorized && callerID == row.CustomerID {
		isAuthorized = true
	}

	if !isAuthorized {
		return nil, nil, ErrForbidden
	}

	// 4. Build StatementDetail.
	detail := &StatementDetail{
		ID:             row.ID,
		SubscriptionID: row.SubscriptionID,
		Customer:       row.CustomerID,
		PeriodStart:    row.PeriodStart,
		PeriodEnd:      row.PeriodEnd,
		IssuedAt:       row.IssuedAt,
		TotalAmount:    row.TotalAmount,
		Currency:       row.Currency,
		Kind:           row.Kind,
		Status:         row.Status,
	}

	// 5. Return detail and warnings.
	return detail, warnings, nil
}

// ListByCustomer retrieves a list of StatementDetails for the given customerID, filtered and paginated according to the query parameters.
func (s *statementService) ListByCustomer(ctx context.Context, callerID string, roles []string, customerID string, q repository.StatementQuery) (*ListStatementsDetail, int, []string, error) {
	var warnings []string

	// 1. RBAC/Ownership check.
	isAuthorized := false
	for _, role := range roles {
		if role == "admin" {
			isAuthorized = true
			break
		}
		if role == "merchant" {
			isAuthorized = true
			break
		}
	}
	if !isAuthorized && callerID == customerID {
		isAuthorized = true
	}

	if !isAuthorized {
		return nil, 0, nil, ErrForbidden
	}

	// 2. Fetch statement rows for customer with filters and pagination.
	rows, count, err := s.stmtRepo.ListByCustomerID(ctx, customerID, q)
	if err != nil {
		return nil, 0, nil, err
	}

	// 3. Build StatementDetail slice.
	result := &ListStatementsDetail{
		Statements: make([]*StatementDetail, 0, len(rows)),
	}
	for _, row := range rows {
		result.Statements = append(result.Statements, &StatementDetail{
			ID:             row.ID,
			SubscriptionID: row.SubscriptionID,
			Customer:       row.CustomerID,
			PeriodStart:    row.PeriodStart,
			PeriodEnd:      row.PeriodEnd,
			IssuedAt:       row.IssuedAt,
			TotalAmount:    row.TotalAmount,
			Currency:       row.Currency,
			Kind:           row.Kind,
			Status:         row.Status,
		})
	}

	// 4. Return details, total count, and warnings.
	return result, count, warnings, nil
}
