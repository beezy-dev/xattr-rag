package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"hash/fnv" // For a simple mock embedding vector
	
	"github.com/pkg/xattr"
)

// Document represents a piece of content on the file system
type Document struct {
	Path    string
	Content string
	XAttrs  map[string]string // Parsed xattrs
}

// DocumentEmbedding holds the document and its mock vector for retrieval
type DocumentEmbedding struct {
	Document   Document
	MockVector []float64 // In a real scenario, this would be a high-dimensional float vector
}

// Global "vector database" for simplicity
var documentIndex = make(map[string]DocumentEmbedding) // Key: Document Path, Value: Embedding

// --- Utility Functions ---

// checkXattrSupport attempts to set and remove a temporary xattr to check if the filesystem supports it.
func checkXattrSupport(dir string) bool {
	tempFile := filepath.Join(dir, "temp_xattr_test.txt")
	err := os.WriteFile(tempFile, []byte("test"), 0644)
	if err != nil {
		log.Printf("Warning: Could not create temp file for xattr test: %v", err)
		return false
	}
	defer os.Remove(tempFile)

	testKey := "user.test.xattr"
	testValue := "testvalue"

	err = xattr.Set(tempFile, testKey, []byte(testValue))
	if err != nil {
		// On some systems, even if xattr exists, it might not be enabled for certain filesystems.
		// Check for specific error types if needed, but a generic error is usually enough.
		log.Printf("Warning: Filesystem at '%s' does not seem to support xattr. Error: %v", dir, err)
		return false
	}

	retrievedValue, err := xattr.Get(tempFile, testKey)
	if err != nil || string(retrievedValue) != testValue {
		log.Printf("Warning: xattr could be set but not retrieved correctly or value mismatch. Error: %v", err)
		return false
	}

	err = xattr.Remove(tempFile, testKey)
	if err != nil {
		log.Printf("Warning: Could not remove test xattr. Error: %v", err)
		// This might indicate partial support or permission issues, but we'll assume it's okay for now.
	}

	log.Println("Filesystem appears to support xattr.")
	return true
}

// createDocument creates a file and sets extended attributes
func createDocument(filePath, content string, attrs map[string]string) error {
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	for key, value := range attrs {
		// xattr keys often start with "user." for non-privileged attributes
		// We'll enforce this for simplicity, but adjust if needed for other namespaces
		if !strings.HasPrefix(key, "user.") {
			key = "user." + key
		}
		err = xattr.Set(filePath, key, []byte(value))
		if err != nil {
			return fmt.Errorf("failed to set xattr %s on %s: %w", key, filePath, err)
		}
	}
	return nil
}

// readDocument reads file content and all its extended attributes
func readDocument(filePath string) (Document, error) {
	contentBytes, err := os.ReadFile(filePath)
	if err != nil {
		return Document{}, fmt.Errorf("failed to read file content: %w", err)
	}

	doc := Document{
		Path:    filePath,
		Content: string(contentBytes),
		XAttrs:  make(map[string]string),
	}

	keys, err := xattr.List(filePath)
	if err != nil {
		// Log error but continue, as some files might not have xattrs
		log.Printf("Warning: Failed to list xattrs for %s: %v", filePath, err)
	} else {
		for _, key := range keys {
			valueBytes, err := xattr.Get(filePath, key)
			if err != nil {
				log.Printf("Warning: Failed to get xattr '%s' for %s: %v", key, filePath, err)
				continue
			}
			doc.XAttrs[key] = string(valueBytes)
		}
	}
	return doc, nil
}

