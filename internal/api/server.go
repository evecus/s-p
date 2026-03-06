name: Build & Release

on:
  push:
    tags: ['v*.*.*']
  workflow_dispatch:
    inputs:
      version:
        description: 'Version tag (e.g. v1.0.0)'
        required: false
        default: 'dev'

permissions:
  contents: write

jobs:
  build-web:
    name: Build Vue3 Frontend
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20'

      - name: Install dependencies
        working-directory: web
        run: npm install

      - name: Build frontend
        working-directory: web
        run: npm run build

      - name: Upload dist artifact
        uses: actions/upload-artifact@v4
        with:
          name: web-dist
          path: cmd/panel/dist/
          retention-days: 1

  build-go:
    name: Build Go Binary (${{ matrix.target }})
    needs: build-web
    runs-on: ubuntu-latest
    strategy:
      matrix:
        include:
          - goos: linux
            goarch: amd64
            target: linux-amd64
          - goos: linux
            goarch: arm64
            target: linux-arm64

    steps:
      - uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'
          cache: true

      - name: Download web dist
        uses: actions/download-artifact@v4
        with:
          name: web-dist
          path: cmd/panel/dist/

      - name: Download Go dependencies
        run: go mod download

      - name: Build binary
        env:
          GOOS: ${{ matrix.goos }}
          GOARCH: ${{ matrix.goarch }}
          GOARM: ${{ matrix.goarm }}
          CGO_ENABLED: 0
        run: |
          VERSION=${{ github.ref_name || inputs.version || 'dev' }}
          BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
          go build \
            -ldflags="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}" \
            -trimpath \
            -o singbox-panel-${{ matrix.target }} \
            ./cmd/panel

      - name: Compress binary
        run: |
          tar czf singbox-panel-${{ matrix.target }}.tar.gz singbox-panel-${{ matrix.target }}

      - name: Upload binary artifact
        uses: actions/upload-artifact@v4
        with:
          name: singbox-panel-${{ matrix.target }}
          path: |
            singbox-panel-${{ matrix.target }}
            singbox-panel-${{ matrix.target }}.tar.gz

  release:
    name: Create GitHub Release
    needs: build-go
    runs-on: ubuntu-latest
    if: startsWith(github.ref, 'refs/tags/')
    steps:
      - uses: actions/checkout@v4

      - name: Download all artifacts
        uses: actions/download-artifact@v4
        with:
          path: artifacts/
          pattern: singbox-panel-*
          merge-multiple: true

      - name: Generate checksums
        run: |
          cd artifacts
          sha256sum singbox-panel-*.tar.gz > SHA256SUMS.txt
          cat SHA256SUMS.txt

      - name: Create Release
        uses: softprops/action-gh-release@v2
        with:
          name: "Singbox Panel ${{ github.ref_name }}"
          body: |
            ## Singbox Panel ${{ github.ref_name }}

            ### 下载

            | 平台 | 架构 | 下载 |
            |------|------|------|
            | Linux | amd64 (x86_64) | `singbox-panel-linux-amd64.tar.gz` |
            | Linux | arm64 (aarch64) | `singbox-panel-linux-arm64.tar.gz` |
            | Linux | armv7 | `singbox-panel-linux-armv7.tar.gz` |

            ### 快速安装

            ```bash
            # 下载（以 amd64 为例）
            wget https://github.com/${{ github.repository }}/releases/download/${{ github.ref_name }}/singbox-panel-linux-amd64.tar.gz
            tar xzf singbox-panel-linux-amd64.tar.gz
            chmod +x singbox-panel-linux-amd64
            sudo ./singbox-panel-linux-amd64
            # 访问 http://<IP>:8080
            ```

            ### 变更日志
            ${{ github.event.head_commit.message }}
          files: |
            artifacts/singbox-panel-*.tar.gz
            artifacts/SHA256SUMS.txt
          draft: false
          prerelease: false
