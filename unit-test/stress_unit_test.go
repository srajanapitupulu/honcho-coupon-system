package unittest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	coupon "github.com-personal/srajanapitupulu/honcho-coupon-system/pkg/coupon"
	"github.com/stretchr/testify/assert"
)

// Scenario 1: The "Flash Sale" Attack
// 50 concurrent requests for a coupon with only 5 items in stock.
func TestFlashSaleAttack(t *testing.T) {
	couponName := "FLASH_SALE_5"
	limit := 5
	totalRequests := 50

	// Create the limited coupon
	createBody, _ := json.Marshal(map[string]interface{}{
		"name":   couponName,
		"amount": limit,
	})
	req, _ := http.NewRequest("POST", "/api/coupons", bytes.NewBuffer(createBody))
	router := coupon.SetupRouter(testDB)
	router.ServeHTTP(httptest.NewRecorder(), req)

	// Prepare concurrent requests
	var wg sync.WaitGroup
	results := make(chan int, totalRequests)

	for i := 0; i < totalRequests; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			claimBody, _ := json.Marshal(map[string]interface{}{
				"user_id":     fmt.Sprintf("user_%d", id),
				"coupon_name": couponName,
			})
			req, _ := http.NewRequest("POST", "/api/coupons/claim", bytes.NewBuffer(claimBody))
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			results <- w.Code
		}(i)
	}

	wg.Wait()
	close(results)

	// Verify results
	successCount := 0
	for code := range results {
		if code == http.StatusOK {
			successCount++
		}
	}

	assert.Equal(t, limit, successCount, "Flash Sale failed: System allowed more claims than stock!")
	
	// Verify DB state
	var remaining int
	testDB.QueryRow("SELECT remaining_count FROM coupons WHERE name = $1", couponName).Scan(&remaining)
	assert.Equal(t, 0, remaining, "Flash Sale failed: Remaining count should be 0")
}

// Scenario 2: The "Double Dip" Attack
// 10 concurrent requests from the SAME user for the same coupon.
func TestDoubleDipAttack(t *testing.T) {
	couponName := "DOUBLE_DIP_COUPON"
	userID := "greedy_user_88"

	// Create a coupon with plenty of stock
	createBody, _ := json.Marshal(map[string]interface{}{
		"name":   couponName,
		"amount": 100,
	})
	router := coupon.SetupRouter(testDB)
	req, _ := http.NewRequest("POST", "/api/coupons", bytes.NewBuffer(createBody))
	router.ServeHTTP(httptest.NewRecorder(), req)

	// Fire 10 requests for the same user simultaneously
	var wg sync.WaitGroup
	totalRequests := 10
	results := make(chan int, totalRequests)

	for i := 0; i < totalRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			claimBody, _ := json.Marshal(map[string]interface{}{
				"user_id":     userID,
				"coupon_name": couponName,
			})
			req, _ := http.NewRequest("POST", "/api/coupons/claim", bytes.NewBuffer(claimBody))
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			results <- w.Code
		}()
	}

	wg.Wait()
	close(results)

	// Verify exactly 1 claim success for the user
	successCount := 0
	for code := range results {
		if code == http.StatusOK {
			successCount++
		}
	}

	assert.Equal(t, 1, successCount, "Double Dip failed: User claimed the same coupon multiple times!")
}