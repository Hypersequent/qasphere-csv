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
		Folder:        []string{"root", "child"},
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
		Requirement: &Requirement{Title: "req1", URL: "http://req1"},
		Files: []File{
			{
				Name:     "file-1.csv",
				MimeType: "text/csv",
				Size:     10,
				URL:      "http://file1",
			}, {
				Name:     "file-1.csv",
				ID:       "file-id",
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
		Folder:   []string{"root"},
		Priority: "high",
	},
	{
		Title:         "tc-with-special-chars.,<>/@$%\"\"''*&()[]{}+-`!~;",
		LegacyID:      "legacy-id",
		Folder:        []string{"root", "child"},
		Priority:      "high",
		Tags:          []string{"tag1.,<>/@$%\"\"''*&()[]{}+-`!~;"},
		Preconditions: "preconditions.,<>/@$%\"\"''*&()[]{}+-`!~;",
		Steps: []Step{
			{
				Action:   "action.,<>/@$%\"\"''*&()[]{}+-`!~;",
				Expected: "expected.,<>/@$%\"\"''*&()[]{}+-`!~;",
			},
		},
		Requirement: &Requirement{Title: "req.,<>/@$%\"\"''*&()[]{}+-`!~;"},
		Files: []File{
			{
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
		Folder:        []string{"root"},
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
		Requirement: &Requirement{URL: "http://req1"},
		Files: []File{
			{
				Name:     "file-1.csv",
				MimeType: "text/csv",
				Size:     10,
				URL:      "http://file1",
			}, {
				Name:     "file-1.csv",
				ID:       "file-id",
				MimeType: "text/csv",
				Size:     10,
			},
		},
		Links: []Link{},
		Draft: true,
	},
}

const successTestCasesCSV = `Folder,Name,Legacy ID,Draft,Priority,Tags,Requirements,Links,Files,Preconditions,Step 1,Expected 1,Step 2,Expected 2
root,tc-with-minimal-fields,,false,high,,,,,,,,,
root,tc-with-partial-fields,,true,low,,[](http://req1),,"[{""file_name"":""file-1.csv"",""url"":""http://file1"",""mime_type"":""text/csv"",""size"":10},{""file_name"":""file-1.csv"",""id"":""file-id"",""mime_type"":""text/csv"",""size"":10}]",,action-1,,,expected-2
root/child,tc-with-all-fields,legacy-id,false,high,"tag1,tag2",[req1](http://req1),"[link-1](http://link1),[link-2](http://link2)","[{""file_name"":""file-1.csv"",""url"":""http://file1"",""mime_type"":""text/csv"",""size"":10},{""file_name"":""file-1.csv"",""id"":""file-id"",""mime_type"":""text/csv"",""size"":10}]",preconditions,action-1,expected-1,action-2,expected-2
root/child,"tc-with-special-chars.,<>/@$%""""''*&()[]{}+-[BACKTICK]!~;",legacy-id,false,high,"tag1.,<>/@$%""""''*&()[]{}+-[BACKTICK]!~;","[req.,<>/@$%""""''*&()[]{}+-[BACKTICK]!~;]()","[link-1.,<>/@$%""""''*&()[]{}+-[BACKTICK]!~;](http://link1)","[{""file_name"":""file-1.csv"",""url"":""http://file1"",""mime_type"":""text/csv"",""size"":10}]","preconditions.,<>/@$%""""''*&()[]{}+-[BACKTICK]!~;","action.,<>/@$%""""''*&()[]{}+-[BACKTICK]!~;","expected.,<>/@$%""""''*&()[]{}+-[BACKTICK]!~;",,
`

var failureTestCases = []TestCase{
	{
		Title:    "",
		Folder:   []string{"root"},
		Priority: "high",
	}, {
		Title:    "very-long-title-very-long-title-very-long-title-very-long-title-very-long-title-very-long-title-very-long-title-very-long-title-very-long-title-very-long-title-very-long-title-very-long-title-very-long-title-very-long-title-very-long-title-very-long-title-",
		Folder:   []string{"root"},
		Priority: "high",
	}, {
		Title:    "no folder",
		Folder:   []string{},
		Priority: "high",
	}, {
		Title:    "folder with empty title",
		Folder:   []string{"root/child"},
		Priority: "high",
	}, {
		Title:    "folder title with slash",
		Folder:   []string{"root/child"},
		Priority: "high",
	}, {
		Title:    "wrong priority",
		Folder:   []string{"root"},
		Priority: "very high",
	}, {
		Title:    "empty tag",
		Folder:   []string{"root"},
		Priority: "high",
		Tags:     []string{""},
	}, {
		Title:    "long tag",
		Folder:   []string{"root"},
		Priority: "high",
		Tags:     []string{"very-long-tag-very-long-tag-very-long-tag-very-long-tag-very-long-tag-very-long-tag-very-long-tag-very-long-tag-very-long-tag-very-long-tag-very-long-tag-very-long-tag-very-long-tag-very-long-tag-very-long-tag-very-long-tag-very-long-tag-very-long-tag-very"},
	}, {
		Title:       "requirement without title and url",
		Folder:      []string{"root"},
		Priority:    "high",
		Requirement: &Requirement{},
	}, {
		Title:       "requirement with invalid url",
		Folder:      []string{"root"},
		Priority:    "high",
		Requirement: &Requirement{URL: "ftp://req1"},
	}, {
		Title:    "link without title and url",
		Folder:   []string{"root"},
		Priority: "high",
		Links:    []Link{{}},
	}, {
		Title:    "link with no url",
		Folder:   []string{"root"},
		Priority: "high",
		Links:    []Link{{Title: "link-1"}},
	}, {
		Title:    "link with no title",
		Folder:   []string{"root"},
		Priority: "high",
		Links:    []Link{{URL: "http://link1"}},
	}, {
		Title:    "link with invalid url",
		Folder:   []string{"root"},
		Priority: "high",
		Links:    []Link{{Title: "link-1", URL: "ftp://link1"}},
	}, {
		Title:    "file without name",
		Folder:   []string{"root"},
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
		Folder:   []string{"root"},
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
		Folder:   []string{"root"},
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
	require.Equal(t, strings.ReplaceAll(string(b), "[BACKTICK]", "`"), string(b))
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
