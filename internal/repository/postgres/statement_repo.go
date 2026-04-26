package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"stellarbill-backend/internal/repository"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// StatementRepo implements repository.StatementRepository against a live Postgres database.
type StatementRepo struct {
	pool *pgxpool.Pool
}

// NewStatementRepo constructs a StatementRepo using the provided connection pool.
func NewStatementRepo(pool *pgxpool.Pool) *StatementRepo {
	return &StatementRepo{pool: pool}
}

// FindByID fetches the statement with the given ID.
// Returns repository.ErrNotFound if no row exists.
func (r *StatementRepo) FindByID(ctx context.Context, id string) (*repository.StatementRow, error) {
	const q = `
		SELECT id, subscription_id, customer_id, period_start, period_end, issued_at, 
		       total_amount, currency, kind, status, deleted_at
		FROM statements
		WHERE id = $1`

	var s repository.StatementRow
	var deletedAt *time.Time

	ctx, span := tracer.Start(ctx, "StatementRepo.FindByID",
		otel.WithAttributes(attribute.String("statement.id", id)))
	defer span.End()

	err := r.pool.QueryRow(ctx, q, id).Scan(
		&s.ID, &s.SubscriptionID, &s.CustomerID, &s.PeriodStart, &s.PeriodEnd,
		&s.IssuedAt, &s.TotalAmount, &s.Currency, &s.Kind, &s.Status, &deletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	s.DeletedAt = deletedAt
	return &s, nil
}

// ListByCustomerID fetches statements for a customer with optional filtering and pagination.
func (r *StatementRepo) ListByCustomerID(ctx context.Context, customerID string, q repository.StatementQuery) ([]*repository.StatementRow, int, error) {
	ctx, span := tracer.Start(ctx, "StatementRepo.ListByCustomerID",
		otel.WithAttributes(attribute.String("customer.id", customerID)))
	defer span.End()

	// Build WHERE clause dynamically
	whereClause := "WHERE customer_id = $1"
	args := []interface{}{customerID}
	argPos := 2

	if q.SubscriptionID != "" {
		whereClause += " AND subscription_id = $" + string(rune('0'+argPos))
		args = append(args, q.SubscriptionID)
		argPos++
	}
	if q.Kind != "" {
		whereClause += " AND kind = $" + string(rune('0'+argPos))
		args = append(args, q.Kind)
		argPos++
	}
	if q.Status != "" {
		whereClause += " AND status = $" + string(rune('0'+argPos))
		args = append(args, q.Status)
		argPos++
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM statements " + whereClause
	var totalCount int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&totalCount); err != nil {
		return nil, 0, err
	}

	// Add pagination
	page := q.Page
	if page <= 0 {
		page = 1
	}
	pageSize := q.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	query := `
		SELECT id, subscription_id, customer_id, period_start, period_end, issued_at, 
		       total_amount, currency, kind, status, deleted_at
		FROM statements ` + whereClause + `
		ORDER BY issued_at DESC
		LIMIT $` + string(rune('0'+argPos)) + ` OFFSET $` + string(rune('0'+argPos+1))
	args = append(args, pageSize, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var statements []*repository.StatementRow
	for rows.Next() {
		var s repository.StatementRow
		var deletedAt *time.Time
		if err := rows.Scan(
			&s.ID, &s.SubscriptionID, &s.CustomerID, &s.PeriodStart, &s.PeriodEnd,
			&s.IssuedAt, &s.TotalAmount, &s.Currency, &s.Kind, &s.Status, &deletedAt,
		); err != nil {
			return nil, 0, err
		}
		s.DeletedAt = deletedAt
		statements = append(statements, &s)
	}

	return statements, totalCount, nil
}
