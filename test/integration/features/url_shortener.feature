Feature: URL Shortener
  As a user
  I want to shorten URLs
  So that I can share them easily

  Scenario: Create a short URL
    Given the service is running
    When I create a short URL for "https://example.com"
    Then I should receive a short code
    And the short code should be valid
    And the response should have processing time header

  Scenario: Redirect to original URL
    Given the service is running
    And I have created a short URL for "https://example.com/test"
    When I visit the short URL
    Then I should be redirected to "https://example.com/test"
    And the response should have processing time header

  Scenario: Track click statistics
    Given the service is running
    And I have created a short URL for "https://stats-test.com"
    When I visit the short URL 3 times
    And I check the statistics
    Then the click count should be 3
    And the response should have processing time header

  Scenario: Get statistics for URL
    Given the service is running
    And I have created a short URL for "https://analytics.com"
    When I check the statistics
    Then I should see the long URL "https://analytics.com"
    And I should see the click count
    And I should see the creation time
    And I should see the expiration time
    And the response should have processing time header

  # Note: Expired URL testing is implemented in expired_url_synctest_test.go
  # using Go 1.25's testing/synctest for deterministic time control.
  # This allows testing expiration without waiting for real time to pass.
