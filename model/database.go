package model

// Database is a set of id, name, and the database files of some SQL Server database. Its populated by JSON, via HTTP request. Expects an id, name and files in the request body.
type Database struct {
	ID    string         `json:"id,omitempty"`
	Name  string         `json:"name" binding:"required"`
	Files []DatabaseFile `json:"files,omitempty"`
}

// DatabaseFile is a set of a LogicalName, PhysicalName and a FileType (data or log). It refers to a SQL Server database file.
type DatabaseFile struct {
	LogicalName  string `json:"logicalName"`
	PhysicalName string `json:"physicalName"`
	FileType     string `json:"fileType"`
}

// MergedDatabaseFileInfo is a set of DatabaseId, DatabaseName, LogicalName, PhysicalName, and FileType. Typically, when SELECT is executed on repository.GetDatabases(),
// it returns information about the database and its files. For each file in a database, one row will be returned. This is where MergedDatabaseFileInfo is used.
type MergedDatabaseFileInfo struct {
	DatabaseId   string
	DatabaseName string
	LogicalName  string
	PhysicalName string
	FileType     string
}

// BackupDataFile is a set of LogicalName, PhysicalName, FileType, FileGroupName, Size, MaxSize, FileId, CreateLSN, DropLSN, UniqueId, ReadOnlyLSN, ReadWriteLSN, BackupSizeInBytes,
// SourceBlockSize, FileGroupId, LogGroupGUID, DifferentialBaseLSN, DifferentialBaseGUID, IsReadOnly, IsPresent, TDEThumbprint, SnapshotUrl.
// This is used for allocate RESTORE FILELISTONLY information. More informations about each attribute in https://learn.microsoft.com/en-us/sql/t-sql/statements/restore-statements-filelistonly-transact-sql?view=sql-server-ver16
type BackupDataFile struct {
	LogicalName          string
	PhysicalName         string
	FileType             string
	FileGroupName        *string
	Size                 string
	MaxSize              string
	FileId               string
	CreateLSN            string
	DropLSN              *string
	UniqueId             string
	ReadOnlyLSN          *string
	ReadWriteLSN         *string
	BackupSizeInBytes    string
	SourceBlockSize      string
	FileGroupId          string
	LogGroupGUID         *string
	DifferentialBaseLSN  *string
	DifferentialBaseGUID *string
	IsReadOnly           string
	IsPresent            string
	TDEThumbprint        *string
	SnapshotUrl          *string
}

// DatabaseFromBackupFile is a set of Name, BackupFilePath, and BackupFileInfo. It is used to return information collected from the RESTORE FILELISTONLY statement,
// which reads the .bak file (in the BackupFilePath folder) related to the database and its files in an organized and unified manner.
type DatabaseFromBackupFile struct {
	Name           string
	BackupFilePath string
	BackupFileInfo []BackupDataFile
}

// RestoreDb is a set of BackupPath and Database. Its used to return the RESTORE DATABASE completed.
type RestoreDb struct {
	BackupPath string   `json:"backupPath"`
	Database   Database `json:"database"`
}
