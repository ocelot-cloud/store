package store

import (
	"time"
)

type Version struct {
	Name              string    `json:"name"`
	Id                string    `json:"id"`
	CreationTimestamp time.Time `json:"creation_timestamp"`
}

type NumberString struct {
	Value string `json:"value" validate:"number"`
}

type UserNameString struct {
	Value string `json:"value" validate:"number"`
}

type ChangePasswordForm struct {
	OldPassword string `json:"old_password" validate:"password"`
	NewPassword string `json:"new_password" validate:"password"`
}

type FullVersionInfo struct {
	Id                       int       `json:"id"`
	VersionName              string    `json:"version_name"`
	Maintainer               string    `json:"maintainer"`
	AppName                  string    `json:"app_name"`
	Content                  []byte    `json:"content"`
	VersionCreationTimestamp time.Time `json:"version_creation_timestamp"`
}

type RegistrationForm struct {
	User     string `json:"user" validate:"user_name"`
	Password string `json:"password" validate:"password"`
	Email    string `json:"email" validate:"email"`
}

type LoginCredentials struct {
	User     string `json:"user" validate:"user_name"`
	Password string `json:"password" validate:"password"`
}

type AppNameString struct {
	Value string `json:"value" validate:"app_name"`
}

type AppWithLatestVersion struct {
	Maintainer        string `json:"maintainer"`
	AppId             string `json:"app_id"`
	AppName           string `json:"app_name"`
	LatestVersionId   string `json:"latest_version_id"`
	LatestVersionName string `json:"latest_version_name"`
}

type AppSearchRequest struct {
	SearchTerm         string `json:"search_term" validate:"search_term"`
	ShowUnofficialApps bool   `json:"show_unofficial_apps"`
}

type App struct {
	Maintainer string `json:"user"`
	Name       string `json:"name"`
	Id         string `json:"id"`
}

type VersionUpload struct {
	AppId   string `json:"appId" validate:"number"`
	Version string `json:"version" validate:"version_name"`
	Content []byte `json:"content"`
}
