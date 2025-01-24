package model

type Database struct {
	ID    string         `json:"id"`
	Name  string         `json:"name" binding:"required"`
	Files []DatabaseFile `json:"files"`
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
