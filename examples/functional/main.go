package main

import (
	"log"

	qascsv "github.com/hypersequent/qasphere-csv"
)

func main() {
	generateFolderComments()
	generateEmptyFolders()
	generateEscaping()
}

func generateFolderComments() {
	q := qascsv.NewQASphereCSV()

	if err := q.AddFolder([]string{"Commented Folder"}, "This folder has a comment but also contains test cases"); err != nil {
		log.Fatal(err)
	}
	if err := q.AddTestCase(qascsv.TestCase{
		Title:    "Test inside commented folder",
		Folder:   []string{"Commented Folder"},
		Priority: qascsv.PriorityHigh,
	}); err != nil {
		log.Fatal(err)
	}

	if err := q.AddFolder([]string{"Another Folder"}, "Standalone comment on a folder with no test cases"); err != nil {
		log.Fatal(err)
	}

	if err := q.AddFolder([]string{"Parent", "Child With Comment"}, "Nested folder comment"); err != nil {
		log.Fatal(err)
	}
	if err := q.AddTestCase(qascsv.TestCase{
		Title:    "Test in parent folder",
		Folder:   []string{"Parent"},
		Priority: qascsv.PriorityMedium,
	}); err != nil {
		log.Fatal(err)
	}

	if err := q.WriteCSVToFile("folder_comments.csv"); err != nil {
		log.Fatal(err)
	}
	log.Println("wrote folder_comments.csv")
}

func generateEmptyFolders() {
	q := qascsv.NewQASphereCSV()

	if err := q.AddFolder([]string{"Empty Root Folder"}, ""); err != nil {
		log.Fatal(err)
	}
	if err := q.AddFolder([]string{"Parent", "Empty Child"}, ""); err != nil {
		log.Fatal(err)
	}
	if err := q.AddFolder([]string{"Parent", "Another Empty Child"}, ""); err != nil {
		log.Fatal(err)
	}
	if err := q.AddTestCase(qascsv.TestCase{
		Title:    "Test in parent alongside empty children",
		Folder:   []string{"Parent"},
		Priority: qascsv.PriorityLow,
	}); err != nil {
		log.Fatal(err)
	}

	if err := q.AddFolder([]string{"Deep", "Nested", "Empty"}, ""); err != nil {
		log.Fatal(err)
	}

	if err := q.WriteCSVToFile("empty_folders.csv"); err != nil {
		log.Fatal(err)
	}
	log.Println("wrote empty_folders.csv")
}

func generateEscaping() {
	q := qascsv.NewQASphereCSV()

	if err := q.AddTestCase(qascsv.TestCase{
		Title:    "Test in folder with slash",
		Folder:   []string{"Features/Bugs", "Login"},
		Priority: qascsv.PriorityHigh,
	}); err != nil {
		log.Fatal(err)
	}
	if err := q.AddTestCase(qascsv.TestCase{
		Title:    "Test in folder with multiple slashes",
		Folder:   []string{"A/B/C", "D/E"},
		Priority: qascsv.PriorityMedium,
	}); err != nil {
		log.Fatal(err)
	}

	if err := q.AddFolder([]string{"Empty/Slash/Folder"}, "This empty folder name contains slashes"); err != nil {
		log.Fatal(err)
	}

	if err := q.AddTestCase(qascsv.TestCase{
		Title:    "Test in normal folder for comparison",
		Folder:   []string{"Normal Folder", "Subfolder"},
		Priority: qascsv.PriorityLow,
	}); err != nil {
		log.Fatal(err)
	}

	if err := q.WriteCSVToFile("escaping.csv"); err != nil {
		log.Fatal(err)
	}
	log.Println("wrote escaping.csv")
}
