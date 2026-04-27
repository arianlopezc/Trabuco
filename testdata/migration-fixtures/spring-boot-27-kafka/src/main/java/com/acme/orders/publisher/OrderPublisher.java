package com.acme.orders.publisher;

import com.acme.orders.event.OrderPlacedEvent;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.kafka.core.KafkaTemplate;
import org.springframework.stereotype.Service;

@Service
public class OrderPublisher {

    @Autowired
    private KafkaTemplate<String, OrderPlacedEvent> kafkaTemplate;

    public void publish(OrderPlacedEvent event) {
        kafkaTemplate.send("orders.placed", event.getOrderId(), event);
    }
}
