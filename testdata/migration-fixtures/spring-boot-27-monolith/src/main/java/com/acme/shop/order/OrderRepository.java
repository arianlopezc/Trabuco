package com.acme.shop.order;

import java.util.List;
import org.springframework.data.domain.Pageable;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

@Repository
public interface OrderRepository extends JpaRepository<Order, Long> {

    // Legacy offset-pagination — Trabuco datastore specialist will surface
    // OFFSET_PAGINATION_INCOMPATIBLE and offer keyset migration.
    List<Order> findByCustomerId(Long customerId, Pageable pageable);
}
