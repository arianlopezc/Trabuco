package com.acme.orders.consumer;

import com.acme.orders.event.OrderPlacedEvent;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.kafka.annotation.KafkaListener;
import org.springframework.stereotype.Component;

@Component
public class OrderConsumer {

    private static final Logger log = LoggerFactory.getLogger(OrderConsumer.class);

    @KafkaListener(topics = "orders.placed", groupId = "orders-events-consumer")
    public void onOrderPlaced(OrderPlacedEvent event) {
        log.info("received OrderPlaced order={} customer={} total={}",
            event.getOrderId(), event.getCustomerId(), event.getTotalCents());
    }
}
