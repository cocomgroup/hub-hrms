package main

import (
    "fmt"
    "golang.org/x/crypto/bcrypt"
)

func main() {
    password := "admin123"
    hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    fmt.Printf("UPDATE users SET password_hash = '%s' WHERE email = 'admin@cocomgroup.com';\n", string(hash))
}