// Package qascsv provides APIs to generate CSV files that can be used to import
// test cases in a QA Sphere project.
package qascsv

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

// staticColumns will always be present in the CSV file,
// but there can be additional columns for custom fields.
var staticColumns = []string{
	"Folder", "Folder Comment", "Type", "Name", "Legacy ID", "Draft", "Priority", "Tags", "Requirements",
	"Links", "Files", "Preconditions", "Steps", "Parameter Values", "Template Suffix Params",
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
	// Data is the ordered test data associated with a standalone step or a
	// shared step's sub-step. A step can contain at most 20 data items. (optional)
	Data []StepData
	// SharedStepID references an existing shared step. A positive ID marks this
	// row as a shared step. (optional)
	SharedStepID int
	// Title names a shared step and can be used with SubSteps to recreate one
	// during import. (optional)
	Title string
	// SubSteps contains the standalone steps embedded in a shared step. (optional)
	SubSteps []Step
}

// StepData is a test data item attached to a standalone step or shared-step
// sub-step. The concrete type determines the JSON discriminator.
type StepData interface {
	isStepData()
}

// StepDataText represents text or source-code test data.
type StepDataText struct {
	Label  string
	Value  string
	Format string
}

func (StepDataText) isStepData() {}

// StepDataLink represents an HTTP(S) link used as test data.
type StepDataLink struct {
	Label string
	Value string
}

func (StepDataLink) isStepData() {}

// StepDataFile represents an uploaded file used as test data.
type StepDataFile struct {
	Label string
	Value File
}

func (StepDataFile) isStepData() {}

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
	// CustomFieldTypeText is a plain text field.
	CustomFieldTypeText CustomFieldType = "text"
	// CustomFieldTypeDropdown is a selection field. The value must match one
	// of the options defined for the field in QA Sphere (option values are
	// limited to 255 characters).
	CustomFieldTypeDropdown CustomFieldType = "dropdown"
	// CustomFieldTypeRichtext is a rich text field (e.g. the Description
	// field). Unlike Preconditions and Steps, which take markdown, richtext
	// values are HTML, e.g. "<p>…</p>" or "<pre><code>…</code></pre>".
	// QA Sphere sanitizes the HTML on import using an allowlist of tags and
	// attributes; disallowed markup is stripped.
	CustomFieldTypeRichtext CustomFieldType = "richtext"
)

type CustomField struct {
	SystemName string          `validate:"required,max=64"`
	Type       CustomFieldType `validate:"required,oneof=text dropdown richtext"`
}

// CustomFieldValue represents the value of a custom field on a test case.
// QA Sphere does not limit the length of custom field values, but dropdown
// values must match one of the field's options, which are limited to 255
// characters.
type CustomFieldValue struct {
	Value     string `json:"value"`
	IsDefault bool   `json:"isDefault" validate:"omitempty"`
}

// Folder represents a folder to be created in QA Sphere.
type Folder struct {
	// The folder path segments. (required)
	FolderPath []string `validate:"min=1,dive,required,max=255"`
	// An optional comment for the folder.
	Comment string
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
	FolderPath []string `validate:"min=1,dive,required,max=255"`
	// The priority of the test case. (required)
	Priority Priority `validate:"required,oneof=low medium high"`
	// The tags to assign to the test cases. This can be used to group,
	// filter or organise related test cases and also helps in creating
	// test runs. (optional)
	Tags []string `validate:"dive,required,max=255"`
	// The preconditions for the test case. Markdown is supported. (optional)
	// For test case descriptions, use a richtext custom field instead —
	// see CustomFieldTypeRichtext.
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
	folderTCaseMap   map[string][]TestCase
	folderCommentMap map[string]string
	folderOrder      []string
	validate         *validator.Validate
	customFields     []CustomField

	numTCases int
}

