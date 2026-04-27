package com.acme.shop.customer;

public class CustomerNotFoundException extends RuntimeException {
    public CustomerNotFoundException(String msg) { super(msg); }
}
