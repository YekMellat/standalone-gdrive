package drive

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"google.golang.org/api/drive/v3"
)

// TeamDrive represents a Google Shared Drive (Team Drive)
type TeamDrive struct {
	ID          string
	Name        string
	Description string
	ColorRGB    string
}

// ListTeamDrives returns a list of team drives the user has access to
func (f *Fs) ListTeamDrives(ctx context.Context) ([]TeamDrive, error) {
	if f.svc == nil {
		return nil, errors.New("client not initialized")
	}

	var drives []TeamDrive
	var nextPageToken string
	for {
		call := f.svc.Drives.List().PageSize(100)
		if nextPageToken != "" {
			call = call.PageToken(nextPageToken)
		}

		res, err := call.Context(ctx).Do()
		if err != nil {
			return nil, ProcessError(err)
		}
		for _, d := range res.Drives {
			drives = append(drives, TeamDrive{
				ID:          d.Id,
				Name:        d.Name,
				Description: "", // Description field not available in API
				ColorRGB:    d.ThemeId,
			})
		}

		nextPageToken = res.NextPageToken
		if nextPageToken == "" {
			break
		}
	}

	return drives, nil
}

// GetTeamDrive returns information about a specific team drive
func (f *Fs) GetTeamDrive(ctx context.Context, teamDriveID string) (*TeamDrive, error) {
	if f.svc == nil {
		return nil, errors.New("client not initialized")
	}
	drive, err := f.svc.Drives.Get(teamDriveID).Context(ctx).Do()
	if err != nil {
		return nil, ProcessError(err)
	}
	return &TeamDrive{
		ID:       drive.Id,
		Name:     drive.Name,
		ColorRGB: drive.ThemeId,
	}, nil
}

// ListFilesInTeamDrive lists files in a team drive folder
func (f *Fs) ListFilesInTeamDrive(ctx context.Context, folderID, teamDriveID string, recursive bool) ([]*drive.File, error) {
	if f.svc == nil {
		return nil, errors.New("client not initialized")
	}

	var query string
	if folderID == "root" {
		query = "driveId = '" + teamDriveID + "' and 'root' in parents"
	} else {
		query = "'" + folderID + "' in parents"
	}

	var files []*drive.File
	var nextPageToken string
	for {
		call := f.svc.Files.List().
			SupportsAllDrives(true).
			IncludeItemsFromAllDrives(true).
			DriveId(teamDriveID).
			Spaces("drive").
			Corpora("drive").
			Q(query).
			OrderBy("folder,name").
			PageSize(100)

		if nextPageToken != "" {
			call = call.PageToken(nextPageToken)
		}

		res, err := call.Context(ctx).Fields("files(id,name,mimeType,size,md5Checksum,createdTime,modifiedTime),nextPageToken").Do()
		if err != nil {
			return nil, ProcessError(err)
		}
		for _, file := range res.Files {
			files = append(files, file)

			if recursive && file.MimeType == "application/vnd.google-apps.folder" {
				subFiles, err := f.ListFilesInTeamDrive(ctx, file.Id, teamDriveID, true)
				if err != nil {
					return nil, err
				}
				files = append(files, subFiles...)
			}
		}

		nextPageToken = res.NextPageToken
		if nextPageToken == "" {
			break
		}
	}

	return files, nil
}

// UploadFileToTeamDrive uploads a file to a specific team drive
func (f *Fs) UploadFileToTeamDrive(ctx context.Context, localPath, parentID, teamDriveID, filename string) (*drive.File, error) {
	if f.svc == nil {
		return nil, errors.New("client not initialized")
	}
	// Open the local file
	file, err := os.Open(localPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// If filename is not provided, use the original filename
	if filename == "" {
		filename = filepath.Base(localPath)
	}

	// Create the file metadata
	fileMetadata := &drive.File{
		Name:    filename,
		Parents: []string{parentID},
	}

	// Perform the upload with team drive parameters
	res, err := f.svc.Files.Create(fileMetadata).
		Media(file).
		SupportsAllDrives(true).
		Context(ctx).
		Do()

	if err != nil {
		return nil, err
	}

	return res, nil
}
