package models

import "time"

// User Registration Request
type SignUpRequest struct {
	Name            string `json:"name"`
	Email           string `json:"email"`
	Password        string `json:"password"`
	UserType        string `json:"user_type"` // Admin or Applicant
	ProfileHeadline string `json:"profile_headline"`
	Address         string `json:"address"`
}

// User Login Request
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Create Job Request
type CreateJobRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	CompanyName string `json:"company_name"`
}

// Response for a single job with its applicants
type JobWithApplicantsResponse struct {
	ID                uint               `json:"id"`
	Title             string             `json:"title"`
	Description       string             `json:"description"`
	CompanyName       string             `json:"company_name"`
	PostedOn          time.Time          `json:"posted_on"`
	TotalApplications int                `json:"total_applications"`
	Applicants        []ApplicantDetails `json:"applicants"`
}

// Simplified applicant details for responses
type ApplicantDetails struct {
	ID              uint   `json:"id"`
	Name            string `json:"name"`
	Email           string `json:"email"`
	ProfileHeadline string `json:"profile_headline"`
}
