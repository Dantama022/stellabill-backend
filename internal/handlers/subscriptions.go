package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"stellarbill-backend/internal/requestparams"
)

type Subscription struct {
	ID          string `json:"id"`
	PlanID      string `json:"plan_id"`
	Customer    string `json:"customer"`
	Status      string `json:"status"`
	Amount      string `json:"amount"`
	Interval    string `json:"interval"`
	NextBilling string `json:"next_billing,omitempty"`
}

func ListSubscriptions(c *gin.Context) {
	if _, err := requestparams.SanitizeQuery(c.Request.URL.Query(), requestparams.QueryRules{
		Strings: map[string]requestparams.StringRule{
			"customer": requestparams.IdentifierRule(64),
			"plan_id":  requestparams.IdentifierRule(64),
			"status":   requestparams.EnumRule(16, true, "active", "past_due", "canceled", "trialing"),
		},
		Ints: map[string]requestparams.IntRule{
			"limit": {Min: 1, Max: 100},
			"page":  {Min: 1, Max: 100000},
		},
	}); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: load from DB, filter by merchant from JWT/API key
	subscriptions := []Subscription{}
	c.JSON(http.StatusOK, gin.H{"subscriptions": subscriptions})
}

func GetSubscription(c *gin.Context) {
	if _, err := requestparams.SanitizeQuery(c.Request.URL.Query(), requestparams.QueryRules{}); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := requestparams.NormalizePathID("id", c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: load from DB by id
	c.JSON(http.StatusOK, gin.H{
		"id":           id,
		"plan_id":      "plan_placeholder",
		"customer":     "customer_placeholder",
		"status":       "placeholder",
		"amount":       "0",
		"interval":     "monthly",
		"next_billing": "2026-04-01T00:00:00Z",
	})
}
