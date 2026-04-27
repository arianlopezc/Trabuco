package com.acme.shop.order;

import com.acme.shop.customer.Customer;
import com.acme.shop.customer.CustomerService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.data.domain.PageRequest;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;

@Service
public class OrderService {

    @Autowired
    private OrderRepository orderRepository;

    @Autowired
    private CustomerService customerService;

    @Transactional
    public Order place(Long customerId, List<OrderLine> lines) {
        Customer customer = customerService.findById(customerId);
        Order order = new Order();
        order.setCustomer(customer);
        for (OrderLine l : lines) {
            l.setOrder(order);
            order.getLines().add(l);
        }
        return orderRepository.save(order);
    }

    @Transactional(readOnly = true)
    public List<Order> listForCustomer(Long customerId, int page, int pageSize) {
        return orderRepository.findByCustomerId(customerId, PageRequest.of(page, pageSize));
    }
}
