# Go parameters
PROJECT_NAME=indece-monitor-agent-linux
GOCMD=go
GOPATH=$(shell $(GOCMD) env GOPATH))
GOBUILD=$(GOCMD) build
GOGENERATE=$(GOCMD) generate
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
DIR_SOURCE=./src
DIR_DIST=./dist
DIR_DIST_DEBIAN=./dist/deb
DIR_DIST_RPM=./dist/rpm
DIR_GENERATED_MODEL=$(DIR_SOURCE)/generated/model
BINARY_NAME=$(DIR_DIST)/bin/$(PROJECT_NAME)
BUILD_DATE=$(shell date +%Y%m%d.%H%M%S)
BUILD_VERSION ?= $(shell git rev-parse --short HEAD)
LDFLAGS := 
LDFLAGS := $(LDFLAGS) -X github.com/indece-official/monitor-agent-linux/src/buildvars.ProjectName=$(PROJECT_NAME)
LDFLAGS := $(LDFLAGS) -X github.com/indece-official/monitor-agent-linux/src/buildvars.BuildDate=$(BUILD_DATE)
LDFLAGS := $(LDFLAGS) -X github.com/indece-official/monitor-agent-linux/src/buildvars.BuildVersion=$(BUILD_VERSION)

all: generate test build_amd64 build_amd64_full build_arm

