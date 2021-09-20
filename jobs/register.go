package jobs

func (r *Runner) registerJobs() {
	SendNewsletterConfirmationEmail(r, r.emailer)
}