// generateEmbedding mocks the embedding generation
// In a real scenario, this would call an LLM embedding API or a local model.
// Here, we concatenate content and xattrs into a single string and then hash it.
func generateEmbedding(doc Document) DocumentEmbedding {
	// Construct the text that would be fed to an embedding model
	textForEmbedding := doc.Content
	xattrList := []string{}
	for k, v := range doc.XAttrs {
		xattrList = append(xattrList, fmt.Sprintf("%s=%s", k, v))
	}
	if len(xattrList) > 0 {
		textForEmbedding += "\n\nExtended Attributes: " + strings.Join(xattrList, ", ")
	}

	// Mocking the embedding vector with a simple FNV hash
	h := fnv.New64a()
	h.Write([]byte(textForEmbedding))
	hashValue := float64(h.Sum64())

	// Create a simple mock vector (e.g., 2 dimensions for demonstration)
	// In reality, this would be hundreds or thousands of dimensions
	mockVector := []float64{hashValue, float64(len(textForEmbedding))}

	fmt.Printf("  -> Generated embedding for '%s' (Text for embedding: '%s'...) with mock vector: %v\n",
		filepath.Base(doc.Path), textForEmbedding[:min(len(textForEmbedding), 100)], mockVector)

	return DocumentEmbedding{
		Document:   doc,
		MockVector: mockVector,
	}
}

// min helper for string slicing
func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}

// indexDocument adds a document's embedding to our mock database
func indexDocument(doc Document) {
	embedding := generateEmbedding(doc)
	documentIndex[doc.Path] = embedding
}

// retrieveDocuments performs retrieval with xattr-based filtering
func retrieveDocuments(query string, userContext map[string]string) []Document {
	fmt.Printf("\n--- Retrieving documents for query: '%s' with user context: %v ---\n", query, userContext)
	var relevantDocuments []Document

	// Mock query embedding (not strictly needed for this xattr-focused demo, but good practice)
	// In a real RAG, you'd embed the query here and use it for semantic similarity.
	// queryEmbedding := generateEmbedding(Document{Content: query}) // Example

	for _, docEmb := range documentIndex {
		doc := docEmb.Document
		fmt.Printf("  Checking document: '%s'\n", filepath.Base(doc.Path))

		// --- XAttr-based Filtering Logic ---
		// This is where user context is compared to document xattrs.
		// If *any* required xattr condition is not met, the document is skipped.

		// Example 1: User ID matching
		if requiredUserID, ok := userContext["user_id"]; ok {
			if docUserID, docOk := doc.XAttrs["user.user_id"]; docOk {
				if docUserID != requiredUserID {
					fmt.Printf("    - Skipped: User ID mismatch (Doc:%s, User:%s)\n", docUserID, requiredUserID)
					continue // Skip this document
				} else {
					fmt.Printf("    - User ID matched (%s)\n", docUserID)
				}
			} else {
				fmt.Printf("    - Document has no user_id xattr, but user context requires one. Consider policy.\n")
				// Depending on policy, you might skip or include. For now, let's include if not explicitly restricted.
			}
		}

		// Example 2: Location matching (if specified in user context)
		if requiredLocation, ok := userContext["location"]; ok {
			if docLocation, docOk := doc.XAttrs["user.location"]; docOk {
				if docLocation != requiredLocation {
					fmt.Printf("    - Skipped: Location mismatch (Doc:%s, User:%s)\n", docLocation, requiredLocation)
					continue // Skip this document
				} else {
					fmt.Printf("    - Location matched (%s)\n", docLocation)
				}
			} else {
				fmt.Printf("    - Document has no location xattr, but user context requires one. Consider policy.\n")
			}
		}

		// Example 3: Sensitivity Level (only allow if user has sufficient clearance, or specific department)
		if docSensitivity, ok := doc.XAttrs["user.sensitivity"]; ok {
			switch docSensitivity {
			case "confidential":
				// Only users with 'department: IT' or 'user_id: 123' can access confidential docs
				if userContext["department"] != "IT" && userContext["user_id"] != "123" {
					fmt.Printf("    - Skipped: Not authorized for 'confidential' document.\n")
					continue
				} else {
					fmt.Printf("    - Authorized for 'confidential' document.\n")
				}
			case "internal":
				// Allow 'internal' for anyone in the organization (mocked by having a user_id)
				if _, ok := userContext["user_id"]; !ok {
					fmt.Printf("    - Skipped: 'Internal' document requires a user ID in context.\n")
					continue
				} else {
					fmt.Printf("    - Authorized for 'internal' document.\n")
				}
			}
		}
		// --- End XAttr Filtering ---

		// In a real RAG, you'd now compare `queryEmbedding` with `docEmb.MockVector`
		// and add to `relevantDocuments` based on similarity score.
		// For this demo, if it passes the xattr filter, it's considered "relevant".
		fmt.Printf("    - **Included:** Document passed xattr filters.\n")
		relevantDocuments = append(relevantDocuments, doc)
	}

	return relevantDocuments
}

