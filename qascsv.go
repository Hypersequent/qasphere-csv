// Package qascsv provides APIs to generate CSV files that can be used to import
// test cases in a QA Sphere project.
package qascsv

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

// staticColumns will always be present in the CSV file
// but there can be additional columns for steps and custom fields.
var staticColumns = []string{
	"Folder", "Type", "Name", "Legacy ID", "Draft", "Priority", "Tags", "Requirements",
	"Links", "Files", "Preconditions", "Parameter Values", "Template Suffix Params",
}

// Priority represents the priority of a test case in QA Sphere.
type Priority string

// The priorities available in QA Sphere.
const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
)

type TestCaseType string

const (
	TestCaseTypeStandalone TestCaseType = "standalone"
	TestCaseTypeTemplate   TestCaseType = "template"
)

// Requirement represent important requirements and reference document
// associated with a test case. At least one of title/url is required.
type Requirement struct {
	Title string `validate:"required_without=URL,max=255"`
	URL   string `validate:"required_without=Title,omitempty,http_url,max=255"`
}

// Link represents a URL.
type Link struct {
	Title string `validate:"required,max=255"`
	URL   string `validate:"required,http_url,max=255"`
}

// File represents an attachment or file associated with a test case.
// These files need to be uploaded to the QA Sphere project via the API
// See API documentation for more details.
//
// https://docs.qasphere.com/api/upload_file
type File struct {
	Name     string `validate:"required" json:"fileName"`
	ID       string `validate:"required" json:"id"`
	URL      string `validate:"required" json:"url"`
	MimeType string `validate:"required" json:"mimeType"`
	Size     int64  `validate:"required" json:"size"`
}

// Step represents a single action to perform in a test case.
type Step struct {
	// The action to perform. Markdown is supported. (optional)
	Action string
	// The expected result of the action. Markdown is supported. (optional)
	Expected string
}

// ParameterValue represents parameter values that you provide for template test cases.
// Template test cases are test cases where the body of the test case can contain some placeholders
// of the form ${parameter_name}. Then the users need to provide the values for these parameters
// in the form of a map. These are used to generate a filled test case which replaced the placeholders.
type ParameterValue struct {
	Priority *Priority         `json:"priority,omitempty" validate:"oneof=low medium high"`
	Values   map[string]string `json:"values" validate:"required,dive,keys,max=255,endkeys"`
}

type CustomFieldType string

const (
	CustomFieldTypeText     CustomFieldType = "text"
	CustomFieldTypeDropdown CustomFieldType = "dropdown"
)

type CustomField struct {
	SystemName string          `validate:"required,max=64"`
	Type       CustomFieldType `validate:"required,oneof=text dropdown"`
}

type CustomFieldValue struct {
	Value     string `json:"value" validate:"max=255"`
	IsDefault bool   `json:"isDefault" validate:"omitempty"`
}

// TestCase represents a test case in QA Sphere.
type TestCase struct {
	// The title of the test case. (required)
	Title string `validate:"required,max=511"`
	// The type of the test case. (optional)
	// If not specified, it defaults to "standalone".
	Type TestCaseType `validate:"omitempty,oneof=standalone template"`
	// In case of migrating from another test management system, the
	// test case ID in the existing test management system. This is only
	// for reference. (optional)
	LegacyID string `validate:"max=255"`
	// The complete folder path to the test case. (required)
	Folder []string `validate:"min=1,dive,required,max=255,excludesall=/"`
	// The priority of the test case. (required)
	Priority Priority `validate:"required,oneof=low medium high"`
	// The tags to assign to the test cases. This can be used to group,
	// filter or organise related test cases and also helps in creating
	// test runs. (optional)
	Tags []string `validate:"dive,required,max=255"`
	// The preconditions (or description) for the test case. Markdown is
	// supported. (optional)
	Preconditions string
	// The sequence of (ordered) actions to be performed while executing
	// the test case. (optional)
	Steps []Step
	// Primary requirement or reference document associated with the
	// test case. (optional)
	Requirements []Requirement `validate:"dive"`
	// Any other files relevant to the test case. (optional)
	Files []File `validate:"dive"`
	// Any other links relevant to the test case. (optional)
	Links []Link `validate:"dive"`
	// Whether the test case is still work in progress and not in its
	// final state. The test case should later be updated as and then
	// published. (optional)
	Draft bool
	// The parameter values to be used for the test case. (optional)
	// This is used for template test cases where the body of the test case
	// can contain some placeholders of the form ${parameter_name}.
	// For each ParameterValue provided in this array, we generate a distinct filled test case
	// See ParameterValue for more details.
	ParameterValues []ParameterValue `validate:"dive"`
	// The filled template suffix params to be used for template test cases.
	// For easy identification, we add a suffix to the filled test case title
	// For example, for a template with title "Template_title" and you provide SuffixParams as param1, param2
	// The generated filled test case will have title "Template_title (param1=val1, param2=val2)".
	FilledTCaseTitleSuffixParams []string `validate:"dive,max=255"`
	// The custom fields to be used for the test case. (optional)
	CustomFields map[string]CustomFieldValue `validate:"dive,keys,max=64,endkeys,required"`
}

