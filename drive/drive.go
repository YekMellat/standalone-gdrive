// Package drive implements a Google Drive client for standalone usage
package drive

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/standalone-gdrive/fs"
	"github.com/standalone-gdrive/fs/hash"
	"github.com/standalone-gdrive/lib/dircache"
	"github.com/standalone-gdrive/lib/oauthutil"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	drive_v2 "google.golang.org/api/drive/v2"
	drive "google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

// Constants
const (
	rcloneClientID              = "202264815644.apps.googleusercontent.com"
	rcloneEncryptedClientSecret = "eX8GpZTVx3vxMWVkuuBdDWmAUE6rGhTwVrvG9GhllYccSdj2-mvHVg"
	driveFolderType             = "application/vnd.google-apps.folder"
	shortcutMimeType            = "application/vnd.google-apps.shortcut"
	shortcutMimeTypeDangling    = "application/vnd.google-apps.shortcut.dangling" // synthetic mime type for internal use
	timeFormatIn                = time.RFC3339
	timeFormatOut               = "2006-01-02T15:04:05.000000000Z07:00"
	defaultMinSleep             = fs.Duration(100 * time.Millisecond)
	defaultBurst                = 100
	defaultExportExtensions     = "docx,xlsx,pptx,svg"
	scopePrefix                 = "https://www.googleapis.com/auth/"
	defaultScope                = "drive"
	// chunkSize is the size of the chunks created during a resumable upload and should be a power of two.
	// 1<<18 is the minimum size supported by the Google uploader, and there is no maximum.
	minChunkSize     = fs.SizeSuffix(googleapi.MinUploadChunkSize)
	defaultChunkSize = 8 * fs.MiByte
	partialFields    = "id,name,size,md5Checksum,sha1Checksum,sha256Checksum,trashed,explicitlyTrashed,modifiedTime,createdTime,mimeType,parents,webViewLink,shortcutDetails,exportLinks,resourceKey"
)

// Globals
var (
	// Description of how to auth for this app
	driveConfig = &oauthutil.Config{
		Scopes:       []string{scopePrefix + "drive"},
		AuthURL:      google.Endpoint.AuthURL,
		TokenURL:     google.Endpoint.TokenURL,
		ClientID:     rcloneClientID,
		ClientSecret: rcloneEncryptedClientSecret, // Use the encrypted client secret
		RedirectURL:  oauthutil.RedirectURL,
	}

	// MIME type mapping
	_mimeTypeToExtension = map[string]string{
		"application/epub+zip":                            ".epub",
		"application/json":                                ".json",
		"application/msword":                              ".doc",
		"application/pdf":                                 ".pdf",
		"application/rtf":                                 ".rtf",
		"application/vnd.ms-excel":                        ".xls",
		"application/vnd.oasis.opendocument.presentation": ".odp",
		"application/vnd.oasis.opendocument.spreadsheet":  ".ods",
		"application/vnd.oasis.opendocument.text":         ".odt",
		"application/vnd.openxmlformats-officedocument.presentationml.presentation": ".pptx",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         ".xlsx",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document":   ".docx",
		"application/x-msmetafile":  ".wmf",
		"application/zip":           ".zip",
		"image/bmp":                 ".bmp",
		"image/jpeg":                ".jpg",
		"image/pjpeg":               ".pjpeg",
		"image/png":                 ".png",
		"image/svg+xml":             ".svg",
		"text/csv":                  ".csv",
		"text/html":                 ".html",
		"text/plain":                ".txt",
		"text/tab-separated-values": ".tsv",
		"text/markdown":             ".md",
	}
)

