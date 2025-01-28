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

type BackupDataFile struct {
	LogicalName          string
	PhysicalName         string
	FileType             string
	FileGroupName        *string
	Size                 string
	MaxSize              string
	FileId               string
	CreateLSN            string
	DropLSN              string
	UniqueId             string
	ReadOnlyLSN          string
	ReadWriteLSN         string
	BackupSizeInBytes    string
	SourceBlockSize      string
	FileGroupId          string
	LogGroupGUID         *string
	DifferentialBaseLSN  string
	DifferentialBaseGUID string
	IsReadOnly           string
	IsPresent            string
	TDEThumbprint        *string
	SnapshotUrl          *string
}

type DatabaseFromBackupFile struct {
	Name           string
	BackupFilePath string
	BackupFileInfo []BackupDataFile
}

type BackupFiles struct {
	Path string `json:"backupFilesPath"`
}

type RestoreDb struct {
	BackupPath string
	Database   Database
}
