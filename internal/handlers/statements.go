package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"stellarbill-backend/internal/auth"

	"github.com/gin-gonic/gin"
	"stellarbill-backend/internal/repository"
	"stellarbill-backend/internal/service"
)

// NewGetStatementHandler returns a gin.HandlerFunc that retrieves a full
// statement detail using the provided StatementService.
func NewGetStatementHandler(svc service.StatementService) gin.HandlerFunc {
	return func(c *gin.Context) {
		callerID, exists := c.Get("callerID")
		if !exists {
			c.Header("Content-Type", "application/json; charset=utf-8")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		roles := auth.ExtractRoles(c)

		// 2. Validate :id path param.
		id := c.Param("id")
		if strings.TrimSpace(id) == "" {
			c.Header("Content-Type", "application/json; charset=utf-8")
			c.JSON(http.StatusBadRequest, gin.H{"error": "statement id required"})
			return
		}

		// 3. Call service.
		detail, warnings, err := svc.GetDetail(c.Request.Context(), callerID.(string), rolesToStrings(roles), id)
		if err != nil {
			c.Header("Content-Type", "application/json; charset=utf-8")
			switch err {
			case service.ErrNotFound:
				c.JSON(http.StatusNotFound, gin.H{"error": "statement not found"})
			case service.ErrDeleted:
				c.JSON(http.StatusGone, gin.H{"error": "statement has been deleted"})
			case service.ErrForbidden:
				c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			}
			return
		}

		// 4. Build response envelope.
		resp := service.ResponseEnvelope{
			APIVersion: "2025-01-01",
			Data:       detail,
			Warnings:   warnings,
		}

		c.Header("Content-Type", "application/json; charset=utf-8")
		c.JSON(http.StatusOK, resp)
	}
}

// NewListStatementsHandler returns a gin.HandlerFunc that lists billing
// statements for a customer using the provided StatementService.
func NewListStatementsHandler(svc service.StatementService) gin.HandlerFunc {
	return func(c *gin.Context) {
		callerID, exists := c.Get("callerID")
		if !exists {
			c.Header("Content-Type", "application/json; charset=utf-8")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		roles := auth.ExtractRoles(c)

		// 2. Build query from optional filters.
		q := repository.StatementQuery{
			SubscriptionID: c.Query("subscription_id"),
			Kind:           c.Query("kind"),
			Status:         c.Query("status"),
			StartAfter:     c.Query("start_after"),
			EndBefore:      c.Query("end_before"),
			StartingAfter:  c.Query("starting_after"),
			EndingBefore:   c.Query("ending_before"),
			Order:          c.Query("order"),
		}

		if l := c.Query("limit"); l != "" {
			if v, err := strconv.Atoi(l); err == nil {
				q.Limit = v
			}
		} else if ps := c.Query("page_size"); ps != "" { // backward compatibility
			if v, err := strconv.Atoi(ps); err == nil {
				q.Limit = v
			}
		}

		// 3. Call service.
		customerID := c.Query("customer_id")
		if customerID == "" {
			customerID = callerID.(string)
		}

		detail, count, warnings, err := svc.ListByCustomer(c.Request.Context(), callerID.(string), rolesToStrings(roles), customerID, q)
		if err != nil {
			c.Header("Content-Type", "application/json; charset=utf-8")
			switch err {
			case service.ErrForbidden:
				c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			}
			return
		}

		// 4. Normalise pagination values for response.
		limit := q.Limit
		if limit <= 0 {
			limit = 10
		}

		// 5. Build response envelope with pagination.
		resp := service.ResponseEnvelopeWithPagination{
			ResponseEnvelope: service.ResponseEnvelope{
				APIVersion: "2025-01-01",
				Data:       detail,
				Warnings:   warnings,
			},
			Pagination: service.PaginationMetadata{
				TotalCount: count,
				HasMore:    len(detail.Statements) >= limit,
			},
		}
		// Set next cursor if we have more results. 
		// In a real app, this would be the ID or IssuedAt of the last item.
		if resp.Pagination.HasMore && len(detail.Statements) > 0 {
			resp.Pagination.NextCursor = detail.Statements[len(detail.Statements)-1].ID
		}

		c.Header("Content-Type", "application/json; charset=utf-8")
		c.JSON(http.StatusOK, resp)
	}
}

func rolesToStrings(roles []auth.Role) []string {
	strs := make([]string, len(roles))
	for i, r := range roles {
		strs[i] = string(r)
	}
	return strs
}