// Options defines the configuration for this backend
type Options struct {
	ConfigDir                 string        `json:"config_dir"`
	Scope                     string        `json:"scope"`
	RootFolderID              string        `json:"root_folder_id"`
	ServiceAccountFile        string        `json:"service_account_file"`
	ServiceAccountCredentials string        `json:"service_account_credentials"`
	TeamDriveID               string        `json:"team_drive"`
	AuthOwnerOnly             bool          `json:"auth_owner_only"`
	UseTrash                  bool          `json:"use_trash"`
	SkipGdocs                 bool          `json:"skip_gdocs"`
	SkipChecksumGphotos       bool          `json:"skip_checksum_gphotos"`
	SharedWithMe              bool          `json:"shared_with_me"`
	TrashedOnly               bool          `json:"trashed_only"`
	StarredOnly               bool          `json:"starred_only"`
	Formats                   string        `json:"formats"`
	ExportFormats             string        `json:"export_formats"`
	ImportFormats             string        `json:"import_formats"`
	AllowImportNameChange     bool          `json:"allow_import_name_change"`
	UseCreatedDate            bool          `json:"use_created_date"`
	UseSharedDate             bool          `json:"use_shared_date"`
	ListChunk                 int           `json:"list_chunk"`
	Impersonate               string        `json:"impersonate"`
	AlternateExport           bool          `json:"alternate_export"`
	UploadCutoff              fs.SizeSuffix `json:"upload_cutoff"`
	ChunkSize                 fs.SizeSuffix `json:"chunk_size"`
	AcknowledgeAbuse          bool          `json:"acknowledge_abuse"`
	KeepRevisionForever       bool          `json:"keep_revision_forever"`
	SizeAsQuota               bool          `json:"size_as_quota"`
	PacerMinSleep             fs.Duration   `json:"pacer_min_sleep"`
	PacerBurst                int           `json:"pacer_burst"`
	ServerSideAcrossConfigs   bool          `json:"server_side_across_configs"`
	DisableHTTP2              bool          `json:"disable_http2"`
	StopOnUploadLimit         bool          `json:"stop_on_upload_limit"`
	StopOnDownloadLimit       bool          `json:"stop_on_download_limit"`
	SkipShortcuts             bool          `json:"skip_shortcuts"`
	SkipDanglingShortcuts     bool          `json:"skip_dangling_shortcuts"`
	ResourceKey               string        `json:"resource_key"`
	V2DownloadMinSize         fs.SizeSuffix `json:"v2_download_min_size"`
	EnvAuth                   bool          `json:"env_auth"`
	LogLevel                  string        `json:"log_level"`
	LogOutput                 string        `json:"log_output"` // path to log file, empty for stderr
}

// Fs represents a remote drive server
type Fs struct {
	name             string                       // name of this remote
	root             string                       // the path we are working on
	opt              Options                      // parsed options
	features         *fs.Features                 // optional features
	svc              *drive.Service               // the connection to the drive server
	v2Svc            *drive_v2.Service            // used to create download links for the v2 api
	client           *http.Client                 // authorized client
	rootFolderID     string                       // the id of the root folder
	dirCache         *dircache.DirCache           // Map of directory path to directory id
	pacer            *fs.Pacer                    // To pace the API calls
	exportExtensions []string                     // preferred extensions to download docs
	importMimeTypes  []string                     // MIME types to convert to docs
	isTeamDrive      bool                         // true if this is a team drive
	dirResourceKeys  *sync.Map                    // map directory ID to resource key
	permissionsMu    *sync.Mutex                  // protect the below
	permissions      map[string]*drive.Permission // map permission IDs to Permissions
	logger           *Logger                      // logging system
}

type baseObject struct {
	fs           *Fs      // what this object is part of
	remote       string   // The remote path
	id           string   // Drive Id of this object
	modifiedDate string   // RFC3339 time it was last modified
	mimeType     string   // The object MIME type
	bytes        int64    // size of the object
	parents      []string // IDs of the parent directories
	resourceKey  *string  // resourceKey is needed for link shared objects
}

type documentObject struct {
	baseObject
	url              string // Download URL of this object
	documentMimeType string // the original document MIME type
	extLen           int    // The length of the added export extension
}

type linkObject struct {
	baseObject
	content []byte // The file content generated by a link template
	extLen  int    // The length of the added export extension
}

// Object describes a drive object
type Object struct {
	baseObject
	url        string // Download URL of this object
	md5sum     string // md5sum of the object
	sha1sum    string // sha1sum of the object
	sha256sum  string // sha256sum of the object
	v2Download bool   // generate v2 download link ondemand
}

// Directory describes a drive directory
type Directory struct {
	baseObject
}

// ------------------------------------------------------------

// Name of the remote (as passed into NewFs)
func (f *Fs) Name() string {
	return f.name
}

// Root of the remote (as passed into NewFs)
func (f *Fs) Root() string {
	return f.root
}

// String converts this Fs to a string
func (f *Fs) String() string {
	return fmt.Sprintf("Google drive root '%s'", f.root)
}

