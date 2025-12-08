package main

import (
	"fmt"
	"os"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run hash_password.go <password>")
		os.Exit(1)
	}
	
	password := os.Args[1]
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println("Error generating hash:", err)
		os.Exit(1)
	}
	
	fmt.Println("Password:", password)
	fmt.Println("Bcrypt Hash:")
	fmt.Println(string(hash))
	fmt.Println()
	
	// Verify it works
	err = bcrypt.CompareHashAndPassword(hash, []byte(password))
	if err != nil {
		fmt.Println("❌ Verification FAILED!")
		os.Exit(1)
	} else {
		fmt.Println("✓ Verification successful!")
	}
	
	// SQL update command
	fmt.Println()
	fmt.Println("SQL Command to update user:")
	fmt.Printf("UPDATE users SET password_hash = '%s' WHERE email = 'admin@cocomgroup.com';\n", string(hash))
}