// QASphereCSV provides APIs to generate CSV that can be used to import
// test cases in a project on QA Sphere.
type QASphereCSV struct {
	folderTCaseMap map[string][]TestCase
	validate       *validator.Validate
	customFields   []CustomField

	numTCases int
	maxSteps  int
}

func NewQASphereCSV() *QASphereCSV {
	return &QASphereCSV{
		folderTCaseMap: make(map[string][]TestCase),
		validate:       validator.New(),
	}
}

// AddCustomField adds a custom field to the QASphereCSV.
// The custom fields need to pre-declared by using AddCustomField or AddCustomFields
func (q *QASphereCSV) AddCustomField(cf CustomField) error {
	if err := q.validate.Struct(cf); err != nil {
		return errors.Wrap(err, "custom field validation")
	}

	// Check for duplicate custom field SystemName
	for _, existingCF := range q.customFields {
		if existingCF.SystemName == cf.SystemName {
			return errors.Errorf("custom field with SystemName %q already exists", cf.SystemName)
		}
	}

	q.customFields = append(q.customFields, cf)
	return nil
}

// AddCustomFields adds multiple custom fields to the QASphereCSV.
func (q *QASphereCSV) AddCustomFields(cfs []CustomField) error {
	var err error
	for _, cf := range cfs {
		if retErr := q.AddCustomField(cf); retErr != nil {
			err = multierror.Append(err, retErr)
		}
	}
	return err
}
func (q *QASphereCSV) AddTestCase(tc TestCase) error {
	if tc.Type == TestCaseType("") {
		tc.Type = TestCaseTypeStandalone
	}

	if err := q.validateTestCase(tc); err != nil {
		return errors.Wrap(err, "test case validation")
	}

	q.addTCase(tc)
	return nil
}

func (q *QASphereCSV) AddTestCases(tcs []TestCase) error {
	var err error
	for i, tc := range tcs {
		if tc.Type == TestCaseType("") {
			tc.Type = TestCaseTypeStandalone
			tcs[i].Type = TestCaseTypeStandalone
		}

		if retErr := q.validateTestCase(tc); retErr != nil {
			err = multierror.Append(err, errors.Wrapf(retErr, "test case %d", i))
		}
	}
	if err != nil {
		return errors.Wrap(err, "validation")
	}

	for _, tc := range tcs {
		q.addTCase(tc)
	}

	return nil
}

func (q *QASphereCSV) GenerateCSV() (string, error) {
	w := &strings.Builder{}
	if err := q.writeCSV(w); err != nil {
		return "", errors.Wrap(err, "generate csv")
	}
	return w.String(), nil
}

func (q *QASphereCSV) WriteCSVToFile(file string) error {
	f, err := os.Create(file)
	if err != nil {
		return errors.Wrap(err, "create csv")
	}
	defer f.Close()

	if err := q.writeCSV(f); err != nil {
		return errors.Wrap(err, "write csv")
	}

	return nil
}

