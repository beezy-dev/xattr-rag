# xattr-rag
Showcasing the usage of extended attributes with RAG


```
go run xattr-rag-demo.go
```

```
go: downloading github.com/pkg/xattr v0.4.10
go: downloading golang.org/x/sys v0.0.0-20220408201424-a24fb2fb8a0f
Using temporary directory: /var/folders/fx/qwmr6n3564v9bv6qbydq27ww0000gp/T/xattr_rag_demo2116708620
2025/05/28 19:14:29 Filesystem appears to support xattr.

--- Creating Sample Documents ---
Created: /var/folders/fx/qwmr6n3564v9bv6qbydq27ww0000gp/T/xattr_rag_demo2116708620/public_announcement.txt
Created: /var/folders/fx/qwmr6n3564v9bv6qbydq27ww0000gp/T/xattr_rag_demo2116708620/user_123_personal_notes.txt
Created: /var/folders/fx/qwmr6n3564v9bv6qbydq27ww0000gp/T/xattr_rag_demo2116708620/internal_dev_report.txt
Created: /var/folders/fx/qwmr6n3564v9bv6qbydq27ww0000gp/T/xattr_rag_demo2116708620/it_security_policy.txt

--- Reading and Indexing Documents ---
  -> Generated embedding for 'internal_dev_report.txt' (Text for embedding: 'Internal development report for Q3. Access restricted to Dev department.

Extended Attributes: com.a'...) with mock vector: [8.033560895293351e+18 193]
  -> Generated embedding for 'it_security_policy.txt' (Text for embedding: 'Official IT security policy. For IT department only.

Extended Attributes: com.apple.provenance=�'...) with mock vector: [1.734426386575502e+19 176]
  -> Generated embedding for 'public_announcement.txt' (Text for embedding: 'This is a public announcement about upcoming office changes.

Extended Attributes: com.apple.provena'...) with mock vector: [1.8395543918480054e+19 155]
  -> Generated embedding for 'user_123_personal_notes.txt' (Text for embedding: 'My personal notes about project Alpha. Do not share.

Extended Attributes: com.apple.provenance=�'...) with mock vector: [3.3900277131379364e+18 180]
Indexed 4 documents.

--- Retrieving documents for query: 'What are my project notes?' with user context: map[location:New York user_id:123] ---
  Checking document: 'it_security_policy.txt'
    - User ID matched (123)
    - Document has no location xattr, but user context requires one. Consider policy.
    - Authorized for 'confidential' document.
    - **Included:** Document passed xattr filters.
  Checking document: 'public_announcement.txt'
    - Document has no user_id xattr, but user context requires one. Consider policy.
    - Document has no location xattr, but user context requires one. Consider policy.
    - **Included:** Document passed xattr filters.
  Checking document: 'user_123_personal_notes.txt'
    - User ID matched (123)
    - Location matched (New York)
    - Authorized for 'confidential' document.
    - **Included:** Document passed xattr filters.
  Checking document: 'internal_dev_report.txt'
    - Skipped: User ID mismatch (Doc:456, User:123)

--- Simulating LLM Response ---
LLM would receive the following prompt context:
--------------------------------------------------
User query: What are my project notes?

Retrieved information:
### Document 1 (Path: it_security_policy.txt)
Content: Official IT security policy. For IT department only.
XAttrs: com.apple.provenance=��ƨEe�, user.department=IT, user.sensitivity=confidential, user.user_id=123

### Document 2 (Path: public_announcement.txt)
Content: This is a public announcement about upcoming office changes.
XAttrs: com.apple.provenance=��ƨEe�, user.published_by=HR, user.type=public

### Document 3 (Path: user_123_personal_notes.txt)
Content: My personal notes about project Alpha. Do not share.
XAttrs: com.apple.provenance=��ƨEe�, user.location=New York, user.sensitivity=confidential, user.user_id=123
```