// simulateLLMResponse mocks passing retrieved content to an LLM
func simulateLLMResponse(query string, retrievedDocs []Document) {
	fmt.Println("\n--- Simulating LLM Response ---")
	if len(retrievedDocs) == 0 {
		fmt.Printf("LLM received: No relevant documents found for query '%s'.\n", query)
		fmt.Println("LLM Response: I couldn't find any information relevant to your request based on the available data and your permissions.")
		return
	}

	var contextBuilder strings.Builder
	contextBuilder.WriteString(fmt.Sprintf("User query: %s\n\n", query))
	contextBuilder.WriteString("Retrieved information:\n")
	for i, doc := range retrievedDocs {
		contextBuilder.WriteString(fmt.Sprintf("### Document %d (Path: %s)\n", i+1, filepath.Base(doc.Path)))
		contextBuilder.WriteString(fmt.Sprintf("Content: %s\n", doc.Content))
		contextBuilder.WriteString("XAttrs: ")
		xattrList := []string{}
		for k, v := range doc.XAttrs {
			xattrList = append(xattrList, fmt.Sprintf("%s=%s", k, v))
		}
		contextBuilder.WriteString(strings.Join(xattrList, ", "))
		contextBuilder.WriteString("\n\n")
	}

	fmt.Println("LLM would receive the following prompt context:")
	fmt.Println("--------------------------------------------------")
	fmt.Println(contextBuilder.String())
	fmt.Println("--------------------------------------------------")
	fmt.Println("\nLLM Response (Simulated): Based on the retrieved documents and your query, I can provide information about...")
	// A real LLM would generate a coherent answer here.
}

