package models

import (
	"time"

	"github.com/google/uuid"
)

// PayrollPeriod represents a payroll period
type PayrollPeriod struct {
	ID          uuid.UUID  `json:"id"`
	StartDate   time.Time  `json:"start_date"`
	EndDate     time.Time  `json:"end_date"`
	PayDate     time.Time  `json:"pay_date"`
	Status      string     `json:"status"`
	ProcessedBy *uuid.UUID `json:"processed_by,omitempty"`
	ProcessedAt *time.Time `json:"processed_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// PayStub represents an employee's pay stub
type PayStub struct {
	ID                  uuid.UUID `json:"id"`
	EmployeeID          uuid.UUID `json:"employee_id"`
	PayrollPeriodID     uuid.UUID `json:"payroll_period_id"`
	GrossPay            float64   `json:"gross_pay"`
	FederalTax          float64   `json:"federal_tax"`
	StateTax            float64   `json:"state_tax"`
	SocialSecurity      float64   `json:"social_security"`
	Medicare            float64   `json:"medicare"`
	OtherDeductions     float64   `json:"other_deductions"`
	NetPay              float64   `json:"net_pay"`
	HoursWorked         *float64  `json:"hours_worked,omitempty"`
	OvertimeHours       *float64  `json:"overtime_hours,omitempty"`
	HourlyRate          *float64  `json:"hourly_rate,omitempty"`
	BenefitsDeductions  float64   `json:"benefits_deductions"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// EmployeeCompensation represents compensation details for an employee
type EmployeeCompensation struct {
	ID                   uuid.UUID  `json:"id" db:"id"`
	EmployeeID           uuid.UUID  `json:"employee_id" db:"employee_id"`
	EmploymentType       string     `json:"employment_type" db:"employment_type"` // W2 or 1099
	PayType              string     `json:"pay_type" db:"pay_type"`               // hourly, salary, commission
	HourlyRate           *float64   `json:"hourly_rate,omitempty" db:"hourly_rate"`
	AnnualSalary         *float64   `json:"annual_salary,omitempty" db:"annual_salary"`
	PayFrequency         string     `json:"pay_frequency" db:"pay_frequency"` // weekly, biweekly, semimonthly, monthly
	EffectiveDate        time.Time  `json:"effective_date" db:"effective_date"`
	EndDate              *time.Time `json:"end_date,omitempty" db:"end_date"`
	OvertimeEligible     bool       `json:"overtime_eligible" db:"overtime_eligible"`
	StandardHoursPerWeek float64    `json:"standard_hours_per_week" db:"standard_hours_per_week"`
	CreatedAt            time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at" db:"updated_at"`
}

// W2TaxWithholding represents tax withholding for W2 employees
type W2TaxWithholding struct {
	ID                    uuid.UUID `json:"id" db:"id"`
	EmployeeID            uuid.UUID `json:"employee_id" db:"employee_id"`
	FilingStatus          string    `json:"filing_status" db:"filing_status"` // single, married, head_of_household
	FederalAllowances     int       `json:"federal_allowances" db:"federal_allowances"`
	StateAllowances       int       `json:"state_allowances" db:"state_allowances"`
	AdditionalWithholding float64   `json:"additional_withholding" db:"additional_withholding"`
	ExemptFederal         bool      `json:"exempt_federal" db:"exempt_federal"`
	ExemptState           bool      `json:"exempt_state" db:"exempt_state"`
	ExemptFICA            bool      `json:"exempt_fica" db:"exempt_fica"`
	CreatedAt             time.Time `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time `json:"updated_at" db:"updated_at"`
}

// PayStubEarning represents individual earnings on a pay stub
type PayStubEarning struct {
	ID          uuid.UUID `json:"id" db:"id"`
	PayStubID   uuid.UUID `json:"pay_stub_id" db:"pay_stub_id"`
	EarningType string    `json:"earning_type" db:"earning_type"` // regular, overtime, bonus, commission, etc.
	Description string    `json:"description" db:"description"`
	Hours       *float64  `json:"hours,omitempty" db:"hours"`
	Rate        *float64  `json:"rate,omitempty" db:"rate"`
	Amount      float64   `json:"amount" db:"amount"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// PayStubDeduction represents individual deductions on a pay stub
type PayStubDeduction struct {
	ID             uuid.UUID `json:"id" db:"id"`
	PayStubID      uuid.UUID `json:"pay_stub_id" db:"pay_stub_id"`
	DeductionType  string    `json:"deduction_type" db:"deduction_type"` // 401k, health, dental, etc.
	Description    string    `json:"description" db:"description"`
	Amount         float64   `json:"amount" db:"amount"`
	EmployerMatch  *float64  `json:"employer_match,omitempty" db:"employer_match"`
	PreTax         bool      `json:"pre_tax" db:"pre_tax"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

// PayStubTax represents individual tax withholdings on a pay stub
type PayStubTax struct {
	ID          uuid.UUID `json:"id" db:"id"`
	PayStubID   uuid.UUID `json:"pay_stub_id" db:"pay_stub_id"`
	TaxType     string    `json:"tax_type" db:"tax_type"` // federal, state, local, fica_ss, fica_medicare
	Description string    `json:"description" db:"description"`
	Amount      float64   `json:"amount" db:"amount"`
	TaxableWage float64   `json:"taxable_wage" db:"taxable_wage"`
	TaxRate     *float64  `json:"tax_rate,omitempty" db:"tax_rate"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// Form1099 represents a 1099 form for contractors
type Form1099 struct {
	ID                 uuid.UUID  `json:"id" db:"id"`
	EmployeeID         uuid.UUID  `json:"employee_id" db:"employee_id"`
	TaxYear            int        `json:"tax_year" db:"tax_year"`
	TotalPayments      float64    `json:"total_payments" db:"total_payments"`
	FederalTaxWithheld float64    `json:"federal_tax_withheld" db:"federal_tax_withheld"`
	StateTaxWithheld   float64    `json:"state_tax_withheld" db:"state_tax_withheld"`
	Status             string     `json:"status" db:"status"` // draft, filed, corrected
	FiledDate          *time.Time `json:"filed_date,omitempty" db:"filed_date"`
	CorrectedFormID    *uuid.UUID `json:"corrected_form_id,omitempty" db:"corrected_form_id"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at" db:"updated_at"`
}

// DTOs for API requests/responses

type CreateCompensationRequest struct {
	EmployeeID           uuid.UUID  `json:"employee_id"`
	EmploymentType       string     `json:"employment_type"`
	PayType              string     `json:"pay_type"`
	HourlyRate           *float64   `json:"hourly_rate,omitempty"`
	AnnualSalary         *float64   `json:"annual_salary,omitempty"`
	PayFrequency         string     `json:"pay_frequency"`
	EffectiveDate        time.Time  `json:"effective_date"`
	OvertimeEligible     bool       `json:"overtime_eligible"`
	StandardHoursPerWeek float64    `json:"standard_hours_per_week"`
}

type UpdateTaxWithholdingRequest struct {
	FilingStatus          string  `json:"filing_status"`
	FederalAllowances     int     `json:"federal_allowances"`
	StateAllowances       int     `json:"state_allowances"`
	AdditionalWithholding float64 `json:"additional_withholding"`
	ExemptFederal         bool    `json:"exempt_federal"`
	ExemptState           bool    `json:"exempt_state"`
	ExemptFICA            bool    `json:"exempt_fica"`
}

type CreatePayrollPeriodRequest struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	PayDate   time.Time `json:"pay_date"`
}

type ProcessPayrollRequest struct {
	PayrollPeriodID uuid.UUID `json:"payroll_period_id"`
	DryRun          bool      `json:"dry_run"` // Preview without creating records
}

type PayrollSummary struct {
	Period          *PayrollPeriod `json:"period"`
	TotalEmployees  int            `json:"total_employees"`
	W2Employees     int            `json:"w2_employees"`
	Contractors1099 int            `json:"contractors_1099"`
	TotalGrossPay   float64        `json:"total_gross_pay"`
	TotalTaxes      float64        `json:"total_taxes"`
	TotalDeductions float64        `json:"total_deductions"`
	TotalNetPay     float64        `json:"total_net_pay"`
	Status          string         `json:"status"`
}

type EmployeePayrollInfo struct {
	Employee       *Employee             `json:"employee"`
	Compensation   *EmployeeCompensation `json:"compensation"`
	TaxWithholding *W2TaxWithholding     `json:"tax_withholding,omitempty"`
	RecentPayStubs []*PayStub            `json:"recent_pay_stubs,omitempty"`
	YTDEarnings    float64               `json:"ytd_earnings"`
	YTDTaxes       float64               `json:"ytd_taxes"`
}

type PayStubDetail struct {
	PayStub        *PayStub            `json:"pay_stub"`
	Employee       *Employee           `json:"employee"`
	PayrollPeriod  *PayrollPeriod      `json:"payroll_period"`
	Earnings       []PayStubEarning    `json:"earnings"`
	Deductions     []PayStubDeduction  `json:"deductions"`
	Taxes          []PayStubTax        `json:"taxes"`
	YTDGrossPay    float64             `json:"ytd_gross_pay"`
	YTDNetPay      float64             `json:"ytd_net_pay"`
	YTDFederalTax  float64             `json:"ytd_federal_tax"`
	YTDStateTax    float64             `json:"ytd_state_tax"`
}