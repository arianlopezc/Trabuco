package com.test.fixtures;

// This file is INTENTIONALLY VULNERABLE. Each block plants one OWASP
// antipattern that the corresponding owasp.<id> rule in
// .github/scripts/review-checks.sh must catch. Do NOT use any of these
// patterns in real code. CI's `audit-rules-check` job uses this file as
// a positive control for the deterministic OWASP review.
//
// Each finding's rule ID is in the comment immediately above the line so
// failure messages tell us which antipattern is missing.

import java.io.ObjectInputStream;
import java.security.MessageDigest;
import java.io.InputStream;
import javax.crypto.Cipher;
import javax.naming.InitialContext;
import org.springframework.web.client.RestTemplate;
// Stand-in for com.nimbusds.jose.JWSAlgorithm; this file is grepped, never compiled.
import com.nimbusds.jose.JWSAlgorithm;

public class PlantedVulnerabilities {

    // owasp.a02-weak-hash
    public byte[] weakHashMd5() throws Exception {
        return MessageDigest.getInstance("MD5").digest(new byte[]{1});
    }

    // owasp.a02-weak-cipher
    public Cipher weakCipherDes() throws Exception {
        return Cipher.getInstance("DES");
    }

    // owasp.a02-jwt-none
    public JWSAlgorithm jwtAlgNone() {
        return JWSAlgorithm.NONE;
    }

    // owasp.a03-runtime-exec-concat
    public Process commandInjection(String userInput) throws Exception {
        return Runtime.getRuntime().exec("ls " + userInput);
    }

    // owasp.a03-jndi-user-input
    public Object jndiInjection(String userArg) throws Exception {
        InitialContext ctx = new InitialContext();
        return ctx.lookup(userArg);
    }

    // owasp.a05-cors-wildcard — value, not a Java API call
    // (the rule matches the literal string in @CrossOrigin annotations
    //  and CorsRegistry.allowedOrigins(...)).
    public void corsWildcard() {
        // simulating: registry.addMapping("/**").allowedOrigins("*")
        Object _ = "allowedOrigins(\"*\")";
    }

    // owasp.a07-default-creds
    public String[] defaultCreds() {
        return new String[]{"admin", "admin"};
    }

    // owasp.a08-jackson-default-typing
    public void enableDefaultTyping(com.fasterxml.jackson.databind.ObjectMapper mapper) {
        mapper.activateDefaultTyping(null, null);
    }

    // owasp.a08-objectinputstream
    public Object insecureDeserialize(InputStream in) throws Exception {
        ObjectInputStream ois = new ObjectInputStream(in);
        return ois.readObject();
    }

    // owasp.a10-resttemplate-user-url
    public String ssrf(RestTemplate restTemplate, String userUrl) {
        return restTemplate.getForObject("https://example.com/" + userUrl, String.class);
    }
}
