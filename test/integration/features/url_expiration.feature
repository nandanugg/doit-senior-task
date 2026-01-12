Feature: URL Expiration
  As a URL shortener service
  I want URLs to expire after their TTL
  So that old links are automatically removed

  Scenario: Expired URL returns 404
    Given I create a short URL for "https://example.com/expired"
    When I expire the last created short URL
    And I visit the short URL
    Then the HTTP status code should be 404

  Scenario: Valid URL before expiration works correctly
    Given I create a short URL for "https://example.com/valid"
    When I visit the short URL
    Then the HTTP status code should be 302
    And I should be redirected to "https://example.com/valid"
    When I expire the last created short URL
    And I visit the short URL
    Then the HTTP status code should be 404
