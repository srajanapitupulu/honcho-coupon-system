package unittest

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	coupon "github.com-personal/srajanapitupulu/honcho-coupon-system/pkg/coupon"
	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
)

// Global test variables
var testRouter *gin.Engine
var testDB *sql.DB

// TestMain runs once before all tests to setup the shared DB and Router
func TestMain(m *testing.M) {
	dsn := "postgres://postgres:postgres@localhost:5432/honcho_db?sslmode=disable"
	var err error
	testDB, err = sql.Open("pgx", dsn)
	if err != nil {
		panic("Failed to connect to test database")
	}

	// Clean database before tests
	_, err = testDB.Exec("TRUNCATE TABLE claims, coupons RESTART IDENTITY CASCADE")
    if err != nil {
        panic("Failed to clean database for testing: " + err.Error())
    }

	testRouter = coupon.SetupRouter(testDB)
	m.Run()
}

func TestCreateCoupon(t *testing.T) {
	body, _ := json.Marshal(map[string]interface{}{
		"name":   "UNIT_TEST_COUPON",
		"amount": 100,
	})
	
	req, _ := http.NewRequest("POST", "/api/coupons", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestCreateCouponDuplicates(t *testing.T) {
	body, _ := json.Marshal(map[string]interface{}{
		"name":   "UNIT_TEST_COUPON",
		"amount": 100,
	})
	
	req, _ := http.NewRequest("POST", "/api/coupons", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestClaimCouponSuccess(t *testing.T) {
	// Note: We use the coupon created in the previous step or create a new one
	body, _ := json.Marshal(map[string]interface{}{
		"user_id":     "tester_1",
		"coupon_name": "UNIT_TEST_COUPON",
	})
	
	req, _ := http.NewRequest("POST", "/api/coupons/claim", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestClaimCouponAlreadyClaimed(t *testing.T) {
	// Note: We use the coupon created in the previous step or create a new one
	body, _ := json.Marshal(map[string]interface{}{
		"user_id":     "tester_1",
		"coupon_name": "UNIT_TEST_COUPON",
	})
	
	req, _ := http.NewRequest("POST", "/api/coupons/claim", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestClaimCouponNotFound(t *testing.T) {
	// Note: We use the coupon created in the previous step or create a new one
	body, _ := json.Marshal(map[string]interface{}{
		"user_id":     "tester_1",
		"coupon_name": "TEST_COUPON_NOT_EXIST",
	})
	
	req, _ := http.NewRequest("POST", "/api/coupons/claim", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetCouponDetailsSuccess(t *testing.T) {
	req, _ := http.NewRequest("GET", "/api/coupons/UNIT_TEST_COUPON", nil)
	w := httptest.NewRecorder()
	
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	
	assert.Equal(t, "UNIT_TEST_COUPON", resp["name"])
	assert.Contains(t, resp["claimed_by"], "tester_1")
}


func TestGetCouponDetailsNotFound(t *testing.T) {
	req, _ := http.NewRequest("GET", "/api/coupons/UNIT_TEST_COUPON_NOT_FOUND", nil)
	w := httptest.NewRecorder()
	
	testRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}