package com.acme.blog.user;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

@Service
public class UserService {

    @Autowired
    private UserRepository repository;

    public User register(String handle, String displayName) {
        return repository.save(new User(handle, displayName));
    }

    public User byHandle(String handle) {
        return repository.findByHandle(handle).orElseThrow(() ->
            new IllegalArgumentException("user not found: " + handle));
    }
}
