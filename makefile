# 定义要构建的目标名称
BINARY_NAME=idebug

# 定义要构建的目标平台和架构
PLATFORMS=darwin/amd64 darwin/arm64 linux/amd64 linux/386 linux/arm linux/arm64 windows/amd64 windows/386

# 定义 Go 编译器和标志
GO=go
BUILD_FLAGS=-ldflags="-s -w " -tags "!test"

# 定义 UPX 压缩命令
UPX=upx
UPX_FLAGS=-9

# 构建所有平台的目标
all: clean build compress

# 清理构建输出
clean:
	rm -rf ./bin

# 构建目标
build:
	mkdir -p ./bin
	$(foreach platform, $(PLATFORMS), \
		$(eval GOOS=$(word 1, $(subst /, ,$(platform)))) \
		$(eval GOARCH=$(word 2, $(subst /, ,$(platform)))) \
		$(eval BINARY_EXT=$(if $(filter $(GOOS),windows),.exe,)) \
		env GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build $(BUILD_FLAGS) -o ./bin/$(BINARY_NAME)-$(GOOS)-$(GOARCH)$(BINARY_EXT) .; \
	)


# 压缩二进制文件
compress:
	$(foreach file, $(wildcard ./bin/*), \
		$(UPX) $(UPX_FLAGS) $(file); \
	)

.PHONY: all clean build compress
