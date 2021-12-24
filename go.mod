module github.com/chaosnote/go-build

go 1.12

replace github.com/chaosnote/go-kernel => ../go-kernel

require (
	github.com/chaosnote/go-kernel v0.0.0-20211016090353-ff4963675818
	github.com/go-redis/redis v6.15.9+incompatible
	github.com/go-sql-driver/mysql v1.5.0
	github.com/gorilla/websocket v1.4.2
	github.com/onsi/ginkgo v1.15.2 // indirect
	github.com/onsi/gomega v1.11.0 // indirect
	github.com/shopspring/decimal v1.3.1
	go.uber.org/zap v1.15.0
)