func (q *QASphereCSV) validateTestCase(tc TestCase) error {
	if tc.CustomFields != nil {
		for systemName := range tc.CustomFields {
			var found bool
			for _, cf := range q.customFields {
				if cf.SystemName == systemName {
					found = true
					break
				}
			}
			if !found {
				return errors.Errorf("custom field %s is not defined in QASphereCSV.customFields", systemName)
			}
		}
	}

	return q.validate.Struct(tc)
}

func (q *QASphereCSV) addTCase(tc TestCase) {
	folderPath := strings.Join(tc.Folder, "/")
	q.folderTCaseMap[folderPath] = append(q.folderTCaseMap[folderPath], tc)

	q.numTCases++
	if (len(tc.Steps)) > q.maxSteps {
		q.maxSteps = len(tc.Steps)
	}
}

func (q *QASphereCSV) getFolders() []string {
	var folders []string
	for folder := range q.folderTCaseMap {
		folders = append(folders, folder)
	}
	slices.Sort(folders)
	return folders
}

func (q *QASphereCSV) getCSVRows() ([][]string, error) {
	rows := make([][]string, 0, q.numTCases+1)
	numCols := len(staticColumns) + 2*q.maxSteps + len(q.customFields)

	rows = append(rows, append(make([]string, 0, numCols), staticColumns...))
	for i := 0; i < q.maxSteps; i++ {
		rows[0] = append(rows[0], fmt.Sprintf("Step %d", i+1), fmt.Sprintf("Expected %d", i+1))
	}

	customFieldsMap := make(map[string]int)
	for i, cf := range q.customFields {
		customFieldHeader := fmt.Sprintf("custom_field_%s_%s", cf.Type, cf.SystemName)
		rows[0] = append(rows[0], customFieldHeader)
		customFieldsMap[cf.SystemName] = i
	}

	folders := q.getFolders()
	for _, f := range folders {
		for _, tc := range q.folderTCaseMap[f] {
			var requirements []string
			for _, req := range tc.Requirements {
				if req.Title == "" && req.URL == "" {
					continue
				}

				requirements = append(requirements, fmt.Sprintf("[%s](%s)", req.Title, req.URL))
			}

			var links []string
			for _, link := range tc.Links {
				links = append(links, fmt.Sprintf("[%s](%s)", link.Title, link.URL))
			}

			var files string
			if len(tc.Files) > 0 {
				filesb, err := json.Marshal(tc.Files)
				if err != nil {
					return nil, errors.Wrap(err, "json marshal files")
				}
				files = string(filesb)
			}

			var parameterValues string
			if len(tc.ParameterValues) > 0 {
				parameterValuesb, err := json.Marshal(tc.ParameterValues)
				if err != nil {
					return nil, errors.Wrap(err, "json marshal parameter values")
				}
				parameterValues = string(parameterValuesb)
			}

			row := make([]string, 0, numCols)
			row = append(row, f, string(tc.Type), tc.Title, tc.LegacyID, strconv.FormatBool(tc.Draft),
				string(tc.Priority), strings.Join(tc.Tags, ","), strings.Join(requirements, ","),
				strings.Join(links, ","), files, tc.Preconditions, parameterValues,
				strings.Join(tc.FilledTCaseTitleSuffixParams, ","))

			numSteps := len(tc.Steps)
			for i := 0; i < q.maxSteps; i++ {
				if i < numSteps {
					row = append(row, tc.Steps[i].Action, tc.Steps[i].Expected)
				} else {
					row = append(row, "", "")
				}
			}

			customFieldCols := make([]string, len(customFieldsMap))
			for systemName, cfValue := range tc.CustomFields {
				cfValueJSON, err := json.Marshal(cfValue)
				if err != nil {
					return nil, errors.Wrap(err, "json marshal custom field value")
				}

				customFieldCols[customFieldsMap[systemName]] = string(cfValueJSON)
			}
			row = append(row, customFieldCols...)

			rows = append(rows, row)
		}
	}

	return rows, nil
}

func (q *QASphereCSV) writeCSV(w io.Writer) error {
	rows, err := q.getCSVRows()
	if err != nil {
		return errors.Wrap(err, "get csv rows")
	}
	return csv.NewWriter(w).WriteAll(rows)
}
