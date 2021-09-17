package jobs

func (r *Runner) setupJobs() {
	SendNewsletterConfirmationEmail(r, r.emailer)
}
