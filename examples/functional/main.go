package main

import (
	"log"

	qascsv "github.com/hypersequent/qasphere-csv"
)

func main() {
	if err := generateFolderComments(); err != nil {
		log.Fatal(err)
	}
	if err := generateEmptyFolders(); err != nil {
		log.Fatal(err)
	}
	if err := generateEscaping(); err != nil {
		log.Fatal(err)
	}
}

func generateFolderComments() error {
	q := qascsv.NewQASphereCSV()

	if err := q.AddFolder(qascsv.Folder{
		FolderPath: []string{"Commented Folder"},
		Comment:    "This folder has a comment but also contains test cases",
	}); err != nil {
		return err
	}
	if err := q.AddTestCase(qascsv.TestCase{
		Title:    "Test inside commented folder",
		Folder:   []string{"Commented Folder"},
		Priority: qascsv.PriorityHigh,
	}); err != nil {
		return err
	}

	if err := q.AddFolder(qascsv.Folder{
		FolderPath: []string{"Another Folder"},
		Comment:    "Standalone comment on a folder with no test cases",
	}); err != nil {
		return err
	}

	if err := q.AddFolder(qascsv.Folder{
		FolderPath: []string{"Parent", "Child With Comment"},
		Comment:    "Nested folder comment",
	}); err != nil {
		return err
	}
	if err := q.AddTestCase(qascsv.TestCase{
		Title:    "Test in parent folder",
		Folder:   []string{"Parent"},
		Priority: qascsv.PriorityMedium,
	}); err != nil {
		return err
	}

	if err := q.WriteCSVToFile("folder_comments.csv"); err != nil {
		return err
	}
	log.Println("wrote folder_comments.csv")
	return nil
}

func generateEmptyFolders() error {
	q := qascsv.NewQASphereCSV()

	if err := q.AddFolder(qascsv.Folder{
		FolderPath: []string{"Empty Root Folder"},
	}); err != nil {
		return err
	}
	if err := q.AddFolder(qascsv.Folder{
		FolderPath: []string{"Parent", "Empty Child"},
	}); err != nil {
		return err
	}
	if err := q.AddFolder(qascsv.Folder{
		FolderPath: []string{"Parent", "Another Empty Child"},
	}); err != nil {
		return err
	}
	if err := q.AddTestCase(qascsv.TestCase{
		Title:    "Test in parent alongside empty children",
		Folder:   []string{"Parent"},
		Priority: qascsv.PriorityLow,
	}); err != nil {
		return err
	}

	if err := q.AddFolder(qascsv.Folder{
		FolderPath: []string{"Deep", "Nested", "Empty"},
	}); err != nil {
		return err
	}

	if err := q.WriteCSVToFile("empty_folders.csv"); err != nil {
		return err
	}
	log.Println("wrote empty_folders.csv")
	return nil
}

func generateEscaping() error {
	q := qascsv.NewQASphereCSV()

	if err := q.AddTestCase(qascsv.TestCase{
		Title:    "Test in folder with slash",
		Folder:   []string{"Features/Bugs", "Login"},
		Priority: qascsv.PriorityHigh,
	}); err != nil {
		return err
	}
	if err := q.AddTestCase(qascsv.TestCase{
		Title:    "Test in folder with multiple slashes",
		Folder:   []string{"A/B/C", "D/E"},
		Priority: qascsv.PriorityMedium,
	}); err != nil {
		return err
	}

	if err := q.AddFolder(qascsv.Folder{
		FolderPath: []string{"Empty/Slash/Folder"},
		Comment:    "This empty folder name contains slashes",
	}); err != nil {
		return err
	}

	if err := q.AddTestCase(qascsv.TestCase{
		Title:    "Test in normal folder for comparison",
		Folder:   []string{"Normal Folder", "Subfolder"},
		Priority: qascsv.PriorityLow,
	}); err != nil {
		return err
	}

	if err := q.WriteCSVToFile("escaping.csv"); err != nil {
		return err
	}
	log.Println("wrote escaping.csv")
	return nil
}
