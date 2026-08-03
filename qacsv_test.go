package qascsv

import (
	"encoding/csv"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

var successTestCases = []TestCase{
	{
		Title:         "tc-with-all-fields",
		LegacyID:      "legacy-id",
		FolderPath:    []string{"root", "child"},
		Priority:      "high",
		Tags:          []string{"tag1", "tag2"},
		Preconditions: "preconditions",
		Steps: []Step{
			{
				Action:   "action-1",
				Expected: "expected-1",
			},
			{
				Action:   "action-2",
				Expected: "expected-2",
			},
		},
		Requirements: []Requirement{{Title: "req1", URL: "http://req1"}},
		Files: []File{
			{
				ID:       "file-id",
				Name:     "file-1.csv",
				MimeType: "text/csv",
				Size:     10,
				URL:      "http://file1",
			}, {
				Name:     "file-1.csv",
				ID:       "file-id",
				URL:      "http://file1",
				MimeType: "text/csv",
				Size:     10,
			},
		},
		Links: []Link{
			{
				Title: "link-1",
				URL:   "http://link1",
			}, {
				Title: "link-2",
				URL:   "http://link2",
			},
		},
		Draft: false,
	},
	{
		Title:      "tc-with-minimal-fields",
		FolderPath: []string{"root"},
		Priority:   "high",
	},
	{
		Title:         "tc-with-special-chars.,<>/@$%\"\"''*&()[]{}+-`!~;",
		LegacyID:      "legacy-id",
		FolderPath:    []string{"root", "child"},
		Priority:      "high",
		Tags:          []string{"tag1.,<>/@$%\"\"''*&()[]{}+-`!~;"},
		Preconditions: "preconditions.,<>/@$%\"\"''*&()[]{}+-`!~;",
		Steps: []Step{
			{
				Action:   "action.,<>/@$%\"\"''*&()[]{}+-`!~;",
				Expected: "expected.,<>/@$%\"\"''*&()[]{}+-`!~;",
			},
		},
		Requirements: []Requirement{{Title: "req.,<>/@$%\"\"''*&()[]{}+-`!~;"}},
		Files: []File{
			{
				ID:       "file-id",
				Name:     "file-1.csv",
				MimeType: "text/csv",
				Size:     10,
				URL:      "http://file1",
			},
		},
		Links: []Link{
			{
				Title: "link-1.,<>/@$%\"\"''*&()[]{}+-`!~;",
				URL:   "http://link1",
			},
		},
		Draft: false,
	},
	{
		Title:         "tc-with-partial-fields",
		FolderPath:    []string{"root"},
		Priority:      "low",
		Tags:          []string{},
		Preconditions: "",
		Steps: []Step{
			{
				Action: "action-1",
			},
			{
				Expected: "expected-2",
			},
		},
		Requirements: []Requirement{{URL: "http://req1"}},
		Files: []File{
			{
				ID:       "file-id",
				Name:     "file-1.csv",
				MimeType: "text/csv",
				Size:     10,
				URL:      "http://file1",
			}, {
				Name:     "file-1.csv",
				ID:       "file-id",
				URL:      "http://file1",
				MimeType: "text/csv",
				Size:     10,
			},
		},
		Links: []Link{},
		Draft: true,
	},
}

const successTestCasesCSV = `Folder,Folder Comment,Type,Name,Legacy ID,Draft,Priority,Tags,Requirements,Links,Files,Preconditions,Steps,Parameter Values,Template Suffix Params
root/child,,standalone,tc-with-all-fields,legacy-id,false,high,"tag1,tag2",[req1](http://req1),"[link-1](http://link1),[link-2](http://link2)","[{""fileName"":""file-1.csv"",""id"":""file-id"",""url"":""http://file1"",""mimeType"":""text/csv"",""size"":10},{""fileName"":""file-1.csv"",""id"":""file-id"",""url"":""http://file1"",""mimeType"":""text/csv"",""size"":10}]",preconditions,"[{""description"":""action-1"",""expected"":""expected-1""},{""description"":""action-2"",""expected"":""expected-2""}]",,
root/child,,standalone,"tc-with-special-chars.,<>/@$%""""''*&()[]{}+-[BACKTICK]!~;",legacy-id,false,high,"tag1.,<>/@$%""""''*&()[]{}+-[BACKTICK]!~;","[req.,<>/@$%""""''*&()[]{}+-[BACKTICK]!~;]()","[link-1.,<>/@$%""""''*&()[]{}+-[BACKTICK]!~;](http://link1)","[{""fileName"":""file-1.csv"",""id"":""file-id"",""url"":""http://file1"",""mimeType"":""text/csv"",""size"":10}]","preconditions.,<>/@$%""""''*&()[]{}+-[BACKTICK]!~;","[{""description"":""action.,<>/@$%\""\""''*&()[]{}+-[BACKTICK]!~;"",""expected"":""expected.,<>/@$%\""\""''*&()[]{}+-[BACKTICK]!~;""}]",,
root,,standalone,tc-with-minimal-fields,,false,high,,,,,,,,
root,,standalone,tc-with-partial-fields,,true,low,,[](http://req1),,"[{""fileName"":""file-1.csv"",""id"":""file-id"",""url"":""http://file1"",""mimeType"":""text/csv"",""size"":10},{""fileName"":""file-1.csv"",""id"":""file-id"",""url"":""http://file1"",""mimeType"":""text/csv"",""size"":10}]",,"[{""description"":""action-1""},{""expected"":""expected-2""}]",,
`

var failureTestCases = []TestCase{
	{
		Title:      "",
		FolderPath: []string{"root"},
		Priority:   "high",
	}, {
		Title:      strings.Repeat("a", 512), // Exceeds 511 char limit
		FolderPath: []string{"root"},
		Priority:   "high",
	}, {
		Title:      "no folder",
		FolderPath: []string{},
		Priority:   "high",
	}, {
		Title:      "folder with empty segment",
		FolderPath: []string{"root", ""},
		Priority:   "high",
	}, {
		Title:      "folder segment ending with backslash",
		FolderPath: []string{"root\\"},
		Priority:   "high",
	}, {
		Title:      "wrong priority",
		FolderPath: []string{"root"},
		Priority:   "very high",
	}, {
		Title:      "empty tag",
		FolderPath: []string{"root"},
		Priority:   "high",
		Tags:       []string{""},
	}, {
		Title:      "long tag",
		FolderPath: []string{"root"},
		Priority:   "high",
		Tags:       []string{strings.Repeat("a", 256)}, // Exceeds 255 char limit
	}, {
		Title:        "requirement without title and url",
		FolderPath:   []string{"root"},
		Priority:     "high",
		Requirements: []Requirement{{}},
	}, {
		Title:        "requirement with invalid url",
		FolderPath:   []string{"root"},
		Priority:     "high",
		Requirements: []Requirement{{URL: "ftp://req1"}},
	}, {
		Title:      "link without title and url",
		FolderPath: []string{"root"},
		Priority:   "high",
		Links:      []Link{{}},
	}, {
		Title:      "link with no url",
		FolderPath: []string{"root"},
		Priority:   "high",
		Links:      []Link{{Title: "link-1"}},
	}, {
		Title:      "link with no title",
		FolderPath: []string{"root"},
		Priority:   "high",
		Links:      []Link{{URL: "http://link1"}},
	}, {
		Title:      "link with invalid url",
		FolderPath: []string{"root"},
		Priority:   "high",
		Links:      []Link{{Title: "link-1", URL: "ftp://link1"}},
	}, {
		Title:      "file without name",
		FolderPath: []string{"root"},
		Priority:   "high",
		Files: []File{
			{
				MimeType: "text/csv",
				Size:     10,
				URL:      "http://file1",
			},
		},
	}, {
		Title:      "file without id and url",
		FolderPath: []string{"root"},
		Priority:   "high",
		Files: []File{
			{
				Name:     "file-1.csv",
				MimeType: "text/csv",
				Size:     10,
			},
		},
	}, {
		Title:      "file with invalid url",
		FolderPath: []string{"root"},
		Priority:   "high",
		Files: []File{
			{
				Name:     "file-1.csv",
				MimeType: "text/csv",
				Size:     10,
				URL:      "ftp://file1",
			},
		},
	},
}

// Custom field test data
var customFields = []CustomField{
	{
		SystemName: "test_env",
		Type:       CustomFieldTypeDropdown,
	},
	{
		SystemName: "automation",
		Type:       CustomFieldTypeDropdown,
	},
	{
		SystemName: "notes",
		Type:       CustomFieldTypeText,
	},
}

var customFieldSuccessTestCases = []TestCase{
	{
		Title:      "tc-with-single-custom-field",
		FolderPath: []string{"custom-fields"},
		Priority:   "medium",
		CustomFields: map[string]CustomFieldValue{
			"test_env": {
				Value:     "staging",
				IsDefault: false,
			},
		},
	},
	{
		Title:      "tc-with-multiple-custom-fields",
		FolderPath: []string{"custom-fields"},
		Priority:   "high",
		Tags:       []string{"regression", "smoke"},
		CustomFields: map[string]CustomFieldValue{
			"test_env": {
				Value:     "production",
				IsDefault: false,
			},
			"automation": {
				Value:     "Automated",
				IsDefault: false,
			},
			"notes": {
				Value:     "This is a test note with special chars: !@#$%^&*()",
				IsDefault: false,
			},
		},
		Steps: []Step{
			{
				Action:   "Execute test",
				Expected: "Test passes",
			},
		},
	},
	{
		Title:      "tc-with-empty-custom-field-value",
		FolderPath: []string{"custom-fields"},
		Priority:   "low",
		CustomFields: map[string]CustomFieldValue{
			"notes": {
				Value:     "",
				IsDefault: false,
			},
		},
	},
	{
		Title:      "tc-with-default-custom-field",
		FolderPath: []string{"custom-fields"},
		Priority:   "medium",
		CustomFields: map[string]CustomFieldValue{
			"automation": {
				Value:     "",
				IsDefault: false,
			},
		},
	},
	{
		Title:         "tc-with-all-fields-and-custom-fields",
		LegacyID:      "CF-001",
		FolderPath:    []string{"custom-fields", "comprehensive"},
		Priority:      "high",
		Tags:          []string{"custom", "comprehensive"},
		Preconditions: "Custom field test setup",
		Steps: []Step{
			{
				Action:   "Step 1",
				Expected: "Result 1",
			},
		},
		Requirements: []Requirement{{Title: "CF Requirements", URL: "http://cf-req"}},
		Files: []File{
			{
				ID:       "cf-file-id",
				Name:     "cf-test.txt",
				MimeType: "text/plain",
				Size:     100,
				URL:      "http://cf-file",
			},
		},
		Links: []Link{
			{
				Title: "CF Link",
				URL:   "http://cf-link",
			},
		},
		Draft: false,
		CustomFields: map[string]CustomFieldValue{
			"test_env": {
				Value:     "development",
				IsDefault: false,
			},
			"automation": {
				Value:     "In Progress",
				IsDefault: false,
			},
		},
	},
}

const customFieldSuccessTestCasesCSV = `Folder,Folder Comment,Type,Name,Legacy ID,Draft,Priority,Tags,Requirements,Links,Files,Preconditions,Steps,Parameter Values,Template Suffix Params,custom_field_dropdown_test_env,custom_field_dropdown_automation,custom_field_text_notes
custom-fields,,standalone,tc-with-single-custom-field,,false,medium,,,,,,,,,"{""value"":""staging"",""isDefault"":false}",,
custom-fields,,standalone,tc-with-multiple-custom-fields,,false,high,"regression,smoke",,,,,"[{""description"":""Execute test"",""expected"":""Test passes""}]",,,"{""value"":""production"",""isDefault"":false}","{""value"":""Automated"",""isDefault"":false}","{""value"":""This is a test note with special chars: !@#$%^&*()"",""isDefault"":false}"
custom-fields,,standalone,tc-with-empty-custom-field-value,,false,low,,,,,,,,,,,"{""value"":"""",""isDefault"":false}"
custom-fields,,standalone,tc-with-default-custom-field,,false,medium,,,,,,,,,,"{""value"":"""",""isDefault"":false}",
custom-fields/comprehensive,,standalone,tc-with-all-fields-and-custom-fields,CF-001,false,high,"custom,comprehensive",[CF Requirements](http://cf-req),[CF Link](http://cf-link),"[{""fileName"":""cf-test.txt"",""id"":""cf-file-id"",""url"":""http://cf-file"",""mimeType"":""text/plain"",""size"":100}]",Custom field test setup,"[{""description"":""Step 1"",""expected"":""Result 1""}]",,,"{""value"":""development"",""isDefault"":false}","{""value"":""In Progress"",""isDefault"":false}",
`

var customFieldFailureTestCases = []TestCase{
	{
		Title:      "tc-with-undefined-custom-field",
		FolderPath: []string{"custom-fields-errors"},
		Priority:   "high",
		CustomFields: map[string]CustomFieldValue{
			"undefined_field": {
				Value: "some value",
			},
		},
	},
	{
		Title:      "tc-with-very-long-dropdown-value",
		FolderPath: []string{"custom-fields-errors"},
		Priority:   "medium",
		CustomFields: map[string]CustomFieldValue{
			"test_env": {
				Value: strings.Repeat("a", 256), // Dropdown options are limited to 255 chars
			},
		},
	},
}

func TestGenerateCSVSuccess(t *testing.T) {
	qasCSV := NewQASphereCSV()

	for _, tc := range successTestCases {
		err := qasCSV.AddTestCase(tc)
		require.NoError(t, err)
	}

	actualCSV, err := qasCSV.GenerateCSV()
	require.NoError(t, err)

	require.Equal(t, strings.ReplaceAll(successTestCasesCSV, "[BACKTICK]", "`"), actualCSV)
}

func TestWriteCSVMultipleTCasesSuccess(t *testing.T) {
	tempFileName := "temp.csv"
	qasCSV := NewQASphereCSV()

	err := qasCSV.AddTestCases(successTestCases)
	require.NoError(t, err)
	require.NoError(t, qasCSV.WriteCSVToFile(tempFileName))

	f, err := os.Open(tempFileName)
	require.NoError(t, err)
	defer func() {
		_ = f.Close()
		_ = os.Remove(tempFileName)
	}()

	b, err := io.ReadAll(f)
	require.NoError(t, err)
	require.Equal(t, strings.ReplaceAll(successTestCasesCSV, "[BACKTICK]", "`"), string(b))
}

func TestFailureTestCases(t *testing.T) {
	for _, tc := range failureTestCases {
		t.Run(tc.Title, func(t *testing.T) {
			qasCSV := NewQASphereCSV()
			err := qasCSV.AddTestCase(tc)
			require.NotNil(t, err)
		})
	}
}

func TestCustomFieldSuccessTestCases(t *testing.T) {
	qasCSV := NewQASphereCSV()
	if err := qasCSV.AddCustomFields(customFields); err != nil {
		t.Fatalf("Failed to add custom fields: %v", err)
	}

	for _, tc := range customFieldSuccessTestCases {
		err := qasCSV.AddTestCase(tc)
		require.NoError(t, err)
	}
	actualCSV, err := qasCSV.GenerateCSV()
	require.NoError(t, err)
	require.Equal(t, customFieldSuccessTestCasesCSV, actualCSV)
}

func TestCustomFieldFailureTestCases(t *testing.T) {
	qasCSV := NewQASphereCSV()
	if err := qasCSV.AddCustomFields(customFields); err != nil {
		t.Fatalf("Failed to add custom fields: %v", err)
	}

	for _, tc := range customFieldFailureTestCases {
		t.Run(tc.Title, func(t *testing.T) {
			err := qasCSV.AddTestCase(tc)
			require.NotNil(t, err)
		})
	}
}

func TestRichtextCustomField(t *testing.T) {
	qasCSV := NewQASphereCSV()
	require.NoError(t, qasCSV.AddCustomField(CustomField{
		SystemName: "description",
		Type:       CustomFieldTypeRichtext,
	}))

	// Long multi-line HTML value, well over 255 chars, with quotes and commas
	// to exercise CSV and JSON escaping
	longHTML := "<p>This is a \"long\" description, with commas.</p>\n" +
		"<pre><code>func main() {\n\tfmt.Println(\"hello\")\n}</code></pre>\n" +
		"<p>" + strings.Repeat("Lorem ipsum dolor sit amet. ", 20) + "</p>"
	require.Greater(t, len(longHTML), 255)

	require.NoError(t, qasCSV.AddTestCase(TestCase{
		Title:      "tc-with-richtext-description",
		FolderPath: []string{"richtext"},
		Priority:   "medium",
		CustomFields: map[string]CustomFieldValue{
			"description": {Value: longHTML},
		},
	}))

	csvStr, err := qasCSV.GenerateCSV()
	require.NoError(t, err)

	// Parse the CSV back and verify the value round-trips
	records, err := csv.NewReader(strings.NewReader(csvStr)).ReadAll()
	require.NoError(t, err)
	require.Len(t, records, 2)

	header := records[0]
	require.Equal(t, "custom_field_richtext_description", header[len(header)-1])

	var cfValue CustomFieldValue
	require.NoError(t, json.Unmarshal([]byte(records[1][len(header)-1]), &cfValue))
	require.Equal(t, longHTML, cfValue.Value)
}

func TestLongTextCustomFieldValue(t *testing.T) {
	qasCSV := NewQASphereCSV()
	require.NoError(t, qasCSV.AddCustomField(CustomField{
		SystemName: "notes",
		Type:       CustomFieldTypeText,
	}))

	// Text custom field values have no length limit
	err := qasCSV.AddTestCase(TestCase{
		Title:      "tc-with-long-text-value",
		FolderPath: []string{"root"},
		Priority:   "low",
		CustomFields: map[string]CustomFieldValue{
			"notes": {Value: strings.Repeat("a", 600)},
		},
	})
	require.NoError(t, err)
}

func TestFolderSlashEscaping(t *testing.T) {
	qasCSV := NewQASphereCSV()

	err := qasCSV.AddTestCase(TestCase{
		Title:      "tc-in-slash-folder",
		FolderPath: []string{"root/parent", "child/leaf"},
		Priority:   "high",
	})
	require.NoError(t, err)

	csv, err := qasCSV.GenerateCSV()
	require.NoError(t, err)

	expected := `Folder,Folder Comment,Type,Name,Legacy ID,Draft,Priority,Tags,Requirements,Links,Files,Preconditions,Steps,Parameter Values,Template Suffix Params
root\/parent/child\/leaf,,standalone,tc-in-slash-folder,,false,high,,,,,,,,
`
	require.Equal(t, expected, csv)
}

func TestFolderSegmentEndingWithBackslash(t *testing.T) {
	qasCSV := NewQASphereCSV()

	err := qasCSV.AddTestCase(TestCase{
		Title:      "tc-bad-backslash",
		FolderPath: []string{"root\\"},
		Priority:   "high",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "must not end with '\\'")
}

func TestAddFolderEmpty(t *testing.T) {
	qasCSV := NewQASphereCSV()

	err := qasCSV.AddFolder(Folder{FolderPath: []string{"empty-folder"}})
	require.NoError(t, err)

	csv, err := qasCSV.GenerateCSV()
	require.NoError(t, err)

	expected := `Folder,Folder Comment,Type,Name,Legacy ID,Draft,Priority,Tags,Requirements,Links,Files,Preconditions,Steps,Parameter Values,Template Suffix Params
empty-folder,,,,,,,,,,,,,,
`
	require.Equal(t, expected, csv)
}

func TestAddFolderWithComment(t *testing.T) {
	qasCSV := NewQASphereCSV()

	err := qasCSV.AddFolder(Folder{
		FolderPath: []string{"commented-folder"},
		Comment:    "This is a folder comment",
	})
	require.NoError(t, err)

	csv, err := qasCSV.GenerateCSV()
	require.NoError(t, err)

	expected := `Folder,Folder Comment,Type,Name,Legacy ID,Draft,Priority,Tags,Requirements,Links,Files,Preconditions,Steps,Parameter Values,Template Suffix Params
commented-folder,This is a folder comment,,,,,,,,,,,,,
`
	require.Equal(t, expected, csv)
}

func TestAddFolderWithCommentAndTestCases(t *testing.T) {
	qasCSV := NewQASphereCSV()

	err := qasCSV.AddFolder(Folder{
		FolderPath: []string{"my-folder"},
		Comment:    "Folder description",
	})
	require.NoError(t, err)

	err = qasCSV.AddTestCase(TestCase{
		Title:      "tc-in-commented-folder",
		FolderPath: []string{"my-folder"},
		Priority:   "high",
	})
	require.NoError(t, err)

	csv, err := qasCSV.GenerateCSV()
	require.NoError(t, err)

	expected := `Folder,Folder Comment,Type,Name,Legacy ID,Draft,Priority,Tags,Requirements,Links,Files,Preconditions,Steps,Parameter Values,Template Suffix Params
my-folder,Folder description,,,,,,,,,,,,,
my-folder,,standalone,tc-in-commented-folder,,false,high,,,,,,,,
`
	require.Equal(t, expected, csv)
}

func TestAddFolderValidation(t *testing.T) {
	t.Run("empty folder path", func(t *testing.T) {
		qasCSV := NewQASphereCSV()
		err := qasCSV.AddFolder(Folder{Comment: "comment"})
		require.Error(t, err)
	})

	t.Run("empty folder segment", func(t *testing.T) {
		qasCSV := NewQASphereCSV()
		err := qasCSV.AddFolder(Folder{FolderPath: []string{"root", ""}, Comment: "comment"})
		require.Error(t, err)
	})

	t.Run("folder segment ending with backslash", func(t *testing.T) {
		qasCSV := NewQASphereCSV()
		err := qasCSV.AddFolder(Folder{FolderPath: []string{"root\\"}, Comment: "comment"})
		require.Error(t, err)
		require.Contains(t, err.Error(), "must not end with '\\'")
	})

	t.Run("duplicate folder", func(t *testing.T) {
		qasCSV := NewQASphereCSV()
		err := qasCSV.AddFolder(Folder{FolderPath: []string{"root"}, Comment: "comment"})
		require.NoError(t, err)
		err = qasCSV.AddFolder(Folder{FolderPath: []string{"root"}, Comment: "another comment"})
		require.Error(t, err)
		require.Contains(t, err.Error(), "already exists")
	})

	t.Run("folder already has test cases", func(t *testing.T) {
		qasCSV := NewQASphereCSV()
		err := qasCSV.AddTestCase(TestCase{
			Title:      "tc",
			FolderPath: []string{"root"},
			Priority:   "high",
		})
		require.NoError(t, err)
		err = qasCSV.AddFolder(Folder{FolderPath: []string{"root"}, Comment: "comment"})
		require.Error(t, err)
		require.Contains(t, err.Error(), "already exists")
	})
}

func TestAddFolderWithSlashInName(t *testing.T) {
	qasCSV := NewQASphereCSV()

	err := qasCSV.AddFolder(Folder{
		FolderPath: []string{"folder/with/slashes", "child"},
		Comment:    "slash comment",
	})
	require.NoError(t, err)

	csv, err := qasCSV.GenerateCSV()
	require.NoError(t, err)

	expected := `Folder,Folder Comment,Type,Name,Legacy ID,Draft,Priority,Tags,Requirements,Links,Files,Preconditions,Steps,Parameter Values,Template Suffix Params
folder\/with\/slashes/child,slash comment,,,,,,,,,,,,,
`
	require.Equal(t, expected, csv)
}

func TestModernStepsSerialization(t *testing.T) {
	file := File{
		Name:     "evidence.txt",
		ID:       "file-id",
		URL:      "https://files.example.com/evidence.txt",
		MimeType: "text/plain",
		Size:     12,
	}
	qasCSV := NewQASphereCSV()
	require.NoError(t, qasCSV.AddTestCase(TestCase{
		Title:      "modern steps",
		FolderPath: []string{"root"},
		Priority:   PriorityHigh,
		Steps: []Step{
			{},
			{Action: "Open the page", Expected: "The page opens"},
			{Data: []StepData{
				StepDataText{Label: "Input", Value: "SELECT 1", Format: "sql"},
				StepDataLink{Value: "https://example.com"},
				StepDataFile{Label: "Evidence", Value: file},
			}},
			{SharedStepID: 42},
			{
				Title: "Sign in",
				SubSteps: []Step{
					{},
					{Action: "Enter credentials", Data: []StepData{StepDataText{Value: "alice"}}},
				},
			},
		},
	}))

	csvText, err := qasCSV.GenerateCSV()
	require.NoError(t, err)
	records, err := csv.NewReader(strings.NewReader(csvText)).ReadAll()
	require.NoError(t, err)
	require.Len(t, records, 2)
	require.Equal(t, "Steps", records[0][12])
	require.Equal(t, "Parameter Values", records[0][13])
	require.JSONEq(t, `[
		{"description":"Open the page","expected":"The page opens"},
		{"data":[
			{"type":"text","label":"Input","data":{"value":"SELECT 1","format":"sql"}},
			{"type":"link","data":{"value":"https://example.com"}},
			{"type":"file","label":"Evidence","data":{"value":{"fileName":"evidence.txt","id":"file-id","url":"https://files.example.com/evidence.txt","mimeType":"text/plain","size":12}}}
		]},
		{"sharedStepId":42},
		{"title":"Sign in","subSteps":[{"description":"Enter credentials","data":[{"type":"text","data":{"value":"alice"}}]}]}
	]`, records[1][12])
}

func TestEmptyStepsCell(t *testing.T) {
	qasCSV := NewQASphereCSV()
	require.NoError(t, qasCSV.AddTestCase(testCaseWithSteps([]Step{{}, {Action: "", Expected: ""}})))

	csvText, err := qasCSV.GenerateCSV()
	require.NoError(t, err)
	records, err := csv.NewReader(strings.NewReader(csvText)).ReadAll()
	require.NoError(t, err)
	require.Empty(t, records[1][12])
}

func TestStepDataValidationBoundaries(t *testing.T) {
	validFile := File{Name: "file", ID: "id", URL: "https://example.com/file", MimeType: "text/plain", Size: 1}
	validURL := "https://example.com/" + strings.Repeat("a", 255-len("https://example.com/"))
	validCases := map[string][]Step{
		"20 data items":           {{Data: repeatStepData(20)}},
		"20 sub-step data items":  {{Title: "shared", SubSteps: []Step{{Data: repeatStepData(20)}}}},
		"255 character label":     {{Data: []StepData{StepDataText{Label: strings.Repeat("界", 255), Value: "v"}}}},
		"65535 UTF-16 code units": {{Data: []StepData{StepDataText{Value: strings.Repeat("😀", 32767) + "a"}}}},
		"32 character format":     {{Data: []StepData{StepDataText{Value: "v", Format: strings.Repeat("界", 32)}}}},
		"255 character link":      {{Data: []StepData{StepDataLink{Value: validURL}}}},
		"valid file":              {{Data: []StepData{StepDataFile{Value: validFile}}}},
	}
	for name, steps := range validCases {
		t.Run(name, func(t *testing.T) {
			require.NoError(t, NewQASphereCSV().AddTestCase(testCaseWithSteps(steps)))
		})
	}
}

func TestStepValidationFailures(t *testing.T) {
	var nilText *StepDataText
	invalidCases := map[string][]Step{
		"too many data items":               {{Data: repeatStepData(21)}},
		"too many sub-step data items":      {{Title: "shared", SubSteps: []Step{{Data: repeatStepData(21)}}}},
		"long label":                        {{Data: []StepData{StepDataText{Label: strings.Repeat("界", 256), Value: "v"}}}},
		"missing text":                      {{Data: []StepData{StepDataText{}}}},
		"long UTF-16 text":                  {{Data: []StepData{StepDataText{Value: strings.Repeat("😀", 32768)}}}},
		"long format":                       {{Data: []StepData{StepDataText{Value: "v", Format: strings.Repeat("界", 33)}}}},
		"missing link":                      {{Data: []StepData{StepDataLink{}}}},
		"non-HTTP link":                     {{Data: []StepData{StepDataLink{Value: "ftp://example.com/file"}}}},
		"long link":                         {{Data: []StepData{StepDataLink{Value: "https://example.com/" + strings.Repeat("a", 256)}}}},
		"invalid file":                      {{Data: []StepData{StepDataFile{Value: File{}}}}},
		"nil data":                          {{Data: []StepData{nil}}},
		"typed nil data":                    {{Data: []StepData{nilText}}},
		"negative shared ID":                {{SharedStepID: -1}},
		"long shared title":                 {{Title: strings.Repeat("界", 256)}},
		"shared action":                     {{Title: "shared", Action: "not allowed"}},
		"shared expected":                   {{SharedStepID: 1, Expected: "not allowed"}},
		"shared data":                       {{Title: "shared", Data: []StepData{StepDataText{Value: "v"}}}},
		"sub-steps without shared metadata": {{SubSteps: []Step{{Action: "child"}}}},
		"sub-steps on standalone content":   {{Action: "action", SubSteps: []Step{{Action: "child"}}}},
		"nested shared ID":                  {{Title: "shared", SubSteps: []Step{{SharedStepID: 1}}}},
		"nested shared title":               {{Title: "shared", SubSteps: []Step{{Title: "nested"}}}},
		"nested sub-steps":                  {{Title: "shared", SubSteps: []Step{{SubSteps: []Step{{Action: "nested"}}}}}},
	}
	for name, steps := range invalidCases {
		t.Run(name, func(t *testing.T) {
			require.Error(t, NewQASphereCSV().AddTestCase(testCaseWithSteps(steps)))
		})
	}
}

func TestAddTestCasesStepValidationIsAtomicAndIndexed(t *testing.T) {
	qasCSV := NewQASphereCSV()
	testCases := []TestCase{
		testCaseWithSteps([]Step{{Action: "valid"}}),
		testCaseWithSteps([]Step{{Data: repeatStepData(21)}}),
	}
	testCases[0].Title = "valid"
	testCases[1].Title = "invalid"

	err := qasCSV.AddTestCases(testCases)
	require.Error(t, err)
	require.Contains(t, err.Error(), "test case 1")
	require.Zero(t, qasCSV.numTCases)
	require.Empty(t, qasCSV.folderOrder)
}

func testCaseWithSteps(steps []Step) TestCase {
	return TestCase{
		Title:      "steps test",
		FolderPath: []string{"root"},
		Priority:   PriorityHigh,
		Steps:      steps,
	}
}

func repeatStepData(count int) []StepData {
	items := make([]StepData, count)
	for i := range items {
		items[i] = StepDataText{Value: "value"}
	}
	return items
}
