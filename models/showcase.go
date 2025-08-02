package models

import (
	"database/sql"
	"time"
)

// Company represents a company profile
type Company struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Industry      string    `json:"industry"`
	FoundedYear   int       `json:"founded_year"`
	Headquarters  string    `json:"headquarters"`
	Website       string    `json:"website"`
	LogoURL       string    `json:"logo_url"`
	EmployeeCount int       `json:"employee_count"`
	Revenue       float64   `json:"revenue"`
	FundingStage  string    `json:"funding_stage"`
	TotalFunding  float64   `json:"total_funding"`
	Valuation     float64   `json:"valuation"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	CreatedBy     string    `json:"created_by"`
	IsPublic      bool      `json:"is_public"`
}

// Investment represents an investment record
type Investment struct {
	ID             string    `json:"id"`
	CompanyID      string    `json:"company_id"`
	InvestorID     string    `json:"investor_id"`
	Amount         float64   `json:"amount"`
	Currency       string    `json:"currency"`
	InvestmentType string    `json:"investment_type"` // equity, debt, convertible_note, etc.
	Round          string    `json:"round"`           // seed, series_a, series_b, etc.
	Date           time.Time `json:"date"`
	Status         string    `json:"status"` // pending, completed, cancelled
	Notes          string    `json:"notes"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// AnalyticsEvent represents analytics tracking events
type AnalyticsEvent struct {
	ID        string                 `json:"id"`
	UserID    string                 `json:"user_id"`
	EventType string                 `json:"event_type"`
	EventData map[string]interface{} `json:"event_data"`
	Timestamp time.Time              `json:"timestamp"`
	IPAddress string                 `json:"ip_address"`
	UserAgent string                 `json:"user_agent"`
	SessionID string                 `json:"session_id"`
}

// Message represents a chat message
type Message struct {
	ID          string    `json:"id"`
	SenderID    string    `json:"sender_id"`
	ReceiverID  string    `json:"receiver_id"`
	Content     string    `json:"content"`
	MessageType string    `json:"message_type"` // text, image, file, etc.
	IsRead      bool      `json:"is_read"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateShowcaseTables creates the showcase-related tables
func CreateShowcaseTables() error {
	queries := []string{
		// Companies table
		`CREATE TABLE IF NOT EXISTS companies (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			description TEXT,
			industry VARCHAR(100),
			founded_year INTEGER,
			headquarters VARCHAR(255),
			website VARCHAR(255),
			logo_url VARCHAR(500),
			employee_count INTEGER,
			revenue DECIMAL(15,2),
			funding_stage VARCHAR(50),
			total_funding DECIMAL(15,2),
			valuation DECIMAL(15,2),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			created_by UUID REFERENCES users(id),
			is_public BOOLEAN DEFAULT false
		);`,

		// Investments table
		`CREATE TABLE IF NOT EXISTS investments (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			company_id UUID REFERENCES companies(id) ON DELETE CASCADE,
			investor_id UUID REFERENCES users(id) ON DELETE CASCADE,
			amount DECIMAL(15,2) NOT NULL,
			currency VARCHAR(3) DEFAULT 'USD',
			investment_type VARCHAR(50) NOT NULL,
			round VARCHAR(50),
			date DATE NOT NULL,
			status VARCHAR(20) DEFAULT 'pending',
			notes TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`,

		// Analytics events table
		`CREATE TABLE IF NOT EXISTS analytics_events (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID REFERENCES users(id) ON DELETE CASCADE,
			event_type VARCHAR(100) NOT NULL,
			event_data JSONB,
			timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			ip_address INET,
			user_agent TEXT,
			session_id VARCHAR(255)
		);`,

		// Messages table
		`CREATE TABLE IF NOT EXISTS messages (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			sender_id UUID REFERENCES users(id) ON DELETE CASCADE,
			receiver_id UUID REFERENCES users(id) ON DELETE CASCADE,
			content TEXT NOT NULL,
			message_type VARCHAR(20) DEFAULT 'text',
			is_read BOOLEAN DEFAULT false,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`,

		// Sessions table for WebSocket connections
		`CREATE TABLE IF NOT EXISTS sessions (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID REFERENCES users(id) ON DELETE CASCADE,
			session_token VARCHAR(255) UNIQUE NOT NULL,
			expires_at TIMESTAMP NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			is_active BOOLEAN DEFAULT true
		);`,

		// Create indexes
		`CREATE INDEX IF NOT EXISTS idx_companies_industry ON companies(industry);`,
		`CREATE INDEX IF NOT EXISTS idx_companies_funding_stage ON companies(funding_stage);`,
		`CREATE INDEX IF NOT EXISTS idx_companies_is_public ON companies(is_public);`,
		`CREATE INDEX IF NOT EXISTS idx_investments_company_id ON investments(company_id);`,
		`CREATE INDEX IF NOT EXISTS idx_investments_investor_id ON investments(investor_id);`,
		`CREATE INDEX IF NOT EXISTS idx_investments_date ON investments(date);`,
		`CREATE INDEX IF NOT EXISTS idx_analytics_events_user_id ON analytics_events(user_id);`,
		`CREATE INDEX IF NOT EXISTS idx_analytics_events_timestamp ON analytics_events(timestamp);`,
		`CREATE INDEX IF NOT EXISTS idx_messages_sender_id ON messages(sender_id);`,
		`CREATE INDEX IF NOT EXISTS idx_messages_receiver_id ON messages(receiver_id);`,
		`CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at);`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_token ON sessions(session_token);`,

		// Full-text search indexes
		`CREATE INDEX IF NOT EXISTS idx_companies_name_fts ON companies USING GIN(to_tsvector('english', name));`,
		`CREATE INDEX IF NOT EXISTS idx_companies_description_fts ON companies USING GIN(to_tsvector('english', description));`,
	}

	for _, query := range queries {
		if _, err := DB.Exec(query); err != nil {
			return err
		}
	}

	return nil
}

