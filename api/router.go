package api

import (
	"net/http"
	"recruitment-system/auth"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Public routes
	r.Post("/signup", SignUp)
	r.Post("/login", Login)

	// Authenticated routes
	r.Group(func(r chi.Router) {
		r.Use(auth.AuthMiddleware)

		// Routes accessible to all authenticated users
		r.Get("/jobs", GetJobs)

		// Applicant-only routes
		r.Group(func(r chi.Router) {
			r.Use(auth.ApplicantMiddleware)
			r.Post("/uploadResume", UploadResume)
			r.Get("/jobs/apply", ApplyForJob) // Using GET as per spec, though POST might be more RESTful
		})

		// Admin-only routes
		r.Group(func(r chi.Router) {
			r.Use(auth.AdminMiddleware)
			r.Post("/admin/job", CreateJob)
			r.Get("/admin/job/{job_id}", GetJobWithApplicants)
			r.Get("/admin/applicants", GetAllApplicants)
			r.Get("/admin/applicant/{applicant_id}", GetApplicantProfile)
		})
	})

	return r
}
