package v1

// CompanySummary is only returned in membership-protected detail sections.
// Lists do not include this summary, and the full profile requires its own
// server-side membership check.
type CompanySummary struct {
	ID          uint64 `json:"id"`
	CompanyName string `json:"company_name"`
}