// Features returns the optional features of this Fs
func (f *Fs) Features() *fs.Features {
	return f.features
}

// shouldRetry determines whether a given err rates being retried
func (f *Fs) shouldRetry(ctx context.Context, err error) (bool, error) {
	if err == nil {
		return false, nil
	}

	switch gerr := err.(type) {
	case *googleapi.Error:
		if gerr.Code >= 500 && gerr.Code < 600 {
			// All 5xx errors should be retried
			return true, err
		}
		if len(gerr.Errors) > 0 {
			reason := gerr.Errors[0].Reason
			if reason == "rateLimitExceeded" || reason == "userRateLimitExceeded" {
				return true, err
			} else if reason == "downloadQuotaExceeded" {
				return false, errors.New("download quota exceeded")
			} else if reason == "quotaExceeded" || reason == "storageQuotaExceeded" {
				return false, errors.New("storage quota exceeded")
			}
		}
	}
	return false, err
}

// parseDrivePath parses a drive 'url' and validates the path
func parseDrivePath(inputPath string) (root string, err error) {
	// Handle special cases
	if inputPath == "." {
		return "", nil
	}

	// Clean the path using path.Clean (not filepath.Clean) to keep forward slashes
	cleaned := strings.Trim(path.Clean(inputPath), "/")

	// Check for invalid characters
	if strings.ContainsAny(cleaned, "*?") {
		return "", errors.New("invalid characters in path")
	}

	// Remove duplicate slashes and handle relative paths
	parts := strings.Split(cleaned, "/")
	var result []string

	for _, part := range parts {
		if part == "" {
			continue // Skip empty parts (multiple slashes)
		}
		if part == ".." {
			if len(result) > 0 {
				result = result[:len(result)-1] // Remove last element for ../
			}
			continue
		}
		if part == "." {
			continue // Skip current directory references
		}
		result = append(result, part)
	}

	return strings.Join(result, "/"), nil
}

// getClient returns an http client with appropriate timeouts
func getClient(ctx context.Context, opt *Options) *http.Client {
	t := &http.Transport{
		TLSClientConfig: &tls.Config{},
	}
	if opt.DisableHTTP2 {
		t.TLSNextProto = map[string]func(string, *tls.Conn) http.RoundTripper{}
	}
	return &http.Client{
		Transport: t,
	}
}

// Parse the scopes option returning a slice of scopes
func driveScopes(scopesString string) (scopes []string) {
	if scopesString == "" {
		scopesString = defaultScope
	}
	for _, scope := range strings.Split(scopesString, ",") {
		scope = strings.TrimSpace(scope)
		scopes = append(scopes, scopePrefix+scope)
	}
	return scopes
}

func getServiceAccountClient(ctx context.Context, opt *Options, credentialsData []byte) (*http.Client, error) {
	scopes := driveScopes(opt.Scope)
	conf, err := google.JWTConfigFromJSON(credentialsData, scopes...)
	if err != nil {
		return nil, fmt.Errorf("error processing credentials: %w", err)
	}
	if opt.Impersonate != "" {
		conf.Subject = opt.Impersonate
	}
	ctxWithClient := context.WithValue(ctx, oauth2.HTTPClient, getClient(ctx, opt))
	return oauth2.NewClient(ctxWithClient, conf.TokenSource(ctxWithClient)), nil
}

func createOAuthClient(ctx context.Context, opt *Options, name string) (*http.Client, error) {
	var oAuthClient *http.Client
	var err error

	// Try loading service account credentials from env variable, then from a file
	if len(opt.ServiceAccountCredentials) == 0 && opt.ServiceAccountFile != "" {
		loadedCreds, err := os.ReadFile(opt.ServiceAccountFile)
		if err != nil {
			return nil, fmt.Errorf("error opening service account credentials file: %w", err)
		}
		opt.ServiceAccountCredentials = string(loadedCreds)
	}
	if opt.ServiceAccountCredentials != "" {
		oAuthClient, err = getServiceAccountClient(ctx, opt, []byte(opt.ServiceAccountCredentials))
		if err != nil {
			return nil, fmt.Errorf("failed to create oauth client from service account: %w", err)
		}
	} else if opt.EnvAuth {
		scopes := driveScopes(opt.Scope)
		oAuthClient, err = google.DefaultClient(ctx, scopes...)
		if err != nil {
			return nil, fmt.Errorf("failed to create client from environment: %w", err)
		}
	} else {
		// Set custom scopes if needed
		driveConfig.Scopes = driveScopes(opt.Scope)

		// Create config map with config directory
		configMap := map[string]string{
			"config_dir": opt.ConfigDir,
		}

		oAuthClient, _, err = oauthutil.NewClient(ctx, name, configMap, driveConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create oauth client: %w", err)
		}
	}

	return oAuthClient, nil
}

