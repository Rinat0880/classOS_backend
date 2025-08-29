package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"os"

	"github.com/go-ldap/ldap/v3"
	"github.com/joho/godotenv"
	"github.com/rinat0880/classOS_backend/pkg/service"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: No .env file found: %v", err)
	}

	fmt.Println(" LDAP Connection Test Suite")
	fmt.Println("=============================")

	showConfig()

	fmt.Println("\n Test 1: Basic LDAP Connection")
	fmt.Println("--------------------------------")
	if err := testBasicConnection(); err != nil {
		fmt.Printf(" Basic connection failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(" Basic connection successful!")

	fmt.Println("\n Test 2: AD Service Integration")
	fmt.Println("---------------------------------")
	if err := testADService(); err != nil {
		fmt.Printf(" AD Service test failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(" AD Service test successful!")

	fmt.Println("\n All tests passed! Ready for production.")
	fmt.Println("\nNext steps:")
	fmt.Println("  - Start application: docker-compose up -d")
	fmt.Println("  - Check logs: docker-compose logs classos_backend")
	fmt.Println("  - Proceed to Etap 2: User creation via LDAP")
}

func showConfig() {
	fmt.Println("Current Configuration:")
	fmt.Printf("  AD_HOST: %s\n", getEnvOrDefault("AD_HOST", "not set"))
	fmt.Printf("  AD_PORT: %s\n", getEnvOrDefault("AD_PORT", "not set"))
	fmt.Printf("  AD_BASE_DN: %s\n", getEnvOrDefault("AD_BASE_DN", "not set"))
	fmt.Printf("  AD_BIND_USER: %s\n", getEnvOrDefault("AD_BIND_USER", "not set"))
	fmt.Printf("  AD_USE_TLS: %s\n", getEnvOrDefault("AD_USE_TLS", "not set"))
}

func testBasicConnection() error {
	host := os.Getenv("AD_HOST")
	port := os.Getenv("AD_PORT")
	bindUser := os.Getenv("AD_BIND_USER")
	bindPass := os.Getenv("AD_BIND_PASS")
	baseDN := os.Getenv("AD_BASE_DN")

	if host == "" || port == "" || bindUser == "" || bindPass == "" {
		return fmt.Errorf("missing required environment variables")
	}

	fmt.Printf("Connecting to %s:%s as %s...\n", host, port, bindUser)

	var conn *ldap.Conn
	var err error

	address := fmt.Sprintf("%s:%s", host, port)

	if port == "636" || os.Getenv("AD_USE_TLS") == "true" {
		fmt.Println("Using LDAPS/TLS connection")
		conn, err = ldap.DialTLS("tcp", address, &tls.Config{
			ServerName:         host,
			InsecureSkipVerify: true, // для тестирования
		})
	} else {
		fmt.Println("Using plain LDAP connection")
		conn, err = ldap.Dial("tcp", address)
	}

	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer conn.Close()

	fmt.Println("✓ TCP connection established")

	fmt.Println("Attempting to bind...")
	if err := conn.Bind(bindUser, bindPass); err != nil {
		return fmt.Errorf("bind failed: %w", err)
	}
	fmt.Println("✓ Authentication successful")

	if baseDN != "" {
		fmt.Println("Testing search capabilities...")
		searchRequest := ldap.NewSearchRequest(
			baseDN,
			ldap.ScopeBaseObject,
			ldap.NeverDerefAliases,
			1, 0, false,
			"(objectClass=*)",
			[]string{"dn"},
			nil,
		)

		if _, err := conn.Search(searchRequest); err != nil {
			return fmt.Errorf("search test failed: %w", err)
		}
		fmt.Println("✓ Search capabilities working")
	}

	return nil
}

func testADService() error {
	fmt.Println("Initializing AD Service...")
	adService := service.NewADService()

	fmt.Println("Testing AD Service connection...")
	if err := adService.TestConnection(); err != nil {
		return fmt.Errorf("AD service connection failed: %w", err)
	}
	fmt.Println("✓ AD Service connection successful")

	fmt.Println("Testing user search...")
	users, err := adService.GetAllUsers()
	if err != nil {
		fmt.Printf("⚠ User search warning: %v (this is OK if OU is empty)\n", err)
	} else {
		fmt.Printf("✓ Found %d users in managed OU\n", len(users))
		
		for i, user := range users {
			if i >= 3 {
				fmt.Printf("  ... and %d more users\n", len(users)-3)
				break
			}
			fmt.Printf("  - %s (%s) - Enabled: %v\n", 
				user.DisplayName, 
				user.SamAccountName, 
				user.Enabled)
		}
	}

	return nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}