func NewQASphereCSV() *QASphereCSV {
	return &QASphereCSV{
		folderTCaseMap:   make(map[string][]TestCase),
		folderCommentMap: make(map[string]string),
		validate:         validator.New(),
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

func (q *QASphereCSV) AddFolder(f Folder) error {
	if err := q.validate.Struct(f); err != nil {
		return errors.Wrap(err, "folder validation")
	}
	if err := validateFolderSegments(f.FolderPath); err != nil {
		return err
	}

	folderPath := escapeFolderPath(f.FolderPath)
	if _, exists := q.folderTCaseMap[folderPath]; exists {
		return errors.Errorf("folder %q already exists", folderPath)
	}

	q.folderOrder = append(q.folderOrder, folderPath)
	q.folderTCaseMap[folderPath] = nil
	if f.Comment != "" {
		q.folderCommentMap[folderPath] = f.Comment
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
	if err := q.writeCSV(f); err != nil {
		_ = f.Close()
		return errors.Wrap(err, "write csv")
	}

	return errors.Wrap(f.Close(), "close csv")
}

func validateFolderSegments(segments []string) error {
	for _, seg := range segments {
		if strings.HasSuffix(seg, `\`) {
			return errors.Errorf("folder segment %q must not end with '\\'", seg)
		}
	}
	return nil
}

func escapeFolderPath(segments []string) string {
	escaped := make([]string, len(segments))
	for i, seg := range segments {
		escaped[i] = strings.ReplaceAll(seg, "/", `\/`)
	}
	return strings.Join(escaped, "/")
}

func (q *QASphereCSV) validateTestCase(tc TestCase) error {
	if err := validateFolderSegments(tc.FolderPath); err != nil {
		return err
	}

	if tc.CustomFields != nil {
		for systemName, cfValue := range tc.CustomFields {
			var found *CustomField
			for i, cf := range q.customFields {
				if cf.SystemName == systemName {
					found = &q.customFields[i]
					break
				}
			}
			if found == nil {
				return errors.Errorf("custom field %s is not defined in QASphereCSV.customFields", systemName)
			}
			// Dropdown values must match an option defined in QA Sphere,
			// and options are limited to 255 characters.
			if found.Type == CustomFieldTypeDropdown && utf8.RuneCountInString(cfValue.Value) > 255 {
				return errors.Errorf("custom field %s: dropdown value must not exceed 255 characters", systemName)
			}
		}
	}
	if err := q.validateSteps(tc.Steps); err != nil {
		return err
	}

	return q.validate.Struct(tc)
}

func (q *QASphereCSV) validateSteps(steps []Step) error {
	for i, step := range steps {
		if err := q.validateStep(step, false); err != nil {
			return errors.Wrapf(err, "steps[%d]", i)
		}
	}
	return nil
}

func (q *QASphereCSV) validateStep(step Step, subStep bool) error {
	if step.SharedStepID < 0 {
		return errors.New("shared step ID cannot be negative")
	}

	if subStep {
		if step.SharedStepID != 0 || step.Title != "" || step.SubSteps != nil {
			return errors.New("shared-step metadata is not allowed on a sub-step")
		}
		return q.validateStepData(step.Data)
	}

	isShared := step.SharedStepID != 0 || step.Title != ""
	if !isShared {
		if step.SubSteps != nil {
			return errors.New("sub-steps are not allowed on a standalone step")
		}
		return q.validateStepData(step.Data)
	}
	if utf8.RuneCountInString(step.Title) > 255 {
		return errors.New("shared step title must not exceed 255 characters")
	}
	if step.Action != "" || step.Expected != "" || len(step.Data) > 0 {
		return errors.New("action, expected result, and data are not allowed directly on a shared step")
	}
	for i, child := range step.SubSteps {
		if err := q.validateStep(child, true); err != nil {
			return errors.Wrapf(err, "subSteps[%d]", i)
		}
	}
	return nil
}

func (q *QASphereCSV) validateStepData(items []StepData) error {
	if len(items) > 20 {
		return errors.New("step data must not contain more than 20 items")
	}
	for i, item := range items {
		if err := q.validateStepDataItem(item); err != nil {
			return errors.Wrapf(err, "data[%d]", i)
		}
	}
	return nil
}

func (q *QASphereCSV) validateStepDataItem(item StepData) error {
	switch value := item.(type) {
	case StepDataText:
		return validateStepDataText(value)
	case *StepDataText:
		if value == nil {
			return errors.New("step data item cannot be nil")
		}
		return validateStepDataText(*value)
	case StepDataLink:
		return validateStepDataLink(value)
	case *StepDataLink:
		if value == nil {
			return errors.New("step data item cannot be nil")
		}
		return validateStepDataLink(*value)
	case StepDataFile:
		if err := validateStepDataLabel(value.Label); err != nil {
			return err
		}
		return errors.Wrap(q.validate.Struct(value.Value), "file validation")
	case *StepDataFile:
		if value == nil {
			return errors.New("step data item cannot be nil")
		}
		if err := validateStepDataLabel(value.Label); err != nil {
			return err
		}
		return errors.Wrap(q.validate.Struct(value.Value), "file validation")
	case nil:
		return errors.New("step data item cannot be nil")
	default:
		return errors.Errorf("unsupported step data type %T", item)
	}
}

func validateStepDataText(value StepDataText) error {
	if err := validateStepDataLabel(value.Label); err != nil {
		return err
	}
	if value.Value == "" {
		return errors.New("text value is required")
	}
	if utf16Length(value.Value) > 65535 {
		return errors.New("text value must not exceed 65,535 UTF-16 code units")
	}
	if utf8.RuneCountInString(value.Format) > 32 {
		return errors.New("text format must not exceed 32 characters")
	}
	return nil
}

func validateStepDataLink(value StepDataLink) error {
	if err := validateStepDataLabel(value.Label); err != nil {
		return err
	}
	if value.Value == "" {
		return errors.New("link value is required")
	}
	if utf8.RuneCountInString(value.Value) > 255 {
		return errors.New("link value must not exceed 255 characters")
	}
	parsed, err := url.Parse(value.Value)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Hostname() == "" {
		return errors.New("link value must be a valid HTTP(S) URL")
	}
	return nil
}

func validateStepDataLabel(label string) error {
	if utf8.RuneCountInString(label) > 255 {
		return errors.New("step data label must not exceed 255 characters")
	}
	return nil
}

func utf16Length(value string) int {
	return len(utf16.Encode([]rune(value)))
}

func (q *QASphereCSV) addTCase(tc TestCase) {
	folderPath := escapeFolderPath(tc.FolderPath)
	if _, exists := q.folderTCaseMap[folderPath]; !exists {
		q.folderOrder = append(q.folderOrder, folderPath)
	}
	q.folderTCaseMap[folderPath] = append(q.folderTCaseMap[folderPath], tc)

	q.numTCases++
}

func (q *QASphereCSV) getFolders() []string {
	return q.folderOrder
}

func jsonMarshal(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return bytes.TrimSuffix(buf.Bytes(), []byte("\n")), nil
}

func (q *QASphereCSV) getCSVRows() ([][]string, error) {
	rows := make([][]string, 0, q.numTCases+1)
	numCols := len(staticColumns) + len(q.customFields)

	rows = append(rows, append(make([]string, 0, numCols), staticColumns...))

	customFieldsMap := make(map[string]int)
	for i, cf := range q.customFields {
		customFieldHeader := fmt.Sprintf("custom_field_%s_%s", cf.Type, cf.SystemName)
		rows[0] = append(rows[0], customFieldHeader)
		customFieldsMap[cf.SystemName] = i
	}

	folders := q.getFolders()
	for _, f := range folders {
		tcs := q.folderTCaseMap[f]
		comment := q.folderCommentMap[f]

		// Write a folder-only row if the folder is empty or has a comment
		if len(tcs) == 0 || comment != "" {
			row := make([]string, numCols)
			row[0] = f
			row[1] = comment
			rows = append(rows, row)
		}

		for _, tc := range tcs {
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
				filesb, err := jsonMarshal(tc.Files)
				if err != nil {
					return nil, errors.Wrap(err, "json marshal files")
				}
				files = string(filesb)
			}

			var parameterValues string
			if len(tc.ParameterValues) > 0 {
				parameterValuesb, err := jsonMarshal(tc.ParameterValues)
				if err != nil {
					return nil, errors.Wrap(err, "json marshal parameter values")
				}
				parameterValues = string(parameterValuesb)
			}

			var steps string
			stepsJSON, err := marshalSteps(tc.Steps)
			if err != nil {
				return nil, errors.Wrap(err, "json marshal steps")
			}
			if len(stepsJSON) > 0 {
				stepsb, err := jsonMarshal(stepsJSON)
				if err != nil {
					return nil, errors.Wrap(err, "json marshal steps")
				}
				steps = string(stepsb)
			}

			row := make([]string, 0, numCols)
			row = append(row, f, "", string(tc.Type), tc.Title, tc.LegacyID, strconv.FormatBool(tc.Draft),
				string(tc.Priority), strings.Join(tc.Tags, ","), strings.Join(requirements, ","),
				strings.Join(links, ","), files, tc.Preconditions, steps, parameterValues,
				strings.Join(tc.FilledTCaseTitleSuffixParams, ","))

			customFieldCols := make([]string, len(customFieldsMap))
			for systemName, cfValue := range tc.CustomFields {
				cfValueJSON, err := jsonMarshal(cfValue)
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

type stepJSON struct {
	Description  string         `json:"description,omitempty"`
	Expected     string         `json:"expected,omitempty"`
	Data         []stepDataJSON `json:"data,omitempty"`
	SharedStepID int            `json:"sharedStepId,omitempty"`
	Title        string         `json:"title,omitempty"`
	SubSteps     []stepJSON     `json:"subSteps,omitempty"`
}

type stepDataJSON struct {
	Type  string `json:"type"`
	Label string `json:"label,omitempty"`
	Data  any    `json:"data"`
}

type stepDataTextJSON struct {
	Value  string `json:"value"`
	Format string `json:"format,omitempty"`
}

type stepDataValueJSON[T any] struct {
	Value T `json:"value"`
}

func marshalSteps(steps []Step) ([]stepJSON, error) {
	result := make([]stepJSON, 0, len(steps))
	for _, step := range steps {
		converted, keep, err := marshalStep(step)
		if err != nil {
			return nil, err
		}
		if keep {
			result = append(result, converted)
		}
	}
	return result, nil
}

func marshalStep(step Step) (stepJSON, bool, error) {
	data, err := marshalStepData(step.Data)
	if err != nil {
		return stepJSON{}, false, err
	}
	subSteps, err := marshalSteps(step.SubSteps)
	if err != nil {
		return stepJSON{}, false, err
	}
	result := stepJSON{
		Description:  step.Action,
		Expected:     step.Expected,
		Data:         data,
		SharedStepID: step.SharedStepID,
		Title:        step.Title,
		SubSteps:     subSteps,
	}
	keep := result.Description != "" || result.Expected != "" || len(result.Data) > 0 ||
		result.SharedStepID != 0 || result.Title != "" || len(result.SubSteps) > 0
	return result, keep, nil
}

func marshalStepData(items []StepData) ([]stepDataJSON, error) {
	result := make([]stepDataJSON, 0, len(items))
	for _, item := range items {
		converted, err := marshalStepDataItem(item)
		if err != nil {
			return nil, err
		}
		result = append(result, converted)
	}
	return result, nil
}

func marshalStepDataItem(item StepData) (stepDataJSON, error) {
	switch value := item.(type) {
	case StepDataText:
		return stepDataJSON{Type: "text", Label: value.Label, Data: stepDataTextJSON{Value: value.Value, Format: value.Format}}, nil
	case *StepDataText:
		if value == nil {
			return stepDataJSON{}, errors.New("step data item cannot be nil")
		}
		return marshalStepDataItem(*value)
	case StepDataLink:
		return stepDataJSON{Type: "link", Label: value.Label, Data: stepDataValueJSON[string]{Value: value.Value}}, nil
	case *StepDataLink:
		if value == nil {
			return stepDataJSON{}, errors.New("step data item cannot be nil")
		}
		return marshalStepDataItem(*value)
	case StepDataFile:
		return stepDataJSON{Type: "file", Label: value.Label, Data: stepDataValueJSON[File]{Value: value.Value}}, nil
	case *StepDataFile:
		if value == nil {
			return stepDataJSON{}, errors.New("step data item cannot be nil")
		}
		return marshalStepDataItem(*value)
	case nil:
		return stepDataJSON{}, errors.New("step data item cannot be nil")
	default:
		return stepDataJSON{}, errors.Errorf("unsupported step data type %T", item)
	}
}

func (q *QASphereCSV) writeCSV(w io.Writer) error {
	rows, err := q.getCSVRows()
	if err != nil {
		return errors.Wrap(err, "get csv rows")
	}
	return csv.NewWriter(w).WriteAll(rows)
}