func newFs(ctx context.Context, name, path string, opt *Options) (*Fs, error) {
	if opt.ChunkSize < minChunkSize {
		return nil, fmt.Errorf("chunk size must be at least %s", minChunkSize)
	}

	// Set default config directory if not provided
	if opt.ConfigDir == "" {
		home, err := os.UserHomeDir()
		if err == nil {
			opt.ConfigDir = filepath.Join(home, ".config", "standalone-gdrive")
		} else {
			opt.ConfigDir = ".config/standalone-gdrive"
		}
	}

	// Create the client
	oAuthClient, err := createOAuthClient(ctx, opt, name)
	if err != nil {
		return nil, fmt.Errorf("drive: failed when making oauth client: %w", err)
	}

	// Parse the path
	root, err := parseDrivePath(path)
	if err != nil {
		return nil, err
	}
	// Initialize logger
	var logWriter io.Writer
	if opt.LogOutput != "" {
		var err error
		logWriter, err = os.OpenFile(opt.LogOutput, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("couldn't open log file: %w", err)
		}
	}

	// Parse log level
	logLevel := LogLevelInfo // default
	if opt.LogLevel != "" {
		var err error
		logLevel, err = ParseLogLevel(opt.LogLevel)
		if err != nil {
			return nil, err
		}
	}

	// Create the Fs object
	f := &Fs{
		name: name,
		root: root,
		opt:  *opt,
		pacer: fs.NewPacer(ctx, func(state fs.PacerState) time.Duration {
			return time.Millisecond * 100
		}),
		dirResourceKeys: new(sync.Map),
		permissionsMu:   new(sync.Mutex),
		permissions:     make(map[string]*drive.Permission),
		logger:          NewLogger(logLevel, logWriter),
	}

	// Set up features
	f.features = &fs.Features{
		DuplicateFiles:          true,
		ReadMimeType:            true,
		WriteMimeType:           true,
		CanHaveEmptyDirectories: true,
		ServerSideAcrossConfigs: opt.ServerSideAcrossConfigs,
	}

	// Set if this is a team drive
	f.isTeamDrive = opt.TeamDriveID != ""
	// Create a new authorized Drive client
	f.client = oAuthClient

	// Import version package for user-agent string
	userAgent := "standalone-gdrive/unknown"
	if versionUserAgent := getVersionUserAgent(); versionUserAgent != "" {
		userAgent = versionUserAgent
	}

	f.svc, err = drive.NewService(ctx,
		option.WithHTTPClient(f.client),
		option.WithUserAgent(userAgent))
	if err != nil {
		return nil, fmt.Errorf("couldn't create Drive client: %w", err)
	}

	// Create v2 API client if needed for downloading
	if f.opt.V2DownloadMinSize >= 0 {
		f.v2Svc, err = drive_v2.NewService(ctx,
			option.WithHTTPClient(f.client),
			option.WithUserAgent(userAgent))
		if err != nil {
			return nil, fmt.Errorf("couldn't create Drive v2 client: %w", err)
		}
	}

	// Set the root folder ID
	if f.opt.RootFolderID != "" {
		// Use the provided root folder ID
		f.rootFolderID = f.opt.RootFolderID
	} else if f.isTeamDrive {
		// Use the team drive ID as the root
		f.rootFolderID = f.opt.TeamDriveID
	} else {
		// Otherwise use "root"
		f.rootFolderID = "root"
	}

	// Create the directory cache
	f.dirCache = dircache.New(f.root, f.rootFolderID, f)

	// Parse export extensions
	if opt.ExportFormats != "" {
		f.exportExtensions = strings.Split(opt.ExportFormats, ",")
	}

	return f, nil
}

