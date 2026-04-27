package com.acme.shop.customer;

import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;

import static org.assertj.core.api.Assertions.assertThat;

// Legacy uses @SpringBootTest for everything (Trabuco test specialist
// will mark these as ADAPT to sliced tests like @DataJpaTest).
@SpringBootTest
class CustomerServiceTest {

    @Autowired
    private CustomerService service;

    @Test
    void create_thenFindAll() {
        Customer c = service.create(new Customer("Alice", "alice@example.com"));
        assertThat(c.getId()).isNotNull();
        assertThat(service.findAll()).extracting(Customer::getEmail).contains("alice@example.com");
    }
}
