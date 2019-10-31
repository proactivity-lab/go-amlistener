# Makefile for embedding build info into the executable

BUILD_DATE = $(shell date -u '+%Y-%m-%d_%H:%M:%S')
BUILD_DISTRO = $(shell lsb_release -sd)

all:
	@echo "Build with make linux32/linux64/win32/win64/raspberry/raspberry2!"

win32: export GOOS=windows
win32: export GOARCH=386
win32: export FLAVOUR=win32
win32: winbuild

win64: export GOOS=windows
win64: export GOARCH=amd64
win64: export FLAVOUR=win64
win64: winbuild

linux32: export GOOS=linux
linux32: export GOARCH=386
linux32: export FLAVOUR=$(GOOS)-x86
linux32: build

linux64: export GOOS=linux
linux64: export GOARCH=amd64
linux64: export FLAVOUR=$(GOOS)-amd64
linux64: build

raspberry: export GOOS=linux
raspberry: export GOARCH=arm
raspberry: export GOARM=6
raspberry: export FLAVOUR=$(GOOS)-$(GOARCH)$(GOARM)
raspberry: build

raspberry2: export GOOS=linux
raspberry2: export GOARCH=arm
raspberry2: export GOARM=7
raspberry2: export FLAVOUR=$(GOOS)-$(GOARCH)$(GOARM)
raspberry2: build

builddir: $(FLAVOUR)
	mkdir -p build/$(FLAVOUR)

build: builddir
	go build -o build/$(FLAVOUR)/amlistener -ldflags "-X 'main.ApplicationBuildDate=$(BUILD_DATE)' -X 'main.ApplicationBuildDistro=$(BUILD_DISTRO)'"
	tar -C build/$(FLAVOUR) -cjvf build/amlistener-$(FLAVOUR).tar.bz2 amlistener

winbuild:
	go build -o build/$(FLAVOUR)/amlistener.exe -ldflags "-X 'main.ApplicationBuildDate=$(BUILD_DATE)' -X 'main.ApplicationBuildDistro=$(BUILD_DISTRO)'"
	zip build/amlistener-$(FLAVOUR).zip -j build/$(FLAVOUR)/amlistener.exe

clean:
	rm -Rf build
