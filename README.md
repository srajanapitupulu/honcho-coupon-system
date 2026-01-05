# Scalable Coupon Flash Sale System

This project is a high-performance Go-based microservice designed to handle high-concurrency "Flash Sale" scenarios where multiple users attempt to claim limited coupons simultaneously.

## 🚀 Architectural Overview
[cite_start]With over **14 years of software engineering experience**, I have designed this system to prioritize **Consistency and Atomicity (ACID)**. While this is a Go-centric project, it follows the same rigorous **Clean Architecture and SOLID principles** I have applied throughout my career as a Senior Engineer at companies like **Gramedia and Alterra**.[cite_end]

### Key Technical Implementations:
* **Atomic Concurrency Control**: Used a "Pessimistic Locking" strategy at the database level (`UPDATE ... WHERE remaining_count > 0`) to prevent over-claiming during traffic spikes.
* **Idempotency**: Utilized PostgreSQL composite unique constraints to ensure "Double Dip" protection, ensuring each user can only claim a specific coupon once.
* **Modular Design (pkg/ pattern)**: Structured the project to separate business logic from transport layers, adhering to **Separation of Concerns**.
* **Resilience Testing**: Included a dedicated stress test suite simulating concurrent "attacks" to prove the system's stability.

---

## 🛠 Prerequisites
To run this system, ensure you have the following installed:
1.  **Go (1.25+)**
2.  **Docker & Docker Compose**
3.  **PostgreSQL (15+)** (If running outside Docker)

---

## ⚙️ Installation & Setup

1.  **Clone the Repository:**
    ```bash
    git clone https://github.com/srajanapitupulu/honcho-coupon-system.git
    cd honcho-coupon-system
    ```

2.  **Launch via Docker:**
    The system is fully containerized. This command builds the Go binary and initializes the PostgreSQL database.
    ```bash
    docker compose up --build
    ```

3.  **Database Migration:**
    The schema (tables: `coupons`, `claims`) is automatically applied upon startup via the SQL initialization script.

---

## 🧪 Testing

### 1. Manual Verification
**Create a Coupon:**
```bash
curl -X POST http://localhost:8080/api/coupons \
     -H "Content-Type: application/json" \
     -d '{"name": "PROMO_2026", "amount": 10}'
```

**Claim a Coupon:**
```bash
curl -X POST http://localhost:8080/api/coupons/claim \
     -H "Content-Type: application/json" \
     -d '{"user_id": "user_777", "coupon_name": "PROMO_2026"}'
```


### 2. Automated Stress Tests

To run the high-concurrency evaluation scenarios (The "Flash Sale" Attack & The "Double Dip" Attack):

```bash
go test -v ./unit-test/...
```

## 📈 Roadmap for Production Scaling
To prepare this for an enterprise-level deployment, I would implement the following:

1. **RabbitMQ Integration:** Decouple the API from the database using a message broker to "level" traffic spikes and implement Dead Letter Queues (DLQ) for retry strategies.
2. **Distributed Caching (Redis):** Cache coupon counts to reduce DB read pressure and implement distributed locking for multi-node setups.
3. **Saga Pattern:** Utilize distributed transaction patterns for workflows involving external services like loyalty points or payments.
4. **Observability:** Integrate Prometheus and Grafana to monitor Goroutine health and DB connection pools.

## 📄 License
This project is licensed under the GNU General Public License v3.0 (GPL-3.0). See the LICENSE file for details.

## 📬 Contact

**Samuel Oloan Raja Napitupulu**
Senior Software Engineer 


📍 Medan, North Sumatera, Indonesia 
📧 srajanapitupulu@gmail.com 
🔗 LinkedIn: [Samuel Oloan Raja Napitupulu](https://www.linkedin.com/in/samuel-oloan-raja-napitupulu-98008b67)
🖥️ GitHub: [srajanapitupulu](https://github.com/srajanapitupulu)