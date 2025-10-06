# locustfile.py - Load testing for Product API
# Tests thread safety of concurrent reads/writes

from locust import HttpUser, task, between, events
import random
import json


class ProductAPIUser(HttpUser):
    """
    Simulates users making concurrent requests to the Product API.
    Tests read/write concurrency to verify mutex locks work correctly.
    """

    # Wait 0.5-2 seconds between tasks (simulates think time)
    wait_time = between(0.5, 2)

    # Product IDs for testing (smaller range = more contention)
    PRODUCT_IDS = list(range(1, 21))  # 20 products

    def on_start(self):
        """Initialize by creating some initial products"""
        # Seed initial products
        for product_id in range(1, 6):
            self.create_product(product_id, log=False)

    def create_product(self, product_id, log=True):
        """Helper to create/update a product"""
        product_data = {
            "product_id": product_id,
            "sku": f"SKU-{product_id:03d}-{random.randint(1000, 9999)}",
            "manufacturer": random.choice([
                "Apple", "Samsung", "Dell", "HP", "Lenovo",
                "Sony", "LG", "Microsoft", "Google", "Amazon"
            ]),
            "category_id": random.randint(1, 10),
            "weight": random.randint(100, 5000),
            "some_other_id": random.randint(1, 100)
        }

        with self.client.post(
            f"/products/{product_id}/details",
            json=product_data,
            catch_response=True,
            name="/products/[id]/details [POST]"
        ) as response:
            if response.status_code == 204:
                if log:
                    response.success()
            else:
                response.failure(f"Expected 204, got {response.status_code}")

    @task(7)  # 70% of traffic
    def read_product(self):
        """
        Read a random product (simulates GET requests).
        High frequency to test concurrent reads with RLock.
        """
        product_id = random.choice(self.PRODUCT_IDS)

        with self.client.get(
            f"/products/{product_id}",
            catch_response=True,
            name="/products/[id] [GET]"
        ) as response:
            if response.status_code == 200:
                try:
                    data = response.json()
                    if data.get("product_id") == product_id:
                        response.success()
                    else:
                        response.failure(
                            f"Product ID mismatch: expected {product_id}, got {data.get('product_id')}")
                except json.JSONDecodeError:
                    response.failure("Invalid JSON response")
            elif response.status_code == 404:
                # Product doesn't exist yet, that's ok
                response.success()
            else:
                response.failure(f"Unexpected status: {response.status_code}")

    @task(3)  # 30% of traffic
    def write_product(self):
        """
        Create/update a random product (simulates POST requests).
        Tests concurrent writes with Lock.
        """
        product_id = random.choice(self.PRODUCT_IDS)
        self.create_product(product_id)

    @task(1)  # 10% of traffic
    def read_write_same_product(self):
        """
        Stress test: rapidly read and write the same product.
        Tests mutex contention on same resource.
        """
        product_id = random.randint(1, 5)  # Narrow range for high contention

        # Write
        self.create_product(product_id, log=False)

        # Immediately read (race condition if mutex fails)
        with self.client.get(
            f"/products/{product_id}",
            catch_response=True,
            name="/products/[id] [GET] (after write)"
        ) as response:
            if response.status_code == 200:
                response.success()
            else:
                response.failure(
                    f"Read after write failed: {response.status_code}")

    @task(1)  # 10% additional load
    def health_check(self):
        """Health check endpoint (baseline performance)"""
        response = self.client.get("/healthz", name="/healthz")
        if response.status_code != 200:
            response.failure(f"Health check failed: {response.status_code}")


class FastProductAPIUser(HttpUser):
    """
    Fast HTTP user without keep-alive delays.
    Use this to compare performance with ProductAPIUser.
    """
    wait_time = between(0.1, 0.5)  # Faster
    PRODUCT_IDS = list(range(1, 21))

    @task(7)
    def read_product(self):
        product_id = random.choice(self.PRODUCT_IDS)
        self.client.get(f"/products/{product_id}",
                        name="/products/[id] [GET-FAST]")

    @task(3)
    def write_product(self):
        product_id = random.choice(self.PRODUCT_IDS)
        product_data = {
            "product_id": product_id,
            "sku": f"FAST-{product_id:03d}",
            "manufacturer": "FastCo",
            "category_id": 1,
            "weight": 100,
            "some_other_id": 1
        }
        self.client.post(
            f"/products/{product_id}/details",
            json=product_data,
            name="/products/[id]/details [POST-FAST]"
        )


# Event hooks for detailed reporting
@events.test_start.add_listener
def on_test_start(environment, **kwargs):
    print("\n" + "="*60)
    print("🐝 Locust Load Test Starting")
    print("="*60)
    print("Testing thread safety with concurrent reads/writes")
    print(f"Target: {environment.host}")
    print("="*60 + "\n")


@events.test_stop.add_listener
def on_test_stop(environment, **kwargs):
    print("\n" + "="*60)
    print("🏁 Locust Load Test Complete")
    print("="*60)
    print("\nKey Metrics to Review:")
    print("  - Total Requests: Check for any failures")
    print("  - Response Times: Should be consistent")
    print("  - Failures: Should be 0% (thread safety issues would cause failures)")
    print("  - RPS: Requests per second sustained")
    print("="*60 + "\n")
