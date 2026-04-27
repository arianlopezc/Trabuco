package com.acme.blog.post;

import org.springframework.data.annotation.Id;
import org.springframework.data.mongodb.core.mapping.Document;
import org.springframework.data.mongodb.core.index.Indexed;

import java.time.Instant;
import java.util.List;

@Document(collection = "posts")
public class Post {

    @Id
    private String id;

    @Indexed
    private String authorId;

    private String title;
    private String body;
    private List<String> tags;
    private Instant createdAt;

    public Post() {}

    public Post(String authorId, String title, String body, List<String> tags) {
        this.authorId = authorId;
        this.title = title;
        this.body = body;
        this.tags = tags;
        this.createdAt = Instant.now();
    }

    public String getId() { return id; }
    public void setId(String id) { this.id = id; }
    public String getAuthorId() { return authorId; }
    public void setAuthorId(String authorId) { this.authorId = authorId; }
    public String getTitle() { return title; }
    public void setTitle(String title) { this.title = title; }
    public String getBody() { return body; }
    public void setBody(String body) { this.body = body; }
    public List<String> getTags() { return tags; }
    public void setTags(List<String> tags) { this.tags = tags; }
    public Instant getCreatedAt() { return createdAt; }
    public void setCreatedAt(Instant createdAt) { this.createdAt = createdAt; }
}
