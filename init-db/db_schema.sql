-- Table 1: Coupons Table
CREATE TABLE IF NOT EXISTS coupons (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    total_limit INT NOT NULL,
    remaining_count INT NOT NULL
);

-- Table 2: Claim History Table
CREATE TABLE IF NOT EXISTS claims (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    coupon_name VARCHAR(255) NOT NULL,
    claimed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    -- CRITICAL: This enforces that a user can only claim a specific coupon name once
    UNIQUE(user_id, coupon_name)
);

-- Seed some data for testing
INSERT INTO coupons (name, total_limit, remaining_count) 
VALUES ('BLACK_FRIDAY_50', 100, 100)
ON CONFLICT DO NOTHING;