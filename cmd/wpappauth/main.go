package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"terraform-provider-wordpress/internal/wpappauth"
)

func main() {
	var service wpappauth.Service
	var outputJSON bool

	flag.StringVar(&service.BaseURL, "base-url", getenv("WORDPRESS_HOST", "http://localhost:8888/wp-json/wp/v2"), "WordPress REST API base URL")
	flag.StringVar(&service.Username, "username", getenv("WORDPRESS_USERNAME", "admin"), "WordPress username")
	flag.StringVar(&service.Password, "password", getenv("WORDPRESS_PASSWORD", ""), "WordPress password")
	flag.StringVar(&service.ApplicationName, "name", getenv("WORDPRESS_APPLICATION_NAME", "terraform-provider-wordpress"), "application password name")
	flag.BoolVar(&outputJSON, "json", false, "print the full response as JSON")
	flag.Parse()

	result, err := service.CreateApplicationPassword(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	if outputJSON {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(result); err != nil {
			log.Fatal(err)
		}
		return
	}

	fmt.Println(strings.TrimSpace(result.Password))
}

func getenv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
