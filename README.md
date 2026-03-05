# Singbox Panel

一个为 sing-box 设计的 Web 管理面板，单二进制部署，支持 Linux 透明代理全套配置。

## 功能

- 🚀 **核心管理**：一键下载/升级 sing-box，进程管理
- 🌐 **代理模式**：TProxy / Redirect / TUN，支持本机代理和透明网关
- 🔢 **IPv4 + IPv6**：完整的双栈代理支持，自动配置 nftables 规则和策略路由
- ⚙️ **配置编辑**：DNS / 入站 / 出站 / 路由规则 / 规则集 图形化配置
- 📦 **Providers**：订阅管理，支持健康检查、自动更新
- 📋 **实时日志**：滚动日志查看，关键字过滤
- 🔒 **JWT 鉴权**：首次访问设置密码，Token 持久化

## 快速开始

### 下载安装

```bash
# amd64
wget https://github.com/YOUR_USER/singbox-panel/releases/latest/download/singbox-panel-linux-amd64.tar.gz
tar xzf singbox-panel-linux-amd64.tar.gz
chmod +x singbox-panel-linux-amd64
sudo ./singbox-panel-linux-amd64
```

访问 `http://<服务器IP>:8080`，首次访问设置密码。

### 参数

```
./singbox-panel --port 8080 --host 0.0.0.0 --data /etc/singbox-panel
```

### 设置为系统服务

```bash
sudo cp singbox-panel-linux-amd64 /usr/local/bin/singbox-panel

cat > /etc/systemd/system/singbox-panel.service << EOF
[Unit]
Description=Singbox Panel
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/singbox-panel --port 8080 --data /etc/singbox-panel
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable --now singbox-panel
```

## 代理模式说明

### TProxy（推荐）
- 支持 TCP + UDP
- 需要 nftables，Linux 内核 ≥ 4.18
- 自动配置防火墙规则 + ip rule + ip route

### Redirect
- 仅支持 TCP
- 使用 NAT 重定向
- 兼容性最好

### TUN
- sing-box 创建虚拟网卡
- 防火墙配置最简单
- 需要 `/dev/tun`

## 开发构建

```bash
# 安装依赖
cd web && npm install

# 开发模式（需本地运行后端）
npm run dev

# 构建前端
npm run build

# 构建后端
go build -o singbox-panel ./cmd/panel
```

## GitHub Actions

推送 `v*.*.*` 标签自动触发构建并发布 Release：

```bash
git tag v1.0.0
git push origin v1.0.0
```

构建产物：
- `singbox-panel-linux-amd64.tar.gz`
- `singbox-panel-linux-arm64.tar.gz`
- `singbox-panel-linux-armv7.tar.gz`

## 数据目录

默认 `/etc/singbox-panel/`：

```
/etc/singbox-panel/
├── config.json        # sing-box 配置
├── auth.json          # 面板认证
├── proxy_mode.json    # 代理模式配置
├── nftables.conf      # 生成的防火墙规则
├── sing-box           # sing-box 二进制
├── cache.db           # sing-box 缓存
└── providers/         # 订阅缓存
```
