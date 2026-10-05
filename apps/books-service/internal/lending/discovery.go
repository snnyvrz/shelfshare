package lending

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type discoveryOwner struct {
	OwnerID string `json:"ownerId"`
	Rank    int    `json:"rank"`
}

// @Summary Discover public copies by their owners' selected cities
// @Tags Shelves
// @Security BearerAuth
// @Produce json
// @Param mode query string true "city or radius"
// @Param cityId query string true "Origin city ID from auth-service /cities"
// @Param radiusKm query number false "City-centre radius (1–500 km)"
// @Param page query int false "Page (50 copies)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 503 {object} map[string]interface{}
// @Router /nearby/copies [get]
func (s *Service) nearbyCopies(c *gin.Context) {
	base := strings.TrimRight(os.Getenv("AUTH_API_URL"), "/")
	if base == "" {
		base = "http://localhost:3030/api/auth"
	}
	params := url.Values{}
	for _, key := range []string{"mode", "cityId", "radiusKm"} {
		params.Set(key, c.Query(key))
	}
	req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, base+"/nearby/owners?"+params.Encode(), nil)
	if err != nil {
		c.JSON(503, gin.H{"message": "Nearby discovery is unavailable"})
		return
	}
	req.Header.Set("Authorization", c.GetHeader("Authorization"))
	client := &http.Client{Timeout: 8 * time.Second, CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }}
	response, err := client.Do(req)
	if err != nil {
		c.JSON(503, gin.H{"message": "Nearby discovery is unavailable"})
		return
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		status := 503
		if response.StatusCode == 400 || response.StatusCode == 401 {
			status = response.StatusCode
		}
		c.JSON(status, gin.H{"message": "Could not resolve nearby owners; check your city, radius and session"})
		return
	}
	// Fail explicitly on oversized responses rather than silently omitting eligible owners.
	const max = 16 << 20
	body, err := io.ReadAll(io.LimitReader(response.Body, max+1))
	var result struct {
		Data []discoveryOwner `json:"data"`
	}
	if err != nil || len(body) > max || json.Unmarshal(body, &result) != nil || result.Data == nil {
		c.JSON(503, gin.H{"message": "Nearby discovery is unavailable"})
		return
	}
	for _, owner := range result.Data {
		if !validUserID(owner.OwnerID) || owner.Rank < 0 {
			c.JSON(503, gin.H{"message": "Nearby discovery is unavailable"})
			return
		}
	}
	query := nearbyCopyQuery(s.DB, result.Data)
	s.listCopies(c, query)
}

// Join the complete owner set before counting/pagination. One JSON parameter avoids
// PostgreSQL's parameter-count limit and does not depend on reader-directory pages.
func nearbyCopyQuery(db *gorm.DB, owners []discoveryOwner) *gorm.DB {
	query := db.Model(&Copy{}).Preload("Book.Author").Where("copies.visible = true AND copies.archived = false")
	if len(owners) == 0 {
		return query.Where("1 = 0")
	}
	data, _ := json.Marshal(owners)
	return query.Joins(`JOIN jsonb_to_recordset(?::jsonb) AS discovery("ownerId" text, rank int) ON discovery."ownerId" = copies.owner_id`, string(data)).
		Select("copies.*").Order("discovery.rank ASC").Order("copies.created_at DESC").Order("copies.id ASC")
}
