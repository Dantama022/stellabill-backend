package handlers

import (
	"net/http"
	"strconv"

	"stellarbill-backend/internal/pagination"

	"github.com/gin-gonic/gin"
)

type Plan struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Amount      string `json:"amount"` // Used as secondary sort value if needed
	Currency    string `json:"currency"`
	Interval    string `json:"interval"`
	Description string `json:"description,omitempty"`
}

// Ensure Plan implements pagination.Item for in-memory processing
func (p Plan) GetID() string        { return p.ID }
func (p Plan) GetSortValue() string { return p.Amount }

func ListPlans(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit value"})
		return
	}

	cursorStr := c.Query("cursor")
	cursor, err := pagination.Decode(cursorStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid cursor format"})
		return
	}

	// TODO: load from DB using limit and cursor
	// For now, we mock an in-memory DB so you can test it in your browser
	mockDB := []Plan{
		{ID: "pln_1", Name: "Basic", Amount: "1000", Currency: "usd", Interval: "month"},
		{ID: "pln_2", Name: "Pro", Amount: "2000", Currency: "usd", Interval: "month"},
		{ID: "pln_3", Name: "Enterprise", Amount: "5000", Currency: "usd", Interval: "month"},
		{ID: "pln_4", Name: "Ultimate", Amount: "9000", Currency: "usd", Interval: "year"},
	}
	
	page, nextCursor, hasMore := pagination.PaginateSlice(mockDB, cursor, limit)

	c.JSON(http.StatusOK, gin.H{
		"data": page,
		"pagination": gin.H{
			"next_cursor": pagination.Encode(nextCursor),
			"has_more":    hasMore,
		},
	})
}
