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

## Importing Test Cases on QA Sphere

1. Create a new Project, if not already done.
2. Open the project from the **Dashboard** and navigate to the **Test Cases** tab.
3. Select the **Import** option from the dropdown in the top right.

For more details, please check the [documentation](https://docs.qasphere.com/).

## Contributing

We welcome contributions! If you have a feature request, encounter a problem, or have questions, please [create a new issue](https://github.com/Hypersequent/qasphere-csv/issues/new/choose). You can also contribute by opening a pull request.

Before submitting a pull request, please ensure:
1. Appropriate unit tests are added and existing tests pass - `make test`
2. Lint checks pass - `make lint`

## License

This library is available under the MIT License. For more details, please see the [LICENSE](license) file.
