package main

import (
	"fmt"
)

// ============================================================================
// EXAMPLE 1: Why logger is *logger.Logger but queries is LinkQueries
// ============================================================================

// This is like your logger - a CONCRETE struct
type Logger struct {
	level string
	count int
}

// This is like your LinkQueries - an INTERFACE
type Database interface {
	Query() string
}

// This is like your db.Queries - implements the interface
type Queries struct {
	conn string
}

func (q *Queries) Query() string {
	return "SELECT * FROM links"
}

// Service struct - notice the difference!
type Service struct {
	db     Database  // Interface - NO pointer needed!
	logger *Logger   // Concrete struct - pointer needed!
}

func NewService(db Database, logger *Logger) *Service {
	return &Service{
		db:     db,     // No & needed - interface handles it
		logger: logger, // Already a pointer, no & needed
	}
}

// ============================================================================
// EXAMPLE 2: When to use & when passing arguments
// ============================================================================

// Function that MODIFIES the value - needs pointer
func modifyValue(x *int) {
	*x = 42  // Must dereference to modify
}

// Function that READS the value - no pointer needed
func readValue(x int) int {
	return x * 2
}

// Function that modifies a struct - needs pointer
func fillUser(u *User) {
	u.Name = "John"
	u.Age = 30
}

type User struct {
	Name string
	Age  int
}

// ============================================================================
// EXAMPLE 3: Optional fields with pointers
// ============================================================================

type CreateLinkRequest struct {
	URL       string  // Required - no pointer
	Shortcode *string // Optional - pointer (can be nil)
	ExpiresAt *string // Optional - pointer (can be nil)
}

func createLink(req CreateLinkRequest) {
	fmt.Printf("URL: %s\n", req.URL)
	
	if req.Shortcode != nil {
		fmt.Printf("Custom code: %s\n", *req.Shortcode)  // Dereference!
	} else {
		fmt.Println("Using random code")
	}
	
	if req.ExpiresAt != nil {
		fmt.Printf("Expires: %s\n", *req.ExpiresAt)
	} else {
		fmt.Println("No expiration")
	}
}

// ============================================================================
// EXAMPLE 4: Return pointers vs values
// ============================================================================

// Returns pointer - for large structs or shared instances
func NewServicePointer() *Service {
	return &Service{}  // Returns pointer
}

// Returns value - for small, immutable types
func generateID() string {
	return "abc123"  // Returns value
}

// ============================================================================
// EXAMPLE 5: Common mistakes
// ============================================================================

func commonMistakes() {
	// MISTAKE 1: Forgetting & when function needs pointer
	var x int = 10
	// modifyValue(x)  // ❌ WRONG - compile error!
	modifyValue(&x)    // ✅ CORRECT
	fmt.Println(x)     // 42

	// MISTAKE 2: Using & when not needed
	var y int = 5
	result := readValue(y)  // ✅ No & needed
	// result := readValue(&y)  // ❌ WRONG - compile error!
	fmt.Println(result)      // 10

	// MISTAKE 3: Nil pointer dereference
	var logger *Logger
	// logger.count = 5  // ❌ PANIC! nil pointer
	logger = &Logger{level: "info", count: 0}  // ✅ Initialize first
	logger.count = 5

	// MISTAKE 4: Forgetting to dereference
	shortcode := "my-link"
	req := CreateLinkRequest{
		URL:       "https://example.com",
		Shortcode: &shortcode,  // ✅ Use & to get pointer
	}
	
	// fmt.Println(req.Shortcode)  // ❌ Prints memory address!
	fmt.Println(*req.Shortcode)   // ✅ Dereference to get value
}

// ============================================================================
// EXAMPLE 6: Interface vs Concrete Type
// ============================================================================

func interfaceVsConcrete() {
	// Create concrete type (pointer)
	queries := &Queries{conn: "postgres://..."}
	
	// Assign to interface - NO & needed!
	var db Database = queries  // ✅ Go converts automatically
	
	// Create concrete struct
	logger := &Logger{level: "info", count: 0}
	
	// Create service
	svc := NewService(db, logger)  // ✅ No & needed for db (interface)
	                                // ✅ logger already a pointer
	
	fmt.Printf("Service: %+v\n", svc)
}

// ============================================================================
// MAIN - Run examples
// ============================================================================

func main() {
	fmt.Println("=== Example 1: Interface vs Concrete ===")
	interfaceVsConcrete()
	
	fmt.Println("\n=== Example 2: Optional Fields ===")
	customCode := "my-custom-link"
	req := CreateLinkRequest{
		URL:       "https://example.com",
		Shortcode: &customCode,  // Use & to pass pointer
	}
	createLink(req)
	
	fmt.Println("\n=== Example 3: Common Mistakes ===")
	commonMistakes()
}

