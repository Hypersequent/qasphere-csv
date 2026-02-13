package qascsv

import (
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
		Title:    "tc-with-minimal-fields",
		FolderPath: []string{"root"},
		Priority: "high",
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

const successTestCasesCSV = `Folder,Folder Comment,Type,Name,Legacy ID,Draft,Priority,Tags,Requirements,Links,Files,Preconditions,Parameter Values,Template Suffix Params,Step 1,Expected 1,Step 2,Expected 2
root/child,,standalone,tc-with-all-fields,legacy-id,false,high,"tag1,tag2",[req1](http://req1),"[link-1](http://link1),[link-2](http://link2)","[{""fileName"":""file-1.csv"",""id"":""file-id"",""url"":""http://file1"",""mimeType"":""text/csv"",""size"":10},{""fileName"":""file-1.csv"",""id"":""file-id"",""url"":""http://file1"",""mimeType"":""text/csv"",""size"":10}]",preconditions,,,action-1,expected-1,action-2,expected-2
root/child,,standalone,"tc-with-special-chars.,<>/@$%""""''*&()[]{}+-[BACKTICK]!~;",legacy-id,false,high,"tag1.,<>/@$%""""''*&()[]{}+-[BACKTICK]!~;","[req.,<>/@$%""""''*&()[]{}+-[BACKTICK]!~;]()","[link-1.,<>/@$%""""''*&()[]{}+-[BACKTICK]!~;](http://link1)","[{""fileName"":""file-1.csv"",""id"":""file-id"",""url"":""http://file1"",""mimeType"":""text/csv"",""size"":10}]","preconditions.,<>/@$%""""''*&()[]{}+-[BACKTICK]!~;",,,"action.,<>/@$%""""''*&()[]{}+-[BACKTICK]!~;","expected.,<>/@$%""""''*&()[]{}+-[BACKTICK]!~;",,
root,,standalone,tc-with-minimal-fields,,false,high,,,,,,,,,,,
root,,standalone,tc-with-partial-fields,,true,low,,[](http://req1),,"[{""fileName"":""file-1.csv"",""id"":""file-id"",""url"":""http://file1"",""mimeType"":""text/csv"",""size"":10},{""fileName"":""file-1.csv"",""id"":""file-id"",""url"":""http://file1"",""mimeType"":""text/csv"",""size"":10}]",,,,action-1,,,expected-2
`

var failureTestCases = []TestCase{
	{
		Title:    "",
		FolderPath: []string{"root"},
		Priority: "high",
	}, {
		Title:    strings.Repeat("a", 512), // Exceeds 511 char limit
		FolderPath: []string{"root"},
		Priority: "high",
	}, {
		Title:    "no folder",
		FolderPath: []string{},
		Priority: "high",
	}, {
		Title:    "folder with empty segment",
		FolderPath: []string{"root", ""},
		Priority: "high",
	}, {
		Title:    "folder segment ending with backslash",
		FolderPath: []string{"root\\"},
		Priority: "high",
	}, {
		Title:    "wrong priority",
		FolderPath: []string{"root"},
		Priority: "very high",
	}, {
		Title:    "empty tag",
		FolderPath: []string{"root"},
		Priority: "high",
		Tags:     []string{""},
	}, {
		Title:    "long tag",
		FolderPath: []string{"root"},
		Priority: "high",
		Tags:     []string{strings.Repeat("a", 256)}, // Exceeds 255 char limit
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
		Title:    "link without title and url",
		FolderPath: []string{"root"},
		Priority: "high",
		Links:    []Link{{}},
	}, {
		Title:    "link with no url",
		FolderPath: []string{"root"},
		Priority: "high",
		Links:    []Link{{Title: "link-1"}},
	}, {
		Title:    "link with no title",
		FolderPath: []string{"root"},
		Priority: "high",
		Links:    []Link{{URL: "http://link1"}},
	}, {
		Title:    "link with invalid url",
		FolderPath: []string{"root"},
		Priority: "high",
		Links:    []Link{{Title: "link-1", URL: "ftp://link1"}},
	}, {
		Title:    "file without name",
		FolderPath: []string{"root"},
		Priority: "high",
		Files: []File{
			{
				MimeType: "text/csv",
				Size:     10,
				URL:      "http://file1",
			},
		},
	}, {
		Title:    "file without id and url",
		FolderPath: []string{"root"},
		Priority: "high",
		Files: []File{
			{
				Name:     "file-1.csv",
				MimeType: "text/csv",
				Size:     10,
			},
		},
	}, {
		Title:    "file with invalid url",
		FolderPath: []string{"root"},
		Priority: "high",
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
		Title:    "tc-with-single-custom-field",
		FolderPath: []string{"custom-fields"},
		Priority: "medium",
		CustomFields: map[string]CustomFieldValue{
			"test_env": {
				Value:     "staging",
				IsDefault: false,
			},
		},
	},
	{
		Title:    "tc-with-multiple-custom-fields",
		FolderPath: []string{"custom-fields"},
		Priority: "high",
		Tags:     []string{"regression", "smoke"},
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
		Title:    "tc-with-empty-custom-field-value",
		FolderPath: []string{"custom-fields"},
		Priority: "low",
		CustomFields: map[string]CustomFieldValue{
			"notes": {
				Value:     "",
				IsDefault: false,
			},
		},
	},
	{
		Title:    "tc-with-default-custom-field",
		FolderPath: []string{"custom-fields"},
		Priority: "medium",
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

const customFieldSuccessTestCasesCSV = `Folder,Folder Comment,Type,Name,Legacy ID,Draft,Priority,Tags,Requirements,Links,Files,Preconditions,Parameter Values,Template Suffix Params,Step 1,Expected 1,custom_field_dropdown_test_env,custom_field_dropdown_automation,custom_field_text_notes
custom-fields,,standalone,tc-with-single-custom-field,,false,medium,,,,,,,,,,"{""value"":""staging"",""isDefault"":false}",,
custom-fields,,standalone,tc-with-multiple-custom-fields,,false,high,"regression,smoke",,,,,,,Execute test,Test passes,"{""value"":""production"",""isDefault"":false}","{""value"":""Automated"",""isDefault"":false}","{""value"":""This is a test note with special chars: !@#$%^&*()"",""isDefault"":false}"
custom-fields,,standalone,tc-with-empty-custom-field-value,,false,low,,,,,,,,,,,,"{""value"":"""",""isDefault"":false}"
custom-fields,,standalone,tc-with-default-custom-field,,false,medium,,,,,,,,,,,"{""value"":"""",""isDefault"":false}",
custom-fields/comprehensive,,standalone,tc-with-all-fields-and-custom-fields,CF-001,false,high,"custom,comprehensive",[CF Requirements](http://cf-req),[CF Link](http://cf-link),"[{""fileName"":""cf-test.txt"",""id"":""cf-file-id"",""url"":""http://cf-file"",""mimeType"":""text/plain"",""size"":100}]",Custom field test setup,,,Step 1,Result 1,"{""value"":""development"",""isDefault"":false}","{""value"":""In Progress"",""isDefault"":false}",
`

var customFieldFailureTestCases = []TestCase{
	{
		Title:    "tc-with-undefined-custom-field",
		FolderPath: []string{"custom-fields-errors"},
		Priority: "high",
		CustomFields: map[string]CustomFieldValue{
			"undefined_field": {
				Value: "some value",
			},
		},
	},
	{
		Title:    "tc-with-very-long-custom-field-value",
		FolderPath: []string{"custom-fields-errors"},
		Priority: "medium",
		CustomFields: map[string]CustomFieldValue{
			"notes": {
				Value: strings.Repeat("a", 256), // Exceeds 255 char limit
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
		f.Close()
		os.Remove(tempFileName)
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

func TestFolderSlashEscaping(t *testing.T) {
	qasCSV := NewQASphereCSV()

	err := qasCSV.AddTestCase(TestCase{
		Title:    "tc-in-slash-folder",
		FolderPath: []string{"root/parent", "child/leaf"},
		Priority: "high",
	})
	require.NoError(t, err)

	csv, err := qasCSV.GenerateCSV()
	require.NoError(t, err)

	expected := `Folder,Folder Comment,Type,Name,Legacy ID,Draft,Priority,Tags,Requirements,Links,Files,Preconditions,Parameter Values,Template Suffix Params
root\/parent/child\/leaf,,standalone,tc-in-slash-folder,,false,high,,,,,,,
`
	require.Equal(t, expected, csv)
}

func TestFolderSegmentEndingWithBackslash(t *testing.T) {
	qasCSV := NewQASphereCSV()

	err := qasCSV.AddTestCase(TestCase{
		Title:    "tc-bad-backslash",
		FolderPath: []string{"root\\"},
		Priority: "high",
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

	expected := `Folder,Folder Comment,Type,Name,Legacy ID,Draft,Priority,Tags,Requirements,Links,Files,Preconditions,Parameter Values,Template Suffix Params
empty-folder,,,,,,,,,,,,,
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

	expected := `Folder,Folder Comment,Type,Name,Legacy ID,Draft,Priority,Tags,Requirements,Links,Files,Preconditions,Parameter Values,Template Suffix Params
commented-folder,This is a folder comment,,,,,,,,,,,,
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
		Title:    "tc-in-commented-folder",
		FolderPath: []string{"my-folder"},
		Priority: "high",
	})
	require.NoError(t, err)

	csv, err := qasCSV.GenerateCSV()
	require.NoError(t, err)

	expected := `Folder,Folder Comment,Type,Name,Legacy ID,Draft,Priority,Tags,Requirements,Links,Files,Preconditions,Parameter Values,Template Suffix Params
my-folder,Folder description,,,,,,,,,,,,
my-folder,,standalone,tc-in-commented-folder,,false,high,,,,,,,
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
			Title:    "tc",
			FolderPath: []string{"root"},
			Priority: "high",
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

	expected := `Folder,Folder Comment,Type,Name,Legacy ID,Draft,Priority,Tags,Requirements,Links,Files,Preconditions,Parameter Values,Template Suffix Params
folder\/with\/slashes/child,slash comment,,,,,,,,,,,,
`
	require.Equal(t, expected, csv)
}