func actualMain() error {
	// Create a temporary directory for our "filesystem"
	tmpDir, err := os.MkdirTemp("", "xattr_rag_demo")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer func() {
		fmt.Printf("\nCleaning up temporary directory: %s\n", tmpDir)
		if err := os.RemoveAll(tmpDir); err != nil {
			log.Printf("Warning: Failed to clean up temp directory %s: %v", tmpDir, err)
		}
	}()

	fmt.Printf("Using temporary directory: %s\n", tmpDir)

	if !checkXattrSupport(tmpDir) {
		return fmt.Errorf("xattr is not supported on this filesystem or Go environment. Exiting")
	}

	// --- 1. Create Sample Documents with xattrs ---
	fmt.Println("\n--- Creating Sample Documents ---")

	// Public Document
	publicDocPath := filepath.Join(tmpDir, "public_announcement.txt")
	err = createDocument(publicDocPath, "This is a public announcement about upcoming office changes.",
		map[string]string{"type": "public", "published_by": "HR"})
	if err != nil {
		return fmt.Errorf("error creating public doc: %w", err)
	}
	fmt.Printf("Created: %s\n", publicDocPath)

	// User 123's Personal Document
	user123DocPath := filepath.Join(tmpDir, "user_123_personal_notes.txt")
	err = createDocument(user123DocPath, "My personal notes about project Alpha. Do not share.",
		map[string]string{"user_id": "123", "sensitivity": "confidential", "location": "New York"})
	if err != nil {
		return fmt.Errorf("error creating user 123 doc: %w", err)
	}
	fmt.Printf("Created: %s\n", user123DocPath)

	// User 456's Internal Report
	user456DocPath := filepath.Join(tmpDir, "internal_dev_report.txt")
	err = createDocument(user456DocPath, "Internal development report for Q3. Access restricted to Dev department.",
		map[string]string{"user_id": "456", "department": "Dev", "sensitivity": "internal"})
	if err != nil {
		return fmt.Errorf("error creating user 456 doc: %w", err)
	}
	fmt.Printf("Created: %s\n", user456DocPath)

	// User 123's IT Document
	user123ITDocPath := filepath.Join(tmpDir, "it_security_policy.txt")
	err = createDocument(user123ITDocPath, "Official IT security policy. For IT department only.",
		map[string]string{"user_id": "123", "department": "IT", "sensitivity": "confidential"})
	if err != nil {
		return fmt.Errorf("error creating IT doc: %w", err)
	}
	fmt.Printf("Created: %s\n", user123ITDocPath)

	// --- 2. Read and Index Documents (Pre-computation for RAG) ---
	fmt.Println("\n--- Reading and Indexing Documents ---")
	dirEntries, err := os.ReadDir(tmpDir)
	if err != nil {
		return fmt.Errorf("failed to read temp directory: %w", err)
	}

	for _, entry := range dirEntries {
		if !entry.IsDir() && !strings.HasPrefix(entry.Name(), "temp_xattr_test") { // Skip temp xattr test file
			doc, err := readDocument(filepath.Join(tmpDir, entry.Name()))
			if err != nil {
				log.Printf("Error reading document %s: %v", entry.Name(), err)
				continue
			}
			indexDocument(doc)
		}
	}
	fmt.Printf("Indexed %d documents.\n", len(documentIndex))

	// --- 3. Simulate User Queries with Different Contexts ---

	// Scenario 1: User 123 querying for personal notes
	query1 := "What are my project notes?"
	userContext1 := map[string]string{"user_id": "123", "location": "New York"}
	retrieved1 := retrieveDocuments(query1, userContext1)
	simulateLLMResponse(query1, retrieved1)

	// Scenario 2: User 456 querying for "personal notes" (should NOT get 123's notes)
	query2 := "Show me my personal notes."
	userContext2 := map[string]string{"user_id": "456", "location": "London"}
	retrieved2 := retrieveDocuments(query2, userContext2)
	simulateLLMResponse(query2, retrieved2)

	// Scenario 3: Any user querying for public information
	query3 := "Tell me about office changes."
	userContext3 := map[string]string{} // No specific user context, treated as public
	retrieved3 := retrieveDocuments(query3, userContext3)
	simulateLLMResponse(query3, retrieved3)

	// Scenario 4: User 123 (who is in IT) querying for IT policy
	query4 := "What's the IT security policy?"
	userContext4 := map[string]string{"user_id": "123", "department": "IT", "location": "New York"}
	retrieved4 := retrieveDocuments(query4, userContext4)
	simulateLLMResponse(query4, retrieved4)

	// Scenario 5: User 456 (Dev department) querying for IT policy (should NOT get it due to department/sensitivity)
	query5 := "What's the IT security policy?"
	userContext5 := map[string]string{"user_id": "456", "department": "Dev", "location": "London"}
	retrieved5 := retrieveDocuments(query5, userContext5)
	simulateLLMResponse(query5, retrieved5)

	// Scenario 6: User 456 querying for internal development report
	query6 := "Latest development report."
	userContext6 := map[string]string{"user_id": "456", "department": "Dev", "location": "London"}
	retrieved6 := retrieveDocuments(query6, userContext6)
	simulateLLMResponse(query6, retrieved6)

	return nil
}

func main() {
	if err := actualMain(); err != nil {
		log.Printf("Error: %v", err)
		os.Exit(1)
	}
}