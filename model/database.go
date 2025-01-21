package model

type Database struct {
	ID    string
	Name  string
	Files []DatabaseFile
}

type DatabaseFile struct {
	LogicalName  string
	PhysicalName string
	FileType     string
}

type MergedDatabaseFileInfo struct {
	DatabaseId   string
	DatabaseName string
	LogicalName  string
	PhysicalName string
	File_type    string
}
