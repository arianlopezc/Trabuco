package com.acme.blog.post;

import org.springframework.data.domain.Pageable;
import org.springframework.data.mongodb.repository.MongoRepository;

import java.util.List;

public interface PostRepository extends MongoRepository<Post, String> {

    List<Post> findByAuthorIdOrderByCreatedAtDesc(String authorId, Pageable pageable);

    List<Post> findByTagsContaining(String tag);
}
