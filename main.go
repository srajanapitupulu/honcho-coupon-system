package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var db *sql.DB

func main() {
	// Database Connection
	dsn := os.Getenv("DB_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/honcho_db?sslmode=disable"
	}

	var err error
	db, err = sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Router Setup using Gin
	r := gin.Default()

	// API Route
	r.POST("/api/coupons", createCoupon)
	r.POST("/api/coupons/claim", claimCoupon)
	r.GET("/api/coupons/:name", getCouponDetails)

	// Start Server
	log.Println("Server starting on :8080")
	r.Run(":8080")
}


// ClaimRequest represents a request to claim a coupon for a user.
// It contains the user ID and the name of the coupon to be claimed.
type ClaimRequest struct {
	UserID     string `json:"user_id" binding:"required"`
	CouponName string `json:"coupon_name" binding:"required"`
}

func claimCoupon(c *gin.Context) {
	var req ClaimRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Start Database Transaction
	tx, err := db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	defer tx.Rollback()

	// Insert user claim record (Enforces one-per-user)
	_, err = tx.Exec("INSERT INTO claims (user_id, coupon_name) VALUES ($1, $2)", req.UserID, req.CouponName)
	if err != nil {
		// IF error is duplicate key:
		// RESPONSE ERROR: You have already claimed this coupon
		c.JSON(http.StatusConflict, gin.H{"error": "You have already claimed this coupon"})
		return
	}

	// Decrease coupon count for each claim
	result, err := tx.Exec("UPDATE coupons SET remaining_count = remaining_count - 1 WHERE name = $1 AND remaining_count > 0", req.CouponName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		// IF no rows were affected:
		// RESPONSE ERROR: Coupon sold out or does not exist
		c.JSON(http.StatusBadRequest, gin.H{"error": "Coupon sold out or does not exist"})
		return
	}

	// Commit Database Transaction
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to finalize claim"})
		return
	}

	// IF user claim and coupon decrement successful:
	// RESPONSE SUCCESS: Coupon claimed successfully
	c.JSON(http.StatusOK, gin.H{"message": "Coupon claimed successfully!"})
}


// CreateCouponRequest represents the request payload for creating a new coupon.
// It contains the coupon's name and the discount amount to be applied.
type CreateCouponRequest struct {
	Name   string `json:"name" binding:"required"`
	Amount int    `json:"amount" binding:"required"`
}

func createCoupon(c *gin.Context) {
	var req CreateCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	_, err := db.Exec("INSERT INTO coupons (name, total_limit, remaining_count) VALUES ($1, $2, $3)", req.Name, req.Amount, req.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create coupon"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Coupon created successfully!"})
}



// CouponDetailResponse represents the details of a coupon,
// including its configured total limit, how many remain, and list of users who have claimed it.
type CouponDetailResponse struct {
	Name           	string 		`json:"name"`
	TotalLimit     	int    		`json:"amount"`
	RemainingCount 	int    		`json:"remaining_amount"`
	ClaimedBy		[]string    `json:"claimed_by"`
}

func getCouponDetails(c *gin.Context) {
	couponName := c.Param("name")

	// Fetch coupon details
	var coupon CouponDetailResponse
	err := db.QueryRow("SELECT name, total_limit, remaining_count FROM coupons WHERE name = $1", couponName).
		Scan(&coupon.Name, &coupon.TotalLimit, &coupon.RemainingCount)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Coupon not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		}
		return
	}

	// Fetch list of users who have claimed the coupon
	rows, err := db.Query("SELECT user_id FROM claims WHERE coupon_name = $1", couponName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
			return
		}
		coupon.ClaimedBy = append(coupon.ClaimedBy, userID)
	}

	c.JSON(http.StatusOK, coupon)
}