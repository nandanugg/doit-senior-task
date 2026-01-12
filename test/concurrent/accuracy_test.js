import http from "k6/http";
import { check, fail } from "k6";
import { sleep } from "k6";

// Test with fixed iterations to verify exact click count accuracy
export const options = {
   vus: 10000, // 10,000 virtual users
   iterations: 10000, // Total of 10,000 iterations (1 per VU)
   thresholds: {
      http_req_failed: ["rate<0.01"], // Less than 1% of requests should fail
      checks: ["rate>0.99"], // More than 99% of checks should pass
   },
};

const BASE_URL = __ENV.BASE_URL || "http://localhost:8080";

// Setup function runs once before test
export function setup() {
   console.log(`Testing click count accuracy against: ${BASE_URL}`);

   // Check if server is available
   const healthCheck = http.get(`${BASE_URL}/healthz`);
   if (healthCheck.status !== 200) {
      fail(
         `Server is not available at ${BASE_URL}. Expected 200 from /healthz, got ${healthCheck.status}`,
      );
   }
   console.log("Server health check passed");

   // Create a test URL
   const payload = JSON.stringify({
      long_url: "https://accuracy-test.com/concurrent",
   });

   const params = {
      headers: {
         "Content-Type": "application/json",
      },
   };

   const response = http.post(`${BASE_URL}/s`, payload, params);

   if (response.status !== 200) {
      fail(
         `Failed to create short URL. Status: ${response.status}, Body: ${response.body}`,
      );
   }

   const shortCode = JSON.parse(response.body).short_code;
   console.log(`Setup complete. Short code: ${shortCode}`);
   console.log(`Will perform 10,000 concurrent redirects (1 per VU)...`);
   console.log(`All 10,000 VUs will use the same short code: ${shortCode}`);

   // Return shortCode - this data is automatically shared with all VUs
   return { shortCode };
}

export default function (data) {
   // All VUs receive the same shortCode from setup() via the data parameter
   // This is k6's built-in way of sharing data from setup to all VUs
   const shortCode = data.shortCode;

   // Perform redirect (this increments the click count)
   // All 10,000 VUs are hitting the same URL concurrently
   const redirectRes = http.get(`${BASE_URL}/s/${shortCode}`, {
      redirects: 0, // Don't follow redirects
   });

   check(redirectRes, {
      "redirect: status is 302": (r) => r.status === 302,
   });
}

// Teardown function runs once after all VUs complete
export function teardown(data) {
   const shortCode = data.shortCode;

   console.log("\n=== Teardown: Verifying Click Count Accuracy ===");

   // Wait to ensure all stats are persisted
   sleep(2);

   // Get final statistics from the server
   const statsRes = http.get(`${BASE_URL}/stats/${shortCode}`);

   if (statsRes.status !== 200) {
      console.error(
         `Failed to get stats. Status: ${statsRes.status}, Body: ${statsRes.body}`,
      );
      fail("Could not retrieve statistics from server");
   }

   const stats = JSON.parse(statsRes.body);
   const expectedCount = 10000; // We configured 10,000 total iterations (10,000 VUs * 1 iteration)

   console.log(`\nExpected Click Count: ${expectedCount}`);
   console.log(`Actual Click Count:   ${stats.click_count}`);
   console.log(
      `Difference:           ${Math.abs(stats.click_count - expectedCount)}`,
   );

   // Verify exact match
   if (stats.click_count !== expectedCount) {
      console.error(`\n✗ FAIL: Click count mismatch!`);
      console.error(`  Expected: ${expectedCount} (10,000 VUs * 1 iteration)`);
      console.error(`  Got:      ${stats.click_count}`);
      fail(
         `Click count accuracy test failed: expected ${expectedCount}, got ${stats.click_count}`,
      );
   }

   console.log(
      `\n✓ SUCCESS: Click count is accurate (${stats.click_count}/${expectedCount})`,
   );
   console.log("=== Accuracy Test Passed ===\n");
}