generate:
	mkdir -p $(DIR_GENERATED_MODEL)
	rm -rf $(DIR_GENERATED_MODEL)/*
	$(GOGENERATE) -tags=bindata ./...

build_amd64:
	mkdir -p $(DIR_DIST)/bin
	GO111MODULE=on GOOS=linux GOARCH=amd64 $(GOBUILD) -ldflags "$(LDFLAGS)" -buildmode pie -o $(BINARY_NAME)_amd64 -tags=prod -v $(DIR_SOURCE)/main.go

build_amd64_full:
	mkdir -p $(DIR_DIST)/bin
	CGO_ENABLED=0 GO111MODULE=on GOOS=linux GOARCH=amd64 $(GOBUILD) -ldflags "$(LDFLAGS)" -buildmode pie -o $(BINARY_NAME)_amd64_full -tags=prod -v $(DIR_SOURCE)/main.go

build_arm:
	mkdir -p $(DIR_DIST)/bin
	GO111MODULE=on GOOS=linux GOARCH=arm GOARM=5 $(GOBUILD) -ldflags "$(LDFLAGS)" -o $(BINARY_NAME)_arm -tags=prod -v $(DIR_SOURCE)/main.go

test:
	mkdir -p $(DIR_DIST)
ifeq ($(OUTPUT),json)
	CGO_ENABLED=1 $(GOTEST) -v ./...  -cover -coverprofile $(DIR_DIST)/cover.out -json > $(DIR_DIST)/test.json
else
	CGO_ENABLED=1 $(GOTEST) -v ./...  -cover
endif

clean:
	#$(GOCLEAN)
	rm -rf $(DIR_OUT)

run:
	$(GOBUILD) -o $(BINARY_NAME) -v ./...
	./$(BINARY_NAME)

deps:
	echo test
	#$(GOGET) -d -v ./...

changelog:
	mkdir -p $(DIR_DIST)/changelog
	cp ./deploy/changelog $(DIR_DIST)/changelog/
	cp ./deploy/changelog.Debian $(DIR_DIST)/changelog/
	gzip -n -f -9 $(DIR_DIST)/changelog/changelog
	gzip -n -f -9 $(DIR_DIST)/changelog/changelog.Debian

package_rpm:
	mkdir -p $(DIR_DIST_RPM)/files
	BUILD_VERSION=$(BUILD_VERSION) WORK_DIR=$(shell realpath $(DIR_DIST_RPM)) envsubst < deploy/rpm/rpm.tpl.spec > $(DIR_DIST_RPM)/indece-monitor-agent-linux.spec
	objcopy --strip-unneeded $(BINARY_NAME)_amd64 $(DIR_DIST_RPM)/files/indece-monitor-agent-linux
	cp ./deploy/rpm/indece-monitor-agent-linux.service $(DIR_DIST_RPM)/files/
	cp ./deploy/rpm/agent-linux.conf $(DIR_DIST_RPM)/files/
	(cd $(DIR_DIST_RPM) && rpmbuild --target "x86_64" -bb ./indece-monitor-agent-linux.spec)
	rpm -qpivl --changelog --nomanifest ~/rpmbuild/RPMS/x86_64/indece-monitor-agent-linux-$(BUILD_VERSION)-1.x86_64.rpm
	rpm --addsign ~/rpmbuild/RPMS/x86_64/indece-monitor-agent-linux-$(BUILD_VERSION)-1.x86_64.rpm
	rpmlint ~/rpmbuild/RPMS/x86_64/indece-monitor-agent-linux-$(BUILD_VERSION)-1.x86_64.rpm

package_debian:
	mkdir -p $(DIR_DIST_DEBIAN)/indece-monitor-agent-linux_$(BUILD_VERSION)-1_amd64
	mkdir -p $(DIR_DIST_DEBIAN)/indece-monitor-agent-linux_$(BUILD_VERSION)-1_amd64/usr/bin/
	mkdir -p $(DIR_DIST_DEBIAN)/indece-monitor-agent-linux_$(BUILD_VERSION)-1_amd64/usr/lib/systemd/system/
	mkdir -p $(DIR_DIST_DEBIAN)/indece-monitor-agent-linux_$(BUILD_VERSION)-1_amd64/usr/share/doc/indece-monitor-agent-linux
	mkdir -p $(DIR_DIST_DEBIAN)/indece-monitor-agent-linux_$(BUILD_VERSION)-1_amd64/etc/indece-monitor/
	objcopy --strip-unneeded $(BINARY_NAME)_amd64 $(DIR_DIST_DEBIAN)/indece-monitor-agent-linux_$(BUILD_VERSION)-1_amd64/usr/bin/indece-monitor-agent-linux
	cp ./deploy/deb/indece-monitor-agent-linux.service $(DIR_DIST_DEBIAN)/indece-monitor-agent-linux_$(BUILD_VERSION)-1_amd64/usr/lib/systemd/system/
	cp ./deploy/deb/agent-linux.conf $(DIR_DIST_DEBIAN)/indece-monitor-agent-linux_$(BUILD_VERSION)-1_amd64/etc/indece-monitor/
	cp ./deploy/deb/copyright $(DIR_DIST_DEBIAN)/indece-monitor-agent-linux_$(BUILD_VERSION)-1_amd64//usr/share/doc/indece-monitor-agent-linux/
	cp $(DIR_DIST)/changelog/changelog.gz $(DIR_DIST_DEBIAN)/indece-monitor-agent-linux_$(BUILD_VERSION)-1_amd64//usr/share/doc/indece-monitor-agent-linux/
	cp $(DIR_DIST)/changelog/changelog.Debian.gz $(DIR_DIST_DEBIAN)/indece-monitor-agent-linux_$(BUILD_VERSION)-1_amd64//usr/share/doc/indece-monitor-agent-linux/
	mkdir -p $(DIR_DIST_DEBIAN)/indece-monitor-agent-linux_$(BUILD_VERSION)-1_amd64/DEBIAN/
	BUILD_VERSION=$(BUILD_VERSION) envsubst < ./deploy/deb/control.tpl > $(DIR_DIST_DEBIAN)/indece-monitor-agent-linux_$(BUILD_VERSION)-1_amd64/DEBIAN/control
	cp ./deploy/deb/conffiles $(DIR_DIST_DEBIAN)/indece-monitor-agent-linux_$(BUILD_VERSION)-1_amd64/DEBIAN/
	cp ./deploy/deb/postinst $(DIR_DIST_DEBIAN)/indece-monitor-agent-linux_$(BUILD_VERSION)-1_amd64/DEBIAN/
	chmod 0755 $(DIR_DIST_DEBIAN)/indece-monitor-agent-linux_$(BUILD_VERSION)-1_amd64/DEBIAN/postinst
	fakeroot dpkg --build $(DIR_DIST_DEBIAN)/indece-monitor-agent-linux_$(BUILD_VERSION)-1_amd64
	lintian $(DIR_DIST_DEBIAN)/indece-monitor-agent-linux_$(BUILD_VERSION)-1_amd64.deb