// NewFs constructs an Fs from the path, container:path
func NewFs(ctx context.Context, name, path string, m map[string]string) (fs.Fs, error) {
	// Parse config into Options struct
	opt := &Options{
		Scope:             "drive",
		ChunkSize:         defaultChunkSize,
		UploadCutoff:      defaultChunkSize,
		ExportFormats:     defaultExportExtensions,
		UseTrash:          true,
		PacerMinSleep:     defaultMinSleep,
		PacerBurst:        defaultBurst,
		V2DownloadMinSize: -1, // Disabled initially
	}
	// Override with provided config if any
	if m != nil {
		// Parse drive type to set appropriate scope
		if driveType, ok := m["type"]; ok {
			switch strings.ToLower(driveType) {
			case "drive":
				opt.Scope = "drive"
			case "drive.readonly":
				opt.Scope = "drive.readonly"
			case "drive.file":
				opt.Scope = "drive.file"
			case "drive.appdata":
				opt.Scope = "drive.appdata"
			case "drive.metadata":
				opt.Scope = "drive.metadata.readonly"
			case "drive.photos":
				opt.Scope = "drive.photos.readonly"
			default:
				// Use explicit scope if provided, otherwise default to full drive
				if scope, ok := m["scope"]; ok {
					opt.Scope = scope
				}
			}
		} else if scope, ok := m["scope"]; ok {
			opt.Scope = scope
		}

		if rootFolderID, ok := m["root_folder_id"]; ok {
			opt.RootFolderID = rootFolderID
		}
		if teamDriveID, ok := m["team_drive"]; ok {
			opt.TeamDriveID = teamDriveID
		}
	}

	return newFs(ctx, name, path, opt)
}

// FindLeaf implements dircache.DirCacher
func (f *Fs) FindLeaf(ctx context.Context, directoryID, name string) (string, bool, error) {
	var query string
	if directoryID == "" {
		return "", false, errors.New("internal error: directory ID is blank")
	}

	// Make the query
	if f.opt.TrashedOnly {
		query = fmt.Sprintf("name=%q and trashed=true", name)
	} else {
		query = fmt.Sprintf("name=%q and trashed=false", name)
	}

	// Add parent directory filter
	query = fmt.Sprintf("%s and %q in parents", query, directoryID)

	// Add team drive filter if needed
	if f.isTeamDrive {
		query = fmt.Sprintf("%s and driveId=%q", query, f.opt.TeamDriveID)
		if f.opt.SharedWithMe {
			query = fmt.Sprintf("%s and sharedWithMe=true", query)
		}
	}
	// Search for the file/directory
	var fields = partialFields
	var files []*drive.File
	err := f.pacer.Call(ctx, func() error {
		var fileList *drive.FileList
		var err error

		// List the files with the specified query
		fileList, err = f.svc.Files.List().Q(query).Fields(googleapi.Field(fields)).SupportsAllDrives(f.isTeamDrive).IncludeItemsFromAllDrives(f.isTeamDrive).Do()
		if err != nil {
			return err
		}
		files = fileList.Files
		return nil
	})

	if err != nil {
		return "", false, err
	}

	if len(files) == 0 {
		return "", false, nil
	}

	if len(files) > 1 {
		return "", false, fmt.Errorf("found multiple files with the same name: %q", name)
	}

	return files[0].Id, true, nil
}

// CreateDir makes a directory with pathID as parent and name leaf
func (f *Fs) CreateDir(ctx context.Context, parentID, leaf string) (id string, err error) {
	// File metadata
	createInfo := &drive.File{
		Name:     leaf,
		MimeType: driveFolderType,
		Parents:  []string{parentID},
	}
	var info *drive.File
	err = f.pacer.Call(ctx, func() (err error) {
		info, err = f.svc.Files.Create(createInfo).Fields("id").SupportsAllDrives(f.isTeamDrive).Do()
		return err
	})
	if err != nil {
		return "", err
	}
	return info.Id, nil
}

