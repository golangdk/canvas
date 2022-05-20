package jobs

func (r *Runner) registerJobs() {
	SendNewsletterConfirmationEmail(r, r.emailer)
	SendNewsletterWelcomeEmail(r, r.emailer, r.blobStore)
	SendNewsletter(r, r.emailer, r.database)
}
