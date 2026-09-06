package main

// version はビルド時に -ldflags "-X main.version=vX.Y.Z" で上書きする。
var version = "dev"

func versionString() string {
	return version
}
