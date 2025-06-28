package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"recruitment-system/auth"
	"recruitment-system/db"
	"recruitment-system/models"
	"recruitment-system/services"
	"recruitment-system/utils"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

// POST /signup
func SignUp(w http.ResponseWriter, r *http.Request) {
	var req models.SignUpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if req.UserType != "Admin" && req.UserType != "Applicant" {
		utils.RespondWithError(w, http.StatusBadRequest, "UserType must be 'Admin' or 'Applicant'")
		return
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	user := models.User{
		Name:            req.Name,
		Email:           req.Email,
		UserType:        req.UserType,
		PasswordHash:    hashedPassword,
		ProfileHeadline: req.ProfileHeadline,
		Address:         req.Address,
	}

	// Create user and an empty associated profile
	tx := db.DB.Begin()
	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}
	profile := models.Profile{UserID: user.ID}
	if err := tx.Create(&profile).Error; err != nil {
		tx.Rollback()
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create user profile")
		return
	}
	tx.Commit()

	utils.RespondWithJSON(w, http.StatusCreated, map[string]string{"message": "User created successfully"})
}

// POST /login
func Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	var user models.User
	if err := db.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
		utils.RespondWithError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	token, err := auth.GenerateJWT(&user)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"token": token})
}

// POST /uploadResume
func UploadResume(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(uint)

	r.ParseMultipartForm(10 << 20) // 10 MB
	file, handler, err := r.FormFile("resume")
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Could not get uploaded file")
		return
	}
	defer file.Close()

	// Validate file type
	ext := filepath.Ext(handler.Filename)
	if ext != ".pdf" && ext != ".docx" {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid file type. Only PDF and DOCX are allowed.")
		return
	}

	// Create a temporary file
	os.MkdirAll("./uploads", os.ModePerm)
	filePath := fmt.Sprintf("./uploads/resume-%d-%d%s", userID, time.Now().Unix(), ext)
	dst, err := os.Create(filePath)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Could not save file")
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Could not save file")
		return
	}

	// Process with third-party API
	parsedData, err := services.ParseResume(filePath)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to parse resume: %v", err))
		return
	}

	// Update profile in DB
	var profile models.Profile
	if err := db.DB.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Profile not found")
		return
	}

	profile.ResumeFileAddress = filePath
	profile.Skills = strings.Join(parsedData.Skills, ", ")
	profile.Education = services.ConvertToJSONString(parsedData.Education)
	profile.Experience = services.ConvertToJSONString(parsedData.Experience)
	profile.Name = parsedData.Name
	profile.Email = parsedData.Email
	profile.Phone = parsedData.Phone

	if err := db.DB.Save(&profile).Error; err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update profile")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Resume uploaded and processed successfully"})
}

// POST /admin/job
func CreateJob(w http.ResponseWriter, r *http.Request) {
	adminID := r.Context().Value(auth.UserIDKey).(uint)

	var req models.CreateJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	job := models.Job{
		Title:       req.Title,
		Description: req.Description,
		CompanyName: req.CompanyName,
		PostedByID:  adminID,
		PostedOn:    time.Now(),
	}

	if err := db.DB.Create(&job).Error; err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create job")
		return
	}

	utils.RespondWithJSON(w, http.StatusCreated, job)
}

// GET /admin/job/{job_id}
func GetJobWithApplicants(w http.ResponseWriter, r *http.Request) {
	jobIDStr := chi.URLParam(r, "job_id")
	jobID, err := strconv.Atoi(jobIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid job ID")
		return
	}

	var job models.Job
	if err := db.DB.Preload("Applicants").First(&job, jobID).Error; err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Job not found")
		return
	}

	var applicantDetails []models.ApplicantDetails
	for _, applicant := range job.Applicants {
		applicantDetails = append(applicantDetails, models.ApplicantDetails{
			ID:              applicant.ID,
			Name:            applicant.Name,
			Email:           applicant.Email,
			ProfileHeadline: applicant.ProfileHeadline,
		})
	}

	response := models.JobWithApplicantsResponse{
		ID:                job.ID,
		Title:             job.Title,
		Description:       job.Description,
		CompanyName:       job.CompanyName,
		PostedOn:          job.PostedOn,
		TotalApplications: len(job.Applicants),
		Applicants:        applicantDetails,
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

// GET /admin/applicants
func GetAllApplicants(w http.ResponseWriter, r *http.Request) {
	var users []models.User
	if err := db.DB.Where("user_type = ?", "Applicant").Find(&users).Error; err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Could not fetch applicants")
		return
	}
	utils.RespondWithJSON(w, http.StatusOK, users)
}

// GET /admin/applicant/{applicant_id}
func GetApplicantProfile(w http.ResponseWriter, r *http.Request) {
	applicantIDStr := chi.URLParam(r, "applicant_id")
	applicantID, err := strconv.Atoi(applicantIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid applicant ID")
		return
	}

	var profile models.Profile
	if err := db.DB.Where("user_id = ?", applicantID).First(&profile).Error; err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Applicant profile not found")
		return
	}
	utils.RespondWithJSON(w, http.StatusOK, profile)
}

// GET /jobs
func GetJobs(w http.ResponseWriter, r *http.Request) {
	var jobs []models.Job
	if err := db.DB.Find(&jobs).Error; err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Could not fetch jobs")
		return
	}
	utils.RespondWithJSON(w, http.StatusOK, jobs)
}

// GET /jobs/apply?job_id={job_id}
func ApplyForJob(w http.ResponseWriter, r *http.Request) {
	applicantID := r.Context().Value(auth.UserIDKey).(uint)
	jobIDStr := r.URL.Query().Get("job_id")
	jobID, err := strconv.Atoi(jobIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid job ID")
		return
	}

	// Check if job exists
	var job models.Job
	if err := db.DB.First(&job, jobID).Error; err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Job not found")
		return
	}

	// Check for duplicate application
	var existingApplication models.JobApplication
	if db.DB.Where("user_id = ? AND job_id = ?", applicantID, jobID).First(&existingApplication).Error == nil {
		utils.RespondWithError(w, http.StatusConflict, "You have already applied for this job")
		return
	}

	application := models.JobApplication{
		UserID:          applicantID,
		JobID:           uint(jobID),
		ApplicationDate: time.Now(),
	}

	if err := db.DB.Create(&application).Error; err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to apply for job")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Successfully applied for the job"})
}