// GetCompanyByID retrieves a company by ID
func GetCompanyByID(id string) (*Company, error) {
	query := `
		SELECT id, name, description, industry, founded_year, headquarters, 
		       website, logo_url, employee_count, revenue, funding_stage, 
		       total_funding, valuation, created_at, updated_at, created_by, is_public
		FROM companies WHERE id = $1
	`

	var company Company
	err := DB.QueryRow(query, id).Scan(
		&company.ID, &company.Name, &company.Description, &company.Industry,
		&company.FoundedYear, &company.Headquarters, &company.Website, &company.LogoURL,
		&company.EmployeeCount, &company.Revenue, &company.FundingStage,
		&company.TotalFunding, &company.Valuation, &company.CreatedAt,
		&company.UpdatedAt, &company.CreatedBy, &company.IsPublic,
	)

	if err != nil {
		return nil, err
	}

	return &company, nil
}

// CreateCompany creates a new company
func CreateCompany(company *Company) error {
	query := `
		INSERT INTO companies (name, description, industry, founded_year, headquarters,
		                     website, logo_url, employee_count, revenue, funding_stage,
		                     total_funding, valuation, created_by, is_public)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, created_at, updated_at
	`

	return DB.QueryRow(query,
		company.Name, company.Description, company.Industry, company.FoundedYear,
		company.Headquarters, company.Website, company.LogoURL, company.EmployeeCount,
		company.Revenue, company.FundingStage, company.TotalFunding, company.Valuation,
		company.CreatedBy, company.IsPublic,
	).Scan(&company.ID, &company.CreatedAt, &company.UpdatedAt)
}

// UpdateCompany updates an existing company
func UpdateCompany(company *Company) error {
	query := `
		UPDATE companies SET 
			name = $1, description = $2, industry = $3, founded_year = $4,
			headquarters = $5, website = $6, logo_url = $7, employee_count = $8,
			revenue = $9, funding_stage = $10, total_funding = $11, valuation = $12,
			is_public = $13, updated_at = CURRENT_TIMESTAMP
		WHERE id = $14
	`

	result, err := DB.Exec(query,
		company.Name, company.Description, company.Industry, company.FoundedYear,
		company.Headquarters, company.Website, company.LogoURL, company.EmployeeCount,
		company.Revenue, company.FundingStage, company.TotalFunding, company.Valuation,
		company.IsPublic, company.ID,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// SearchCompanies searches companies with filters
func SearchCompanies(query string, industry string, fundingStage string, limit, offset int) ([]*Company, error) {
	baseQuery := `
		SELECT id, name, description, industry, founded_year, headquarters,
		       website, logo_url, employee_count, revenue, funding_stage,
		       total_funding, valuation, created_at, updated_at, created_by, is_public
		FROM companies
		WHERE is_public = true
	`

	var conditions []string
	var args []interface{}
	argIndex := 1

	if query != "" {
		conditions = append(conditions, `(name ILIKE $`+string(rune(argIndex+48))+` OR description ILIKE $`+string(rune(argIndex+48))+`)`)
		args = append(args, "%"+query+"%")
		argIndex++
	}

	if industry != "" {
		conditions = append(conditions, `industry = $`+string(rune(argIndex+48)))
		args = append(args, industry)
		argIndex++
	}

	if fundingStage != "" {
		conditions = append(conditions, `funding_stage = $`+string(rune(argIndex+48)))
		args = append(args, fundingStage)
		argIndex++
	}

	if len(conditions) > 0 {
		baseQuery += " AND " + conditions[0]
		for i := 1; i < len(conditions); i++ {
			baseQuery += " AND " + conditions[i]
		}
	}

	baseQuery += ` ORDER BY created_at DESC LIMIT $` + string(rune(argIndex+48)) + ` OFFSET $` + string(rune(argIndex+49))
	args = append(args, limit, offset)

	rows, err := DB.Query(baseQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var companies []*Company
	for rows.Next() {
		var company Company
		err := rows.Scan(
			&company.ID, &company.Name, &company.Description, &company.Industry,
			&company.FoundedYear, &company.Headquarters, &company.Website, &company.LogoURL,
			&company.EmployeeCount, &company.Revenue, &company.FundingStage,
			&company.TotalFunding, &company.Valuation, &company.CreatedAt,
			&company.UpdatedAt, &company.CreatedBy, &company.IsPublic,
		)
		if err != nil {
			return nil, err
		}
		companies = append(companies, &company)
	}

	return companies, nil
}
