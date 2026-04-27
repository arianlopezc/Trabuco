package com.acme.shop.order;

import com.acme.shop.web.ErrorResponse;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/orders")
public class OrderController {

    @Autowired
    private OrderService service;

    @PostMapping("/customer/{customerId}")
    public ResponseEntity<?> place(@PathVariable Long customerId,
                                    @RequestBody List<OrderLine> lines) {
        try {
            Order o = service.place(customerId, lines);
            return ResponseEntity.ok(o);
        } catch (RuntimeException e) {
            // Bespoke ErrorResponse — Trabuco api specialist will surface
            // LEGACY_ERROR_FORMAT_REQUIRED unless user opts to migrate to RFC 7807.
            return ResponseEntity
                .status(HttpStatus.BAD_REQUEST)
                .body(new ErrorResponse("INVALID_ORDER", e.getMessage()));
        }
    }

    @GetMapping("/customer/{customerId}")
    public List<Order> list(@PathVariable Long customerId,
                            @RequestParam(defaultValue = "0") int page,
                            @RequestParam(defaultValue = "20") int size) {
        return service.listForCustomer(customerId, page, size);
    }
}
