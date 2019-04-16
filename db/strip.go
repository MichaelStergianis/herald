package db

import "path/filepath"

// stripToArtist ...
func stripToArtist(fsPath string, lib Library) (artistPath string) {
	for filepath.Dir(fsPath) != lib.Path {
		fsPath = filepath.Dir(fsPath)
	}
	return fsPath
}

// stripToAlbum ...
func stripToAlbum(fsPath string, artist Artist) string {
	for filepath.Dir(fsPath) != artist.Path {
		fsPath = filepath.Dir(fsPath)
	}
	return fsPath
}
