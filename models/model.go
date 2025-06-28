package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name            string
	Email           string `gorm:"unique;not null"`
	Address         string
	UserType        string `gorm:"not null"` // "Admin" or "Applicant"
	PasswordHash    string `gorm:"not null"`
	ProfileHeadline string
	Profile         Profile `gorm:"foreignKey:UserID"`
}

type Profile struct {
	gorm.Model
	UserID            uint `gorm:"unique;not null"`
	ResumeFileAddress string
	Skills            string
	Education         string
	Experience        string
	Name              string
	Email             string
	Phone             string
}

type Job struct {
	gorm.Model
	Title       string
	Description string
	PostedOn    time.Time
	CompanyName string
	PostedByID  uint
	PostedBy    User   `gorm:"foreignKey:PostedByID"`
	Applicants  []User `gorm:"many2many:job_applications;"`
}

// Join table for many-to-many relationship between User (Applicants) and Job
type JobApplication struct {
	gorm.Model
	UserID          uint
	JobID           uint
	ApplicationDate time.Time
}
