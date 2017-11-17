# Makefile for embedding build info into the executable

BUILD_DATE = $(shell date -u '+%Y-%m-%d_%H:%M:%S')
BUILD_DISTRO = $(shell lsb_release -sd)

all: win64 build

win32: export GOOS=windows
win32: export GOARCH=386
win32: winbuild

win64: export GOOS=windows
win64: export GOARCH=amd64
win64: winbuild

raspberry: export GOOS=linux
raspberry: export GOARCH=arm
raspberry: export GOARM=6
raspberry: build

raspberry2: export GOOS=linux
raspberry2: export GOARCH=arm
raspberry2: export GOARM=7
raspberry2: build

build:
	go build -o amlistener -ldflags "-X 'main.ApplicationBuildDate=$(BUILD_DATE)' -X 'main.ApplicationBuildDistro=$(BUILD_DISTRO)'"

winbuild:
	go build -o amlistener.exe -ldflags "-X 'main.ApplicationBuildDate=$(BUILD_DATE)' -X 'main.ApplicationBuildDistro=$(BUILD_DISTRO)'"
