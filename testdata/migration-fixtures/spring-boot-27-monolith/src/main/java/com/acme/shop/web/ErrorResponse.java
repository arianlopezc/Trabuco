package com.acme.shop.web;

import java.time.Instant;

/**
 * Legacy bespoke error envelope. Spring Boot 2.7-style; not RFC 7807.
 * Trabuco's API specialist surfaces LEGACY_ERROR_FORMAT_REQUIRED for
 * the user to decide whether to migrate to ProblemDetail.
 */
public class ErrorResponse {
    private final String code;
    private final String message;
    private final Instant timestamp = Instant.now();

    public ErrorResponse(String code, String message) {
        this.code = code;
        this.message = message;
    }

    public String getCode() { return code; }
    public String getMessage() { return message; }
    public Instant getTimestamp() { return timestamp; }
}
