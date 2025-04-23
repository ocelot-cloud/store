package tools

import "time"

type VersionUpload struct {
	AppId   string `json:"appId"`
	Version string `json:"version"`
	Content []byte `json:"content"`
}

type Version struct {
	Name              string    `json:"name"`
	Id                string    `json:"id"`
	CreationTimestamp time.Time `json:"creation_timestamp"`
}

type AppWithLatestVersion struct {
	Maintainer        string `json:"maintainer"`
	AppId             string `json:"app_id"`
	AppName           string `json:"app_name"`
	LatestVersionId   string `json:"latest_version_id"`
	LatestVersionName string `json:"latest_version_name"`
}

type App struct {
	Maintainer string `json:"user"`
	Name       string `json:"name"`
	Id         string `json:"id"`
}

type RegistrationForm struct {
	User     string `json:"user"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type LoginCredentials struct {
	User     string `json:"user"`
	Password string `json:"password"`
}

type FullVersionInfo struct {
	Id                       int       `json:"id"`
	VersionName              string    `json:"version_name"`
	Maintainer               string    `json:"maintainer"`
	AppName                  string    `json:"app_name"`
	Content                  []byte    `json:"content"`
	VersionCreationTimestamp time.Time `json:"version_creation_timestamp"`
}

type AppSearchRequest struct {
	SearchTerm         string `json:"search_term"`
	ShowUnofficialApps bool   `json:"show_unofficial_apps"`
}
