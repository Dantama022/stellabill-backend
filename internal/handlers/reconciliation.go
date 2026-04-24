package handlers

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "stellabill-backend/internal/reconciliation"
)

// NewReconcileHandler returns a handler that accepts a list of backend subscriptions
// (JSON array) and compares them against snapshots fetched from the provided Adapter.
// If a non-nil store is provided, reports will be persisted.
// Request body: [{subscription_id,...}, ...]
func NewReconcileHandler(adapter reconciliation.Adapter, store reconciliation.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		var backendSubs []reconciliation.BackendSubscription
		if err := c.ShouldBindJSON(&backendSubs); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Create reconciliation service with deterministic retry logic
		service := reconciliation.NewService(adapter, store)
		
		// Perform reconciliation with retry logic
		reports, err := service.Reconcile(c.Request.Context(), backendSubs,
			reconciliation.WithMaxAttempts(3),
			reconciliation.WithBaseDelay(1*time.Second),
			reconciliation.WithMaxDelay(30*time.Second),
		)
		
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "reconciliation failed: " + err.Error()})
			return
		}

		// summary
		matched := 0
		for _, r := range reports {
			if r.Matched {
				matched++
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"summary": gin.H{"total": len(reports), "matched": matched, "mismatched": len(reports) - matched},
			"reports": reports,
		})
	}
}