// List the objects and directories in dir into entries
func (f *Fs) List(ctx context.Context, dir string) (entries fs.DirEntries, err error) {
	directoryID, err := f.dirCache.FindDir(ctx, dir)
	if err != nil {
		return nil, err
	}

	var query string
	if f.opt.TrashedOnly {
		query = "trashed=true"
	} else {
		query = "trashed=false"
	}

	// Add parent directory filter
	query = fmt.Sprintf("%s and %q in parents", query, directoryID)

	// Add team drive filter if needed
	if f.isTeamDrive {
		query = fmt.Sprintf("%s and driveId=%q", query, f.opt.TeamDriveID)
	}

	// Add starred filter if needed
	if f.opt.StarredOnly {
		query = fmt.Sprintf("%s and starred=true", query)
	}
	// Search for files and directories
	var files []*drive.File
	err = f.pacer.Call(ctx, func() error {
		var fileList *drive.FileList
		var err error

		// List the files with the specified query
		fileList, err = f.svc.Files.List().Q(query).Fields(googleapi.Field(partialFields)).SupportsAllDrives(f.isTeamDrive).IncludeItemsFromAllDrives(f.isTeamDrive).Do()
		if err != nil {
			return err
		}
		files = fileList.Files
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Process the files
	for _, file := range files {
		remote := path.Join(dir, file.Name)

		// Skip files we don't want
		if f.opt.SkipGdocs && isGoogleDocument(file) {
			continue
		}

		if file.MimeType == driveFolderType {
			// Directory
			d := &Directory{
				baseObject: baseObject{
					fs:           f,
					remote:       remote,
					id:           file.Id,
					modifiedDate: file.ModifiedTime,
					mimeType:     file.MimeType,
					bytes:        0,
					parents:      file.Parents,
				},
			}
			entries = append(entries, d)
		} else {
			// File
			o := &Object{
				baseObject: baseObject{
					fs:           f,
					remote:       remote,
					id:           file.Id,
					modifiedDate: file.ModifiedTime,
					mimeType:     file.MimeType,
					bytes:        file.Size,
					parents:      file.Parents,
				},
				md5sum:     file.Md5Checksum,
				sha1sum:    file.Sha1Checksum,
				sha256sum:  file.Sha256Checksum,
				v2Download: f.opt.V2DownloadMinSize >= 0 && file.Size >= int64(f.opt.V2DownloadMinSize),
			}
			entries = append(entries, o)
		}
	}

	return entries, nil
}

// NewObject finds the Object at remote
func (f *Fs) NewObject(ctx context.Context, remote string) (fs.Object, error) {
	// Find directory containing the object
	leaf := path.Base(remote)
	directoryID, err := f.dirCache.FindDir(ctx, path.Dir(remote))
	if err != nil {
		return nil, err
	}

	// Find the object in the directory
	id, found, err := f.FindLeaf(ctx, directoryID, leaf)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, fs.ErrorObjectNotFound
	}
	// Get the object metadata
	var info *drive.File
	err = f.pacer.Call(ctx, func() error {
		info, err = f.svc.Files.Get(id).Fields(googleapi.Field(partialFields)).SupportsAllDrives(f.isTeamDrive).Do()
		return err
	})
	if err != nil {
		return nil, err
	}

	// Create the object
	if info.MimeType == driveFolderType {
		return nil, fs.ErrorIsDir
	}

	// Create the object
	o := &Object{
		baseObject: baseObject{
			fs:           f,
			remote:       remote,
			id:           info.Id,
			modifiedDate: info.ModifiedTime,
			mimeType:     info.MimeType,
			bytes:        info.Size,
			parents:      info.Parents,
		},
		md5sum:     info.Md5Checksum,
		sha1sum:    info.Sha1Checksum,
		sha256sum:  info.Sha256Checksum,
		v2Download: f.opt.V2DownloadMinSize >= 0 && info.Size >= int64(f.opt.V2DownloadMinSize),
	}

	return o, nil
}

// Put uploads a file
func (f *Fs) Put(ctx context.Context, in io.Reader, src fs.ObjectInfo, options ...fs.OpenOption) (fs.Object, error) {
	// Create temporary file to upload
	size := src.Size()
	if size < 0 {
		size = 0
	}

	// Get the directory to upload to
	leaf := path.Base(src.Remote())
	directoryID, err := f.dirCache.FindDir(ctx, path.Dir(src.Remote()))
	if err != nil {
		return nil, err
	}

	createInfo := &drive.File{
		Name:    leaf,
		Parents: []string{directoryID},
	}

	// Set modification time
	modTime := src.ModTime(ctx)
	createInfo.ModifiedTime = modTime.Format(timeFormatOut)

	// Determine upload strategy based on file size
	var info *drive.File
	if size > int64(f.opt.UploadCutoff) {
		// Upload in chunks
		info, err = f.uploadChunked(ctx, in, size, createInfo)
	} else {
		// Simple upload
		info, err = f.upload(ctx, in, size, createInfo)
	}

	if err != nil {
		return nil, err
	}

	// Create a new object from the response
	o := &Object{
		baseObject: baseObject{
			fs:           f,
			remote:       src.Remote(),
			id:           info.Id,
			modifiedDate: info.ModifiedTime,
			mimeType:     info.MimeType,
			bytes:        info.Size,
			parents:      info.Parents,
		},
		md5sum:     info.Md5Checksum,
		sha1sum:    info.Sha1Checksum,
		sha256sum:  info.Sha256Checksum,
		v2Download: f.opt.V2DownloadMinSize >= 0 && info.Size >= int64(f.opt.V2DownloadMinSize),
	}

	return o, nil
}

// upload uploads a file using a simple method
func (f *Fs) upload(ctx context.Context, in io.Reader, size int64, createInfo *drive.File) (*drive.File, error) {
	var info *drive.File
	var err error

	err = f.pacer.Call(ctx, func() error {
		info, err = f.svc.Files.Create(createInfo).
			Media(in, googleapi.ContentType("")).
			Fields(googleapi.Field(partialFields)).
			SupportsAllDrives(f.isTeamDrive).
			KeepRevisionForever(f.opt.KeepRevisionForever).
			Do()
		return err
	})

	if err != nil {
		return nil, err
	}

	return info, nil
}

// uploadChunked uploads a file using a chunked upload protocol
func (f *Fs) uploadChunked(ctx context.Context, in io.Reader, size int64, createInfo *drive.File) (*drive.File, error) {
	// Call the detailed chunked upload implementation
	return f.uploadChunkedDetailed(ctx, in, size, createInfo)
}

// Mkdir creates a directory
func (f *Fs) Mkdir(ctx context.Context, dir string) error {
	_, err := f.dirCache.FindDir(ctx, dir)
	if err == nil {
		// Directory already exists
		return nil
	}
	if err != fs.ErrorDirNotFound {
		return err
	}

	// Create the directory
	_, err = f.dirCache.FindDir(ctx, dir)
	return err
}

// Rmdir removes a directory
func (f *Fs) Rmdir(ctx context.Context, dir string) error {
	directoryID, err := f.dirCache.FindDir(ctx, dir)
	if err != nil {
		return err
	}

	// Check if directory is empty
	list, err := f.List(ctx, dir)
	if err != nil {
		return err
	}
	if len(list) > 0 {
		return fs.ErrorDirectoryNotEmpty
	}
	// Delete the directory
	err = f.pacer.Call(ctx, func() error {
		return f.svc.Files.Delete(directoryID).
			SupportsAllDrives(f.isTeamDrive).
			Do()
	})

	if err != nil {
		return err
	}

	// Remove from directory cache
	f.dirCache.Put("", "")
	return nil
}

// Precision returns the precision of this Fs
func (f *Fs) Precision() time.Duration {
	return time.Millisecond
}

// Hashes returns the supported hash types of the filesystem
func (f *Fs) Hashes() hash.Set {
	return hash.NewHashSet([]hash.Type{hash.MD5, hash.SHA1})
}

// isGoogleDocument returns true if the file is a Google Document
func isGoogleDocument(file *drive.File) bool {
	return strings.HasPrefix(file.MimeType, "application/vnd.google-apps.")
}

// SetLogLevel changes the current log level
func (f *Fs) SetLogLevel(level LogLevel) {
	f.logger.SetLevel(level)
}

// GetLogger returns the fs's logger
func (f *Fs) GetLogger() *Logger {
	return f.logger
}

// Log helper methods
func (f *Fs) LogError(format string, args ...interface{}) {
	f.logger.Error(format, args...)
}

func (f *Fs) LogWarn(format string, args ...interface{}) {
	f.logger.Warn(format, args...)
}

func (f *Fs) LogInfo(format string, args ...interface{}) {
	f.logger.Info(format, args...)
}

func (f *Fs) LogDebug(format string, args ...interface{}) {
	f.logger.Debug(format, args...)
}

func (f *Fs) LogTrace(format string, args ...interface{}) {
	f.logger.Trace(format, args...)
}
