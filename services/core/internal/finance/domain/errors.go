package domain

// ErrorDef describes one finance problem code and the HTTP status the
// contract answers it with.
type ErrorDef struct {
	Code   string
	Status int
}

// Problem codes of the hryvnia equivalent (PRD §5.1).
const (
	CodeEquivalentNotApplicable = "finance.equivalent_not_applicable"
	CodeRateNotPublished        = "finance.rate_not_published"
	CodeRateUnavailable         = "finance.rate_unavailable"
)

// DomainError is a rule violation identified only by its problem code.
type DomainError struct {
	Code string
}

func (e *DomainError) Error() string { return e.Code }

// ErrorRegistry lists every finance.* problem code exactly once.
var ErrorRegistry = []ErrorDef{
	{Code: CodeEquivalentNotApplicable, Status: 409},
	{Code: CodeRateNotPublished, Status: 409},
	{Code: CodeRateUnavailable, Status: 503},
}
