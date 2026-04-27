package com.acme.shop.customer;

import java.util.List;
import javax.validation.Valid;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequestMapping("/api/customers")
public class CustomerController {

    @Autowired
    private CustomerService service;

    @GetMapping
    public List<Customer> list() {
        return service.findAll();
    }

    @PostMapping
    public ResponseEntity<Customer> create(@Valid @RequestBody Customer c) {
        return ResponseEntity.ok(service.create(c));
    }

    @GetMapping("/{id}")
    public Customer get(@PathVariable Long id) {
        return service.findById(id);
    }
}
