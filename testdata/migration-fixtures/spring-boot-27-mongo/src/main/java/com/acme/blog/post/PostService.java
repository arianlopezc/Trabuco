package com.acme.blog.post;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.data.domain.PageRequest;
import org.springframework.stereotype.Service;

import java.util.List;

@Service
public class PostService {

    @Autowired
    private PostRepository repository;

    public Post create(Post post) {
        return repository.save(post);
    }

    public List<Post> byAuthor(String authorId, int page, int size) {
        return repository.findByAuthorIdOrderByCreatedAtDesc(authorId, PageRequest.of(page, size));
    }

    public List<Post> byTag(String tag) {
        return repository.findByTagsContaining(tag);
    }
}
