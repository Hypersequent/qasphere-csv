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

var staticColumns = []string{
	"Folder", "Name", "Legacy ID", "Draft", "Priority", "Tags", "Requirements",
	"Links", "Files", "Preconditions",
}

// Priority represents the priority of a test case in QA Sphere.
type Priority string

// The priorities available in QA Sphere.
const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
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

// File represents an external file.
type File struct {
	// The name of the file. (required)
	Name string `validate:"required" json:"file_name"`
	// If the file is already uploaded on QA Sphere, then its ID. (optional)
	ID string `validate:"required_without=URL" json:"id,omitempty"`
	// The URL of the file. If the file is not uploaded on QA Sphere,
	// the URL is required. (optional)
	URL      string `validate:"required_without=ID,omitempty,http_url" json:"url,omitempty"`
	MimeType string `json:"mime_type"`
	Size     int64  `json:"size"`
}

// Step represents a single action to perform in a test case.
type Step struct {
	// The action to perform. Markdown is supported. (optional)
	Action string
	// The expected result of the action. Markdown is supported. (optional)
	Expected string
}

// TestCase represents a test case in QA Sphere.
type TestCase struct {
	// The title of the test case. (required)
	Title string `validate:"required,max=255"`
	// In case of migrating from another test management system, the
	// test case ID in the existing test management system. This is only
	// for reference. (optional)
	LegacyID string `validate:"max=255"`
	// The complete folder path to the test case. (required)
	Folder []string `validate:"min=1,dive,required,max=127,excludesall=/"`
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
	Requirement *Requirement
	// Any other files relevant to the test case. (optional)
	Files []File `validate:"dive"`
	// Any other links relevant to the test case. (optional)
	Links []Link `validate:"dive"`
	// Whether the test case is still work in progress and not in its
	// final state. The test case should later be updated as and then
	// published. (optional)
	Draft bool
}

// QASphereCSV provides APIs to generate CSV that can be used to import
// test cases in a project on QA Sphere.
type QASphereCSV struct {
	folderTCaseMap map[string][]TestCase
	validate       *validator.Validate

	numTCases int
	maxSteps  int
}

func NewQASphereCSV() *QASphereCSV {
	return &QASphereCSV{
		folderTCaseMap: make(map[string][]TestCase),
		validate:       validator.New(),
	}
}

func (q *QASphereCSV) AddTestCase(tc TestCase) error {
	if err := q.validateTestCase(tc); err != nil {
		return errors.Wrap(err, "test case validation")
	}

	q.addTCase(tc)
	return nil
}

func (q *QASphereCSV) AddTestCases(tcs []TestCase) error {
	var err error
	for i, tc := range tcs {
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
	numCols := len(staticColumns) + 2*q.maxSteps

	rows = append(rows, append(make([]string, 0, numCols), staticColumns...))
	for i := 0; i < q.maxSteps; i++ {
		rows[0] = append(rows[0], fmt.Sprintf("Step %d", i+1), fmt.Sprintf("Expected %d", i+1))
	}

	folders := q.getFolders()
	for _, f := range folders {
		for _, tc := range q.folderTCaseMap[f] {
			var requirement string
			if tc.Requirement != nil {
				requirement = fmt.Sprintf("[%s](%s)", tc.Requirement.Title, tc.Requirement.URL)
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

			row := make([]string, 0, numCols)
			row = append(row, f, tc.Title, tc.LegacyID, strconv.FormatBool(tc.Draft),
				string(tc.Priority), strings.Join(tc.Tags, ","), requirement,
				strings.Join(links, ","), files, tc.Preconditions)

			numSteps := len(tc.Steps)
			for i := 0; i < q.maxSteps; i++ {
				if i < numSteps {
					row = append(row, tc.Steps[i].Action, tc.Steps[i].Expected)
				} else {
					row = append(row, "", "")
				}
			}

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
