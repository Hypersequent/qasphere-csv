# qasphere-csv

The `qasphere-csv` Go library simplifies the creation of CSV files for importing test cases into the [QA Sphere](https://qasphere.com/) Test Management System.

>**What is QA Sphere?**  
>QA Sphere is a Test Management System designed to help teams organize their QA process without the clutter
>of overly complex tools. It provides everything you need to manage test cases, schedule runs, and
>keep track of your progress. With features like AI-powered test case creation and automation integrations,
>QA Sphere focuses on making your QA workflows efficient and straightforward.


## Library Features

- Programmatically create large projects instead of manual entries.
- Facilitate migration from older test management systems by converting exported data into QA Sphere's CSV format.
- Includes in-built validations to ensure CSV files meet QA Sphere's requirements for smooth import.

## How to Use

### Starting from Scratch

Clone the repository and explore the [basic example](examples/basic/main.go). Modify the code to add your test cases and run:

```bash
go run examples/basic/main.go
```

Use the `WriteCSVToFile()` method to write directly to a file.

### Integrating into an Existing Project

To include `qasphere-csv` in your Go project, run:

```bash
go get github.com/hypersequent/qasphere-csv
```

Import the library in your Go project:

```go
import qascsv "github.com/hypersequent/qasphere-csv"
```

Refer to the [basic example](examples/basic/main.go) for API usage.

## Steps and Test Data

`Step.Action` and `Step.Expected` remain the API for ordinary test steps. The library serializes all steps into QA Sphere's modern JSON `Steps` CSV column:

```go
Steps: []qascsv.Step{{
	Action:   "Run the query",
	Expected: "One row is returned",
	Data: []qascsv.StepData{
		qascsv.StepDataText{Label: "Query", Value: "SELECT 1", Format: "sql"},
		qascsv.StepDataLink{Label: "Reference", Value: "https://example.com/spec"},
		qascsv.StepDataFile{Label: "Fixture", Value: qascsv.File{
			Name: "fixture.csv", ID: "uploaded-file-id", URL: "https://example.com/fixture.csv",
			MimeType: "text/csv", Size: 128,
		}},
	},
}}
```

Shared steps can reference an existing QA Sphere shared step by ID, or include a title and sub-steps so the shared step can be recreated during import. Test data belongs on sub-steps, not directly on the shared row:

```go
Steps: []qascsv.Step{{
	SharedStepID: 42,
}, {
	Title: "Sign in",
	SubSteps: []qascsv.Step{{
		Action: "Enter the credentials",
		Data: []qascsv.StepData{
			qascsv.StepDataText{Label: "User", Value: "alice"},
		},
	}},
}}
```

Each standalone step or shared-step sub-step accepts at most 20 data items. Labels and shared-step titles are limited to 255 Unicode characters. Text values are required and limited to 65,535 UTF-16 code units; their optional formats are limited to 32 characters. Link values must be HTTP(S) URLs no longer than 255 characters. File data uses the same validation and JSON shape as test-case files.

> **Migration note:** generated CSVs now always contain one JSON `Steps` column after `Preconditions`. Earlier releases generated variable `Step N` and `Expected N` columns; consumers that inspect headers or post-process CSV output must update to the modern column.

## Importing Test Cases on QA Sphere

1. Create a new Project, if not already done.
2. Open the project from the **Dashboard** and navigate to the **Test Cases** tab.
3. Select the **Import** option from the dropdown in the top right.

For more details, please check the [documentation](https://docs.qasphere.com/).

## Custom Fields

Custom fields must be declared with `AddCustomField`/`AddCustomFields` before adding test cases that use them. Three types are supported:

- `text` — plain text, no length limit.
- `dropdown` — the value must match one of the options defined for the field in QA Sphere (option values are limited to 255 characters).
- `richtext` — rich text, no length limit. **Values are HTML** (e.g. `<p>…</p>`, `<pre><code>…</code></pre>`), unlike `Preconditions` and `Steps`, which take markdown. QA Sphere sanitizes the HTML on import using an allowlist of tags and attributes.

For example, to populate QA Sphere's rich text Description field:

```go
qasCSV := qascsv.NewQASphereCSV()
_ = qasCSV.AddCustomField(qascsv.CustomField{
	SystemName: "description",
	Type:       qascsv.CustomFieldTypeRichtext,
})
_ = qasCSV.AddTestCase(qascsv.TestCase{
	Title:      "Login with valid credentials",
	FolderPath: []string{"Auth"},
	Priority:   qascsv.PriorityHigh,
	CustomFields: map[string]qascsv.CustomFieldValue{
		"description": {Value: "<p>Verifies the standard login flow.</p>"},
	},
})
```

This produces a `custom_field_richtext_description` column matching QA Sphere's CSV export format.

## Contributing

We welcome contributions! If you have a feature request, encounter a problem, or have questions, please [create a new issue](https://github.com/Hypersequent/qasphere-csv/issues/new/choose). You can also contribute by opening a pull request.

Before submitting a pull request, please ensure:
1. Appropriate unit tests are added and existing tests pass - `make test`
2. Lint checks pass - `make lint`

## License

This library is available under the MIT License. For more details, please see the [LICENSE](license) file.
