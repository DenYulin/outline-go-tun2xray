BUILDDIR=$(CURDIR)/build
GOBIN=$(CURDIR)/bin

GOMOBILE=$(GOBIN)/gomobile
IMPORT_PATH=github.com/DenYulin/outline-go-tun2xray

.PHONY: android apple linux windows intra clean clean-all

all: intra android linux apple windows

# Don't strip Android debug symbols so we can upload them to crash reporting tools.
# Add GOBIN to $PATH so `gomobile` can find `gobind`.
ANDROID_BUILD_CMD=env PATH=$(GOBIN):$(PATH) $(GOMOBILE) bind -a -ldflags '-w' -target=android -tags android -work

intra: $(BUILDDIR)/intra/tun2xray.aar

$(BUILDDIR)/intra/tun2xray.aar: $(GOMOBILE)
	mkdir -p $(BUILDDIR)/intra
	$(ANDROID_BUILD_CMD) -o $@ $(IMPORT_PATH)/intra $(IMPORT_PATH)/intra/android $(IMPORT_PATH)/intra/doh $(IMPORT_PATH)/intra/split $(IMPORT_PATH)/intra/protect


android: $(BUILDDIR)/android/tun2xray.aar

$(BUILDDIR)/android/tun2xray.aar: $(GOMOBILE)
	mkdir -p $(BUILDDIR)/android
	$(ANDROID_BUILD_CMD) -o $@ $(IMPORT_PATH)/outline/android $(IMPORT_PATH)/outline/xray


apple: $(BUILDDIR)/apple/tun2xray.xcframework

$(BUILDDIR)/apple/tun2xray.xcframework: $(GOMOBILE)
  # MACOSX_DEPLOYMENT_TARGET and -iosversion should match what outline-client supports.
  # TODO(fortuna): -s strips symbols and is obsolete. Why are we using it?
	export MACOSX_DEPLOYMENT_TARGET=10.14; $(GOMOBILE) bind -iosversion=9.0 -target=ios,iossimulator,macos -o $@ -ldflags '-s -w' -bundleid org.outline.tun2xray $(IMPORT_PATH)/outline/apple $(IMPORT_PATH)/outline/xray


XGO=$(GOBIN)/xgo
TUN2XRAY_VERSION=v0.1.0
XGO_LDFLAGS='-s -w -X main.version=$(TUN2XRAY_VERSION)'
ELECTRON_PATH=$(IMPORT_PATH)/outline/electron


LINUX_BUILDDIR=$(BUILDDIR)/linux

linux: $(LINUX_BUILDDIR)/tun2xray

$(LINUX_BUILDDIR)/tun2xray: $(XGO)
	$(XGO) -ldflags $(XGO_LDFLAGS) --targets=linux/amd64 -dest $(LINUX_BUILDDIR) $(ELECTRON_PATH)
	mv $(LINUX_BUILDDIR)/electron-linux-amd64 $@


WINDOWS_BUILDDIR=$(BUILDDIR)/windows

windows: $(WINDOWS_BUILDDIR)/tun2xray.exe

$(WINDOWS_BUILDDIR)/tun2xray.exe: $(XGO)
	$(XGO) -ldflags $(XGO_LDFLAGS) --targets=windows/386 -dest $(WINDOWS_BUILDDIR) $(ELECTRON_PATH)
	mv $(WINDOWS_BUILDDIR)/electron-windows-4.0-386.exe $@


$(GOMOBILE): go.mod
	env GOBIN=$(GOBIN) go install golang.org/x/mobile/cmd/gomobile
	env GOBIN=$(GOBIN) $(GOMOBILE) init

$(XGO): go.mod
	env GOBIN=$(GOBIN) go install github.com/crazy-max/xgo

go.mod: tools.go
	go mod tidy
	touch go.mod

clean:
	rm -rf $(BUILDDIR)
	go clean

clean-all: clean
	rm -rf $(GOBIN)