```
--------------------------------------------------

LLM Response (Simulated): Based on the retrieved documents and your query, I can provide information about...

--- Retrieving documents for query: 'Show me my personal notes.' with user context: map[location:London user_id:456] ---
  Checking document: 'internal_dev_report.txt'
    - User ID matched (456)
    - Document has no location xattr, but user context requires one. Consider policy.
    - Authorized for 'internal' document.
    - **Included:** Document passed xattr filters.
  Checking document: 'it_security_policy.txt'
    - Skipped: User ID mismatch (Doc:123, User:456)
  Checking document: 'public_announcement.txt'
    - Document has no user_id xattr, but user context requires one. Consider policy.
    - Document has no location xattr, but user context requires one. Consider policy.
    - **Included:** Document passed xattr filters.
  Checking document: 'user_123_personal_notes.txt'
    - Skipped: User ID mismatch (Doc:123, User:456)

--- Simulating LLM Response ---
LLM would receive the following prompt context:
--------------------------------------------------
User query: Show me my personal notes.

Retrieved information:
### Document 1 (Path: internal_dev_report.txt)
Content: Internal development report for Q3. Access restricted to Dev department.
XAttrs: com.apple.provenance=��ƨEe�, user.department=Dev, user.sensitivity=internal, user.user_id=456

### Document 2 (Path: public_announcement.txt)
Content: This is a public announcement about upcoming office changes.
XAttrs: com.apple.provenance=��ƨEe�, user.published_by=HR, user.type=public


--------------------------------------------------

LLM Response (Simulated): Based on the retrieved documents and your query, I can provide information about...

--- Retrieving documents for query: 'Tell me about office changes.' with user context: map[] ---
  Checking document: 'user_123_personal_notes.txt'
    - Skipped: Not authorized for 'confidential' document.
  Checking document: 'internal_dev_report.txt'
    - Skipped: 'Internal' document requires a user ID in context.
  Checking document: 'it_security_policy.txt'
    - Skipped: Not authorized for 'confidential' document.
  Checking document: 'public_announcement.txt'
    - **Included:** Document passed xattr filters.

--- Simulating LLM Response ---
LLM would receive the following prompt context:
--------------------------------------------------
User query: Tell me about office changes.

Retrieved information:
### Document 1 (Path: public_announcement.txt)
Content: This is a public announcement about upcoming office changes.
XAttrs: com.apple.provenance=��ƨEe�, user.published_by=HR, user.type=public


--------------------------------------------------

LLM Response (Simulated): Based on the retrieved documents and your query, I can provide information about...

--- Retrieving documents for query: 'What's the IT security policy?' with user context: map[department:IT location:New York user_id:123] ---
  Checking document: 'public_announcement.txt'
    - Document has no user_id xattr, but user context requires one. Consider policy.
    - Document has no location xattr, but user context requires one. Consider policy.
    - **Included:** Document passed xattr filters.
  Checking document: 'user_123_personal_notes.txt'
    - User ID matched (123)
    - Location matched (New York)
    - Authorized for 'confidential' document.
    - **Included:** Document passed xattr filters.
  Checking document: 'internal_dev_report.txt'
    - Skipped: User ID mismatch (Doc:456, User:123)
  Checking document: 'it_security_policy.txt'
    - User ID matched (123)
    - Document has no location xattr, but user context requires one. Consider policy.
    - Authorized for 'confidential' document.
    - **Included:** Document passed xattr filters.

--- Simulating LLM Response ---
LLM would receive the following prompt context:
--------------------------------------------------
User query: What's the IT security policy?

Retrieved information:
### Document 1 (Path: public_announcement.txt)
Content: This is a public announcement about upcoming office changes.
XAttrs: com.apple.provenance=��ƨEe�, user.published_by=HR, user.type=public

### Document 2 (Path: user_123_personal_notes.txt)
Content: My personal notes about project Alpha. Do not share.
XAttrs: com.apple.provenance=��ƨEe�, user.location=New York, user.sensitivity=confidential, user.user_id=123

### Document 3 (Path: it_security_policy.txt)
Content: Official IT security policy. For IT department only.
XAttrs: user.sensitivity=confidential, user.user_id=123, com.apple.provenance=��ƨEe�, user.department=IT


--------------------------------------------------

LLM Response (Simulated): Based on the retrieved documents and your query, I can provide information about...

--- Retrieving documents for query: 'What's the IT security policy?' with user context: map[department:Dev location:London user_id:456] ---
  Checking document: 'internal_dev_report.txt'
    - User ID matched (456)
    - Document has no location xattr, but user context requires one. Consider policy.
    - Authorized for 'internal' document.
    - **Included:** Document passed xattr filters.
  Checking document: 'it_security_policy.txt'
    - Skipped: User ID mismatch (Doc:123, User:456)
  Checking document: 'public_announcement.txt'
    - Document has no user_id xattr, but user context requires one. Consider policy.
    - Document has no location xattr, but user context requires one. Consider policy.
    - **Included:** Document passed xattr filters.
  Checking document: 'user_123_personal_notes.txt'
    - Skipped: User ID mismatch (Doc:123, User:456)

--- Simulating LLM Response ---
LLM would receive the following prompt context:
--------------------------------------------------
User query: What's the IT security policy?

Retrieved information:
### Document 1 (Path: internal_dev_report.txt)
Content: Internal development report for Q3. Access restricted to Dev department.
XAttrs: com.apple.provenance=��ƨEe�, user.department=Dev, user.sensitivity=internal, user.user_id=456

### Document 2 (Path: public_announcement.txt)
Content: This is a public announcement about upcoming office changes.
XAttrs: user.type=public, com.apple.provenance=��ƨEe�, user.published_by=HR


--------------------------------------------------

LLM Response (Simulated): Based on the retrieved documents and your query, I can provide information about...

--- Retrieving documents for query: 'Latest development report.' with user context: map[department:Dev location:London user_id:456] ---
  Checking document: 'internal_dev_report.txt'
    - User ID matched (456)
    - Document has no location xattr, but user context requires one. Consider policy.
    - Authorized for 'internal' document.
    - **Included:** Document passed xattr filters.
  Checking document: 'it_security_policy.txt'
    - Skipped: User ID mismatch (Doc:123, User:456)
  Checking document: 'public_announcement.txt'
    - Document has no user_id xattr, but user context requires one. Consider policy.
    - Document has no location xattr, but user context requires one. Consider policy.
    - **Included:** Document passed xattr filters.
  Checking document: 'user_123_personal_notes.txt'
    - Skipped: User ID mismatch (Doc:123, User:456)

--- Simulating LLM Response ---
LLM would receive the following prompt context:
--------------------------------------------------
User query: Latest development report.

Retrieved information:
### Document 1 (Path: internal_dev_report.txt)
Content: Internal development report for Q3. Access restricted to Dev department.
XAttrs: com.apple.provenance=��ƨEe�, user.department=Dev, user.sensitivity=internal, user.user_id=456

### Document 2 (Path: public_announcement.txt)
Content: This is a public announcement about upcoming office changes.
XAttrs: user.published_by=HR, user.type=public, com.apple.provenance=��ƨEe�


--------------------------------------------------

LLM Response (Simulated): Based on the retrieved documents and your query, I can provide information about...

Cleaning up temporary directory: /var/folders/fx/qwmr6n3564v9bv6qbydq27ww0000gp/T/xattr_rag_demo2116708620
```
