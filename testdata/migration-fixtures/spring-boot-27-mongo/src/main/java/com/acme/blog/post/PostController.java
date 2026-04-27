package com.acme.blog.post;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/posts")
public class PostController {

    @Autowired
    private PostService service;

    @PostMapping
    public Post create(@RequestBody Post post) {
        return service.create(post);
    }

    @GetMapping
    public List<Post> list(
        @RequestParam String authorId,
        @RequestParam(defaultValue = "0") int page,
        @RequestParam(defaultValue = "20") int size
    ) {
        return service.byAuthor(authorId, page, size);
    }
}
