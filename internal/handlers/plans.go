package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"stellarbill-backend/internal/requestparams"
)

type Plan struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Amount      string `json:"amount"`
	Currency    string `json:"currency"`
	Interval    string `json:"interval"`
	Description string `json:"description,omitempty"`
}

func ListPlans(c *gin.Context) {
	if _, err := requestparams.SanitizeQuery(c.Request.URL.Query(), requestparams.QueryRules{
		Strings: map[string]requestparams.StringRule{
			"currency": requestparams.CurrencyRule(),
			"interval": requestparams.EnumRule(16, true, "daily", "weekly", "monthly", "yearly"),
			"search":   requestparams.SearchRule(64),
		},
		Ints: map[string]requestparams.IntRule{
			"limit": {Min: 1, Max: 100},
			"page":  {Min: 1, Max: 100000},
		},
	}); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: load from DB, filter by merchant
	plans := []Plan{}
	c.JSON(http.StatusOK, gin.H{"plans": plans})
}
