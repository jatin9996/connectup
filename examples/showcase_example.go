package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Example usage of the showcase service

func main() {
	baseURL := "http://localhost:8080"

	// Step 1: Register a user
	fmt.Println("=== User Registration ===")
	userToken := registerUser(baseURL)
	if userToken == "" {
		fmt.Println("Failed to register user")
		return
	}
	fmt.Printf("User registered successfully. Token: %s\n", userToken[:20]+"...")

	// Step 2: Create a company profile
	fmt.Println("\n=== Company Profile Creation ===")
	companyID := createCompany(baseURL, userToken)
	if companyID == "" {
		fmt.Println("Failed to create company")
		return
	}
	fmt.Printf("Company created successfully. ID: %s\n", companyID)

	// Step 3: Create an investment
	fmt.Println("\n=== Investment Creation ===")
	investmentID := createInvestment(baseURL, userToken, companyID)
	if investmentID == "" {
		fmt.Println("Failed to create investment")
		return
	}
	fmt.Printf("Investment created successfully. ID: %s\n", investmentID)

	// Step 4: Search companies
	fmt.Println("\n=== Company Search ===")
	searchCompanies(baseURL, userToken)

	// Step 5: Get company details
	fmt.Println("\n=== Company Details ===")
	getCompanyDetails(baseURL, userToken, companyID)

	// Step 6: Track analytics event
	fmt.Println("\n=== Analytics Tracking ===")
	trackAnalyticsEvent(baseURL, userToken, companyID)

	// Step 7: Get user investments
	fmt.Println("\n=== User Investments ===")
	getUserInvestments(baseURL, userToken)

	fmt.Println("\n=== Showcase Service Demo Completed ===")
}

func registerUser(baseURL string) string {
	registerData := map[string]interface{}{
		"email":      "investor@example.com",
		"password":   "securepassword123",
		"first_name": "John",
		"last_name":  "Investor",
	}

	jsonData, _ := json.Marshal(registerData)
	resp, err := http.Post(baseURL+"/api/v1/auth/register", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error registering user: %v\n", err)
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Registration failed: %s\n", string(body))
		return ""
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if token, ok := result["access_token"].(string); ok {
		return token
	}
	return ""
}

func createCompany(baseURL, token string) string {
	companyData := map[string]interface{}{
		"name":           "TechCorp Innovations",
		"description":    "Leading technology company specializing in AI and machine learning solutions",
		"industry":       "Technology",
		"founded_year":   2020,
		"headquarters":   "San Francisco, CA",
		"website":        "https://techcorp-innovations.com",
		"logo_url":       "https://example.com/logo.png",
		"employee_count": 150,
		"revenue":        5000000,
		"funding_stage":  "Series A",
		"total_funding":  2000000,
		"valuation":      25000000,
		"is_public":      true,
	}

	jsonData, _ := json.Marshal(companyData)
	req, _ := http.NewRequest("POST", baseURL+"/api/v1/showcase/companies", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error creating company: %v\n", err)
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Company creation failed: %s\n", string(body))
		return ""
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if id, ok := result["id"].(string); ok {
		return id
	}
	return ""
}

func createInvestment(baseURL, token, companyID string) string {
	investmentData := map[string]interface{}{
		"company_id":      companyID,
		"amount":          500000,
		"currency":        "USD",
		"investment_type": "equity",
		"round":           "Series A",
		"date":            time.Now().Format("2006-01-02"),
		"status":          "completed",
		"notes":           "Strategic investment in AI technology",
	}

	jsonData, _ := json.Marshal(investmentData)
	req, _ := http.NewRequest("POST", baseURL+"/api/v1/showcase/investments", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error creating investment: %v\n", err)
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Investment creation failed: %s\n", string(body))
		return ""
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if id, ok := result["id"].(string); ok {
		return id
	}
	return ""
}

func searchCompanies(baseURL, token string) {
	req, _ := http.NewRequest("GET", baseURL+"/api/v1/showcase/companies?q=tech&industry=Technology&limit=5", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error searching companies: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Company search failed: %s\n", string(body))
		return
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if companies, ok := result["companies"].([]interface{}); ok {
		fmt.Printf("Found %d companies:\n", len(companies))
		for i, company := range companies {
			if companyMap, ok := company.(map[string]interface{}); ok {
				fmt.Printf("  %d. %s (%s)\n", i+1, companyMap["name"], companyMap["industry"])
			}
		}
	}
}

func getCompanyDetails(baseURL, token, companyID string) {
	req, _ := http.NewRequest("GET", baseURL+"/api/v1/showcase/companies/"+companyID, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error getting company details: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Get company details failed: %s\n", string(body))
		return
	}

	var company map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&company)

	fmt.Printf("Company: %s\n", company["name"])
	fmt.Printf("Industry: %s\n", company["industry"])
	fmt.Printf("Employees: %v\n", company["employee_count"])
	fmt.Printf("Revenue: $%.0f\n", company["revenue"])
	fmt.Printf("Valuation: $%.0f\n", company["valuation"])
}

func trackAnalyticsEvent(baseURL, token, companyID string) {
	eventData := map[string]interface{}{
		"event_type":    "company_viewed",
		"company_id":    companyID,
		"view_duration": 45,
		"source":        "search_results",
		"user_agent":    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
	}

	jsonData, _ := json.Marshal(eventData)
	req, _ := http.NewRequest("POST", baseURL+"/api/v1/showcase/analytics/events", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error tracking analytics event: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Println("Analytics event tracked successfully")
	} else {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Analytics tracking failed: %s\n", string(body))
	}
}

func getUserInvestments(baseURL, token string) {
	req, _ := http.NewRequest("GET", baseURL+"/api/v1/showcase/investments/my", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error getting user investments: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Get user investments failed: %s\n", string(body))
		return
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if investments, ok := result["investments"].([]interface{}); ok {
		fmt.Printf("User has %d investments:\n", len(investments))
		for i, investment := range investments {
			if invMap, ok := investment.(map[string]interface{}); ok {
				fmt.Printf("  %d. $%.0f %s (%s)\n", i+1, invMap["amount"], invMap["currency"], invMap["investment_type"])
			}
		}
	}
}

// WebSocket example (for reference)
func websocketExample() {
	fmt.Println(`
=== WebSocket Connection Example ===

JavaScript code to connect to WebSocket:

const ws = new WebSocket('ws://localhost:8080/ws');

ws.onopen = () => {
    console.log('Connected to WebSocket');
    
    // Send a chat message
    ws.send(JSON.stringify({
        type: 'chat_message',
        receiver_id: 'user-uuid',
        content: 'Hello from showcase service!'
    }));
};

ws.onmessage = (event) => {
    const data = JSON.parse(event.data);
    console.log('Received:', data);
};

ws.onclose = () => {
    console.log('WebSocket connection closed');
};
`)
}
