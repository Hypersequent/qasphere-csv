package main

import (
	"fmt"
	"log"

	qascsv "github.com/hypersequent/qasphere-csv"
)

func main() {
	// Create a new instance of QASphereCSV
	qasCSV := qascsv.NewQASphereCSV()

	// Add a single test case
	if err := qasCSV.AddTestCase(qascsv.TestCase{
		Title:         "Changing to corresponding cursor after hovering the element",
		FolderPath:    []string{"Bistro Delivery", "About Us"},
		Priority:      "low",
		Tags:          []string{"About Us", "Checklist", "REQ-4", "UI"},
		Preconditions: "The \"About Us\" page is opened",
		Steps: []qascsv.Step{{
			Action: "Test the display across various screen sizes (desktop, tablet, mobile) to ensure that blocks and buttons adjust appropriately to different viewport widths",
			Data: []qascsv.StepData{
				qascsv.StepDataText{Label: "Viewports", Value: "1440x900, 768x1024, 390x844"},
				qascsv.StepDataLink{Label: "Design", Value: "https://example.com/about-us-design"},
			},
		}},
	}); err != nil {
		log.Fatal("failed to add single test case", err)
	}

	// Add multiple test cases
	if err := qasCSV.AddTestCases([]qascsv.TestCase{{
		Title:         "Cart should be cleared after making the checkout",
		FolderPath:    []string{"Bistro Delivery", "Cart", "Checkout"},
		Priority:      "medium",
		Tags:          []string{"Cart", "checkout", "REQ-6", "Functional"},
		Preconditions: "1. Order is placed\n2. Successful message is shown",
		Steps: []qascsv.Step{{
			Action:   "Go back to the \"Main\" page",
			Expected: "The \"Cart\" icon is empty",
		}, {
			Action:   "Click the \"Cart\" icon",
			Expected: "The empty state is shown in the \"Cart\" modal",
		}, {
			Title: "Verify an empty cart",
			SubSteps: []qascsv.Step{{
				Action:   "Inspect the cart item count",
				Expected: "The item count is zero",
				Data: []qascsv.StepData{
					qascsv.StepDataText{Label: "Expected count", Value: "0"},
				},
			}},
		}},
	}, {
		Title:         "Changing to corresponding cursor after hovering the element",
		FolderPath:    []string{"Bistro Delivery", "Cart", "Checkout"},
		Priority:      "low",
		Tags:          []string{"Checklist", "REQ-6", "UI", "checkout"},
		Preconditions: "The \"Checkout\" page is opened",
		Steps: []qascsv.Step{{
			Action: "Test the display across various screen sizes (desktop, tablet, mobile) to ensure that blocks and buttons adjust appropriately to different viewport widths",
		}},
	}}); err != nil {
		log.Fatal("failed to add multiple test cases", err)
	}

	// Generate CSV string
	csvStr, err := qasCSV.GenerateCSV()
	if err != nil {
		log.Fatal("failed to generate CSV", err)
	}
	fmt.Println(csvStr)

	// We can also directly write the CSV to a file
	// if err := qascsv.WriteCSVToFile("example.csv"); err != nil {
	// 	log.Fatal("failed to write CSV to file", err)
	// }
}
