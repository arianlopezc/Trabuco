package com.acme.orders.event;

import java.time.Instant;

public class OrderPlacedEvent {
    private String orderId;
    private String customerId;
    private long totalCents;
    private Instant placedAt;

    public OrderPlacedEvent() {}

    public OrderPlacedEvent(String orderId, String customerId, long totalCents, Instant placedAt) {
        this.orderId = orderId;
        this.customerId = customerId;
        this.totalCents = totalCents;
        this.placedAt = placedAt;
    }

    public String getOrderId() { return orderId; }
    public void setOrderId(String orderId) { this.orderId = orderId; }
    public String getCustomerId() { return customerId; }
    public void setCustomerId(String customerId) { this.customerId = customerId; }
    public long getTotalCents() { return totalCents; }
    public void setTotalCents(long totalCents) { this.totalCents = totalCents; }
    public Instant getPlacedAt() { return placedAt; }
    public void setPlacedAt(Instant placedAt) { this.placedAt = placedAt; }
}
