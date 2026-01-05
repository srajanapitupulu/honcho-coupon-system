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
	r.POST("/claim", claimCoupon)

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
		c.JSON(http.StatusGone, gin.H{"error": "Coupon sold out or does not exist"})
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