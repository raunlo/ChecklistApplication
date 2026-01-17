package legal

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// GetPrivacyPolicy serves the privacy policy as markdown
func GetPrivacyPolicy(c *gin.Context) {
	content, err := os.ReadFile("docs/privacy-policy.md")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Policy not found",
			"message": "Privacy policy file could not be read",
		})
		return
	}

	// Serve as markdown with proper content type
	c.Data(http.StatusOK, "text/markdown; charset=utf-8", content)
}

// GetPrivacyPolicyJSON serves the privacy policy metadata as JSON
func GetPrivacyPolicyJSON(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"title":       "Privacy Policy",
		"last_updated": "December 2025",
		"url":         "/legal/privacy-policy",
		"summary":     "We collect minimal data: only Google User ID and your checklists. No email, name, or tracking.",
	})
}
