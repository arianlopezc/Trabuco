package com.acme.shop.customer;

import java.util.List;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

@Service
public class CustomerService {

    // Field injection — legacy pattern Trabuco's shared specialist replaces with constructor injection.
    @Autowired
    private CustomerRepository repository;

    @Transactional(readOnly = true)
    public List<Customer> findAll() {
        return repository.findAll();
    }

    @Transactional
    public Customer create(Customer c) {
        return repository.save(c);
    }

    @Transactional(readOnly = true)
    public Customer findById(Long id) {
        return repository.findById(id)
            .orElseThrow(() -> new CustomerNotFoundException("customer not found: " + id));
    }
